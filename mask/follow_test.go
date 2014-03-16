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

func TestFollowUser(t *testing.T) {
	conn := getConnection()
	Convey("Following users", t, func() {
		from := bson.NewObjectId()
		to := bson.NewObjectId()

		err := FollowUser(from, to, conn)
		Convey("We should get no errors", func() {
			So(err, ShouldEqual, nil)
			count, err := conn.Db.C("follows").Find(bson.M{"user_from": from, "user_to": to}).Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 1)
		})
	})
}

func TestUnfollowUser(t *testing.T) {
	conn := getConnection()
	Convey("Unfollowing users", t, func() {
		from := bson.NewObjectId()
		to := bson.NewObjectId()

		if err := FollowUser(from, to, conn); err != nil {
			panic(err)
		}

		err := UnfollowUser(from, to, conn)
		Convey("We should get no errors", func() {
			So(err, ShouldEqual, nil)
			count, err := conn.Db.C("follows").Find(bson.M{"user_from": from, "user_to": to}).Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})
	})
}

func TestSendFollowRequest(t *testing.T) {
	conn := getConnection()

	Convey("Sending follow requests", t, func() {

		Convey("With invalid request user", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(SendFollowRequest, func(r *http.Request) {}, conn, "/", "/",
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

		Convey("With invalid 'user_to'", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(SendFollowRequest, func(r *http.Request) {
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

		Convey("When 'user_to' does not exist", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("user_to", bson.NewObjectId().Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 404)
					So(errResp.Code, ShouldEqual, CodeUserDoesNotExist)
					So(errResp.Message, ShouldEqual, MsgUserDoesNotExist)
				})
		})

		Convey("When 'user_from' follows 'user_to'", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			if err := FollowUser(user.ID, userTo.ID, conn); err != nil {
				panic(err)
			}

			defer func() {
				if err := UnfollowUser(user.ID, userTo.ID, conn); err != nil {
					panic(err)
				}
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("user_to", userTo.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp.Error, ShouldEqual, false)
					So(errResp.Message, ShouldEqual, "You already follow that user")
				})
		})

		Convey("When 'user_to' follows 'user_from'", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			if err := FollowUser(userTo.ID, user.ID, conn); err != nil {
				panic(err)
			}

			defer func() {
				if err := UnfollowUser(userTo.ID, user.ID, conn); err != nil {
					panic(err)
				}
				if err := UnfollowUser(user.ID, userTo.ID, conn); err != nil {
					panic(err)
				}
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("user_to", userTo.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp.Error, ShouldEqual, false)
					So(errResp.Message, ShouldEqual, "User followed successfully")
				})
		})

		Convey("When user can't receive requests", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("user_to", userTo.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 403)
					So(errResp.Code, ShouldEqual, CodeUserCantBeRequested)
					So(errResp.Message, ShouldEqual, MsgUserCantBeRequested)
				})
		})

		Convey("When everything is OK", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			userTo.Settings.CanReceiveRequests = true
			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
				if err := conn.Db.C("requests").DropCollection(); err != nil {
					panic(err)
				}
			}()

			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("user_to", userTo.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp.Error, ShouldEqual, false)
					So(errResp.Message, ShouldEqual, "Follow request sent successfully")
				})
		})
	})
}
