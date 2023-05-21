package db

import (
	"context"
	"eltneg/goliltemp/src/config"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client is the Mongodb Client
// Datastore defines the store structure
type Datastore struct {
	Database *mongo.Database
}

type DBIndexes map[string][]mongo.IndexModel

//Store holds reference to the mongodb client and database

// Init : Initialize Mongodb instance
func Init(cfg *config.Config, indexes DBIndexes, opts *options.ClientOptions) (context.Context, Datastore, error) {
	store := Datastore{}
	uri := cfg.MongoDBURI
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if opts == nil {
		opts = options.Client().ApplyURI(uri)
	}

	Client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return ctx, store, err
	}
	// Ping the connection
	if err := Client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return ctx, store, err
	}
	store.Database = Client.Database(cfg.DBName)
	log.Infof("Successfully connected to database [ %v ] and pinged.", cfg.DBName)

	err = createUniqueIndex(ctx, store, indexes)
	if err != nil {
		return ctx, store, err
	}

	return ctx, store, nil
}

func createUniqueIndex(ctx context.Context, store Datastore, indexes DBIndexes) error {
	for colName, fieldIndexes := range indexes {
		collection := store.Database.Collection(colName)
		_, err := collection.Indexes().CreateMany(ctx, fieldIndexes)
		if err != nil {
			return err
		}
	}
	return nil
}
