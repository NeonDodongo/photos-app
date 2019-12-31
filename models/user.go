package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `bson:"_id"`
	Username  string             `bson:"username"`
	Password  string             `bson:"password"`
	Email     string             `bson:"email"`
	Posts     []Post             `bson:"posts"`
	Followers []string           `bson:"followers"` //UserIDs
	Following []string           `bson:"following"` //UserIDs
}

func InitUser() User {
	u := User{}
	u.ID = primitive.NewObjectID()
	return u
}
