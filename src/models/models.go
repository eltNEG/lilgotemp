package models

import (
	"context"
	"eltneg/goliltemp/src/db"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type Model interface {
	SetCreatedAt()
	SetUpdatedAt()
}

type Query[T Model] interface{}

var ErrDocumentLocked = errors.New("document is locked")

type DBModel[M Model, Q Query[M]] struct {
	Name              string
	Store             *db.Datastore
	CustomQueryParser func(q ...Q) (primitive.M, error)
}

func (d *DBModel[M, Q]) Drop(ctx context.Context) error {
	c := d.Store.Database.Collection(d.Name)
	return c.Drop(ctx)
}

func (d *DBModel[M, Q]) Create(ctx context.Context, model M) (string, error) {
	c := d.Store.Database.Collection(d.Name)
	model.SetCreatedAt()
	model.SetUpdatedAt()
	result, err := c.InsertOne(ctx, model)
	if err != nil {
		return "", err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return id, err
}

func (d *DBModel[M, Q]) Get(ctx context.Context, key string, value any) (M, error) {
	var doc M
	c := d.Store.Database.Collection(d.Name)
	query := bson.M{
		key: value,
	}

	err := c.FindOne(ctx, query).Decode(&doc)

	if err == mongo.ErrNoDocuments {
		return doc, nil
	}
	if err != nil {
		return doc, err
	}
	return doc, err
}

var ErrNotFound = errors.New("document not found")

func (d *DBModel[M, Q]) FindOne(ctx context.Context, query Q) (M, error) {

	docs, err := d.Find(ctx, query)
	if err != nil {
		return *new(M), err
	}
	if len(docs) == 0 {
		return *new(M), ErrNotFound
	}
	return docs[0], nil
}

func (d *DBModel[M, Q]) Find(ctx context.Context, query Q) ([]M, error) {
	if d.CustomQueryParser != nil {
		_query, err := d.CustomQueryParser(query)
		if err != nil {
			return nil, err
		}
		return d.Finder(ctx, _query)
	}
	_query, err := bsonify(query)
	if err != nil {
		return nil, err
	}
	return d.Finder(ctx, _query)
}

func (d *DBModel[M, Q]) FindMany(ctx context.Context, query []Q) ([]M, error) {
	if d.CustomQueryParser != nil {
		_query, err := d.CustomQueryParser(query...)
		if err != nil {
			return nil, err
		}
		return d.Finder(ctx, _query)
	}
	q := bson.M{
		"$or": query,
	}
	_query, err := bsonify(q)
	if err != nil {
		return nil, err
	}
	return d.Finder(ctx, _query)
}

func (d *DBModel[M, Q]) Finder(ctx context.Context, query primitive.M) ([]M, error) {
	c := d.Store.Database.Collection(d.Name)
	cursor, err := c.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	var docs []M
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc M
		err = cursor.Decode(&doc)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (d *DBModel[M, Q]) Delete(ctx context.Context, key string, value any) (err error) {
	c := d.Store.Database.Collection(d.Name)
	query := bson.M{
		key: value,
	}
	_, err = c.DeleteOne(ctx, query)
	return err
}

func (d *DBModel[M, Q]) FindOneAndDelete(ctx context.Context, query Q) (err error) {
	c := d.Store.Database.Collection(d.Name)
	_query, err := bsonify(query)
	if err != nil {
		return err
	}
	_, err = c.DeleteOne(ctx, _query)
	return err
}

func (d *DBModel[M, Q]) Update(ctx context.Context, query Q, data M) (err error) {
	c := d.Store.Database.Collection(d.Name)
	data.SetUpdatedAt()
	_query, err := bsonify(query)
	if err != nil {
		return err
	}
	_data, err := bsonify(data)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": _data,
	}
	err = c.FindOneAndUpdate(ctx, _query, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(data)
	return err
}

func (d *DBModel[M, Q]) Upsert(ctx context.Context, query Q, data M) error {
	c := d.Store.Database.Collection(d.Name)
	data.SetCreatedAt()
	data.SetUpdatedAt()
	filter, err := bsonify(query)
	if err != nil {
		return err
	}
	upsert := true
	opts := options.FindOneAndUpdateOptions{
		Upsert: &upsert,
	}
	b, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	var update bson.M
	err = bson.Unmarshal(b, &update)
	if err != nil {
		return err
	}
	_query := bson.D{{Key: "$set", Value: update}}
	err = c.FindOneAndUpdate(ctx, filter, _query, &opts).Decode(data)
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func (d *DBModel[M, Q]) CustomWithCollection(ctx context.Context, f func(ctx context.Context, c *mongo.Collection) error) (err error) {
	c := d.Store.Database.Collection(d.Name)
	return f(ctx, c)
}

func bsonify[T any](data T) (primitive.M, error) {
	var update bson.M
	b, err := bson.Marshal(data)
	if err != nil {
		return update, err
	}
	err = bson.Unmarshal(b, &update)
	return update, err
}

func (d *DBModel[M, Q]) Transaction(ctx context.Context, callback func(mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)
	session, err := d.Store.Database.Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, callback, txnOpts)
}

func (d *DBModel[M, Q]) Count(ctx context.Context, query Q) (int64, error) {
	c := d.Store.Database.Collection(d.Name)
	_query, err := bsonify(query)
	if err != nil {
		return 0, err
	}
	return c.CountDocuments(ctx, _query)

}
