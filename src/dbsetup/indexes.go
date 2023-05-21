package dbsetup

import (
	"eltneg/goliltemp/src/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DBindexes = db.DBIndexes{
	"users": {
		{
			Keys:    bson.M{"username": 1},
			Options: options.Index().SetUnique(true),
		},
	},
}
