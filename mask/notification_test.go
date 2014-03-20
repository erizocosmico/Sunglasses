package mask

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSendNotification(t *testing.T) {
	var blankID bson.ObjectId
	conn := getConnection()
	user, token := createRequestUser(conn)

	Convey("Sending notifications", t, func() {
		Convey("should not give an error", func() {
			err := SendNotification(NotificationFollowed, user, blankID, blankID, conn)
			So(err, ShouldEqual, nil)
		})
	})

	conn.Db.C("notifications").RemoveAll(nil)
	user.Remove(conn)
	token.Remove(conn)
}

func TestMarkNotificationAsRead(t *testing.T) {
	var blankID, userNotificationID, userTmpNotificationID bson.ObjectId
	var notification Notification
	conn := getConnection()
	user, token := createRequestUser(conn)
	userTmp := NewUser()
	userTmp.Username = "testing_tmp"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	if err := SendNotification(NotificationFollowed, user, blankID, blankID, conn); err != nil {
		panic(err)
	}

	if err := SendNotification(NotificationFollowed, userTmp, blankID, blankID, conn); err != nil {
		panic(err)
	}

	if err := conn.Db.C("notifications").Find(bson.M{"user_id": user.ID}).One(&notification); err != nil {
		panic(err)
	} else {
		userNotificationID = notification.ID
	}

	if err := conn.Db.C("notifications").Find(bson.M{"user_id": userTmp.ID}).One(&notification); err != nil {
		panic(err)
	} else {
		userTmpNotificationID = notification.ID
	}

	Convey("Marking notifications as read", t, func() {
		Convey("When no valid user is given", func() {
			testPutHandler(MarkNotificationRead, func(r *http.Request) {}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 400)
					So(errResp.Code, ShouldEqual, CodeInvalidData)
					So(errResp.Message, ShouldEqual, MsgInvalidData)
				})
		})

		Convey("When no valid notification_id is given", func() {
			testPutHandler(MarkNotificationRead, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 400)
					So(errResp.Code, ShouldEqual, CodeInvalidData)
					So(errResp.Message, ShouldEqual, MsgInvalidData)
				})
		})

		Convey("When a non existent notification_id is given", func() {
			testPutHandler(MarkNotificationRead, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("notification_id", bson.NewObjectId().Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 404)
					So(errResp.Code, ShouldEqual, CodeNotFound)
					So(errResp.Message, ShouldEqual, MsgNotFound)
				})
		})

		Convey("When the notification doesn't belong to the user", func() {
			testPutHandler(MarkNotificationRead, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("notification_id", userTmpNotificationID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 403)
					So(errResp.Code, ShouldEqual, CodeUnauthorized)
					So(errResp.Message, ShouldEqual, MsgUnauthorized)
				})
		})

		Convey("When everything is OK", func() {
			testPutHandler(MarkNotificationRead, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("notification_id", userNotificationID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}

					if err := conn.Db.C("notifications").FindId(userNotificationID).One(&notification); err != nil {
						panic(err)
					}

					So(notification.Read, ShouldEqual, true)
					So(resp.Code, ShouldEqual, 200)
				})
		})
	})

	conn.Db.C("notifications").RemoveAll(nil)
	user.Remove(conn)
	token.Remove(conn)
	userTmp.Remove(conn)
}
