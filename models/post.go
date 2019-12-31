package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Post struct {
	ID          primitive.ObjectID `bson:"_id"`
	PostID      string             `bson:"postID"`
	CreatedBy   string             `bson:"createdBy"` //Username
	ImgName     string             `bson:"imgName"`
	Caption     string             `bson:"caption"`
	Comments    []Comment          `bson:"comments"`
	CreateDt    time.Time          `bson:"createDt"`
	ImgLocation string             `bson:"imgLocation"`
}

type Comment struct {
	CreatedBy string    `bson:"createdBy"` //Username
	Content   string    `bson:"content"`
	CreateDt  time.Time `bson:"createDt"`
}

func InitPost() (p Post) {
	p.ID = primitive.NewObjectID()
	p.PostID = p.ID.Hex()
	var cSlice []Comment
	p.Comments = cSlice
	p.CreateDt = time.Now()
	return
}
