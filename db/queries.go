package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"photos-app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *MongoBongo) Upsert(t interface{}) error {
	var filter bson.M
	var c string

	if t == nil {
		return errors.New("Cannot upsert a nil document")
	}

	switch t.(type) {
	case models.Post:
		filter = bson.M{"id": t.(models.Post).ID}
		c = db.PostCollection
	case models.User:
		filter = bson.M{"username": t.(models.User).Username}
		c = db.UserCollection
	default:
		return errors.New("Cannot upsert document, incompatible types")
	}

	update := bson.M{"$set": t}
	collection := db.SetCollection(c)
	r, err := collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("Failed to insert one to %s collection [ %w ]", collection.Name(), err)
	}
	if r.MatchedCount == 0 {
		log.Printf("inserted one [ %v ] to collection [ %v ]", t, collection.Name())
	} else if r.ModifiedCount == 1 {
		log.Printf("updated one [ %v ] to collection [ %v ]", t, collection.Name())
	} // TODO: if modified count > 1, attempt delete duplicate records

	return nil
}

// FindUserByUsername find one User by their username
func (db *MongoBongo) FindUserByUsername(un string) (u models.User, err error) {
	filter := bson.M{"username": un}
	col := db.SetCollection(db.UserCollection)
	r := col.FindOne(context.Background(), filter)
	if err = r.Decode(&u); err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (db *MongoBongo) FindUserByID(id string) (u models.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, err
	}
	filter := bson.M{"_id": oid}
	col := db.SetCollection(db.UserCollection)
	r := col.FindOne(context.Background(), filter)
	if err = r.Decode(&u); err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (db *MongoBongo) FindPostByID(id string) (p models.Post, err error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return p, err
	}
	filter := bson.M{"_id": oid}
	col := db.SetCollection(db.PostCollection)
	r := col.FindOne(context.Background(), filter)
	if err = r.Decode(&p); err != nil {
		return p, err
	}

	fmt.Printf("POST: %v\n", p)

	return p, nil
}
