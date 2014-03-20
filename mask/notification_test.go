package mask

import (
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestSendNotification(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	Convey("Sending notifications", t, func() {
		Convey("should not give an error", func() {
			err := SendNotification(NotificationFollowed, user, bson.NewObjectId(), bson.NewObjectId(), conn)
			So(err, ShouldEqual, nil)
		})
	})

	conn.Db.C("notifications").RemoveAll(nil)
	user.Remove(conn)
	token.Remove(conn)
}
