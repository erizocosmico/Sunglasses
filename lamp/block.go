package lamp

import (
	"labix.org/v2/mgo/bson"
	"time"
)

// Block model (same as follow)
type Block Follow

// BlockUser blocks an user ("from" blocks "to")
func BlockUser(from, to bson.ObjectId, conn *Connection) error {
	f := Block{}
	f.ID = bson.NewObjectId()
	f.To = to
	f.From = from
	f.Time = float64(time.Now().Unix())

	if err := conn.Save("blocks", f.ID, f); err != nil {
		return err
	}

	return nil
}

// UnblockUser unblocks an user ("from" unblocks "to")
func UnblockUser(from, to bson.ObjectId, conn *Connection) error {
	if err := conn.Db.C("blocks").Remove(bson.M{"from": from, "to": to}); err != nil {
		return err
	}

	return nil
}

// UserIsBlocked returns if the user is blocked
func UserIsBlocked(from, to bson.ObjectId, conn *Connection) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.Db.C("blocks").Find(bson.M{"from": from, "to": to}).Count(); err != nil {
		return false
	}

	return count > 0
}
