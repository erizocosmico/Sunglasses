package models

import (
	"labix.org/v2/mgo/bson"
	"github.com/mvader/mask/services/interfaces"
	"time"
)

// Block model (same as follow)
type Block Follow

// BlockUser blocks an user ("from" blocks "to")
func BlockUser(from, to bson.ObjectId, conn interfaces.Saver) error {
	f := Block{}
f.ID = bson.NewObjectId()
f.To = to
f.From = from
f.Time = float64(time.Now().Unix())

if err := conn.Save("blocks", f.ID, f); err != nil {
// Remove user blocked from follows
return err
}

return nil
}

// UnblockUser unblocks an user ("from" unblocks "to")
func UnblockUser(from, to bson.ObjectId, conn interfaces.Conn) error {
	if err := conn.C("blocks").Remove(bson.M{"user_from": from, "user_to": to}); err != nil {
return err
}

return nil
}

// UserIsBlocked returns if the user is blocked
func UserIsBlocked(from, to bson.ObjectId, conn interfaces.Conn) bool {
	var (
		err   error
		count int
	)

	if count, err = conn.C("blocks").Find(bson.M{"user_from": from, "user_to": to}).Count(); err != nil {
return false
}

return count > 0
}
