package tests

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	. "github.com/mvader/mask/handlers"
	. "github.com/mvader/mask/models"
	. "github.com/mvader/mask/error"
)

func TestSendNotification(t *testing.T) {
	var blankID bson.ObjectId
	conn := getConnection()
	defer conn.Session.Close()
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
	defer conn.Session.Close()
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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

func TestListNotifications(t *testing.T) {
	var blankID bson.ObjectId
	conn := getConnection()
	defer conn.Session.Close()
	users := make([]bson.ObjectId, 0, 24)
	user, token := createRequestUser(conn)

	for i := 0; i < 24; i++ {
		u := NewUser()
		u.Username = fmt.Sprintf("test_%d", i)
		if err := u.Save(conn); err != nil {
			panic(err)
		}
		u.Settings.Invisible = false
		u.Settings.DisplayAvatarBeforeApproval = true
		if err := u.Save(conn); err != nil {
			panic(err)
		}
		users = append(users, u.ID)
	}

	for _, uid := range users {
		if err := SendNotification(NotificationFollowed, user, blankID, uid, conn); err != nil {
			panic(err)
		}
	}

	Convey("Listing user's notifications", t, func() {
		Convey("When invalid user is provided", func() {
			testGetHandler(ListNotifications, func(r *http.Request) {}, conn, "/", "/",
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

		Convey("When no count params are passed", func() {
			testGetHandler(ListNotifications, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["notifications"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(ListNotifications, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "10")
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(10))
					So(len(errResp["notifications"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(ListNotifications, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "10")
				r.Form.Add("offset", "15")
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(9))
					So(len(errResp["notifications"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count params are passed", func() {
			testGetHandler(ListNotifications, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "2")
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["notifications"].([]interface{})), ShouldEqual, 24)
				})
		})
	})

	conn.Db.C("notifications").RemoveAll(nil)
	user.Remove(conn)
	token.Remove(conn)
	if _, err := conn.Db.C("users").RemoveAll(bson.M{"_id": bson.M{"$in": users}}); err != nil {
		panic(err)
	}
}
