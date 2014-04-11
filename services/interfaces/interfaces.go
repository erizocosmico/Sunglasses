package interfaces

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Saver interface {
	Save(collection string, ID bson.ObjectId, item interface{}) error
}

type Remover interface {
	Remove(collection string, ID bson.ObjectId) error
}

type Conn interface {
	C(collection string) *mgo.Collection
	Saver
	Remover
}
