package models

import "labix.org/v2/mgo/bson"

type TimelineEntry struct {
	ID       bson.ObjectId   `bson:"_id"`
	User     bson.ObjectId   `bson:"user_id"`
	Post     bson.ObjectId   `bson:"post_id"`
	PostUser bson.ObjectId   `bson:"post_user_id"`
	Liked    bool            `bson:"liked"`
	Comments []bson.ObjectId `bson:"comments"`
	Time     float64         `bson:"time"`
}
