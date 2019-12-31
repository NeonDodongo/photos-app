package db

import (
	"context"
	"log"
	"photos-app/cfg"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoBongo struct {
	Client         *mongo.Client
	Database       string
	UserCollection string
	PostCollection string
}

func (db *MongoBongo) SetCollection(c string) *mongo.Collection {
	switch c {
	case db.PostCollection:
		log.Printf("post collection name is being set [ %v ]", c)
		return db.Client.Database(db.Database).Collection(db.PostCollection)
	case db.UserCollection:
		log.Printf("users collection name is being set [ %v ]", c)
		return db.Client.Database(db.Database).Collection(db.UserCollection)
	default:
		log.Print("default collection name is being set to user collection")
		return db.Client.Database(db.Database).Collection(db.UserCollection)
	}
}

func InitDatabase(c cfg.Config) MongoBongo {

	client, err := mongo.NewClient(options.Client().ApplyURI(c.Mongo.URL))
	if err != nil {
		log.Fatalf("error init mongo client [%v]", err)
	}

	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("FAILED TO CONNECT TO MONGO DB [%v]", err)
	}

	db := &MongoBongo{
		Client:         client,
		Database:       c.Mongo.Database,
		UserCollection: c.Mongo.UserCollection,
		PostCollection: c.Mongo.PostCollection,
	}

	return *db
}
