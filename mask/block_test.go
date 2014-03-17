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

func TestBlockUser(t *testing.T) {
	conn := getConnection()
	Convey("Blocking users", t, func() {
		from := bson.NewObjectId()
		to := bson.NewObjectId()

		err := BlockUser(from, to, conn)
		Convey("We should get no errors", func() {
			So(err, ShouldEqual, nil)
			count, err := conn.Db.C("blocks").Find(bson.M{"user_from": from, "user_to": to}).Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 1)
		})
	})
}

func TestUnblockUser(t *testing.T) {
	conn := getConnection()
	Convey("Unblocking users", t, func() {
		from := bson.NewObjectId()
		to := bson.NewObjectId()

		if err := BlockUser(from, to, conn); err != nil {
			panic(err)
		}

		err := UnblockUser(from, to, conn)
		Convey("We should get no errors", func() {
			So(err, ShouldEqual, nil)
			count, err := conn.Db.C("blocks").Find(bson.M{"user_from": from, "user_to": to}).Count()
			So(err, ShouldEqual, nil)
			So(count, ShouldEqual, 0)
		})
	})
}

func TestBlockHandler(t *testing.T) {
	conn := getConnection()

	Convey("Blocking an user", t, func() {
		Convey("With invalid request user", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(BlockHandler, func(r *http.Request) {}, conn, "/", "/",
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
			testPostHandler(BlockHandler, func(r *http.Request) {
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
			testPostHandler(BlockHandler, func(r *http.Request) {
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

		Convey("If user_to is blocked", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			if err := BlockUser(user.ID, userTo.ID, conn); err != nil {
				panic(err)
			}

			defer func() {
				if err := UnblockUser(user.ID, userTo.ID, conn); err != nil {
					panic(err)
				}
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(BlockHandler, func(r *http.Request) {
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
					So(errResp.Message, ShouldEqual, "User was already blocked")
				})
		})

		Convey("When everything is OK", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				if err := UnblockUser(user.ID, userTo.ID, conn); err != nil {
					panic(err)
				}
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(BlockHandler, func(r *http.Request) {
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
					So(errResp.Message, ShouldEqual, "User blocked successfully")
				})
		})
	})
}

func TestUnblock(t *testing.T) {
	conn := getConnection()

	Convey("Unblocking an user", t, func() {

		Convey("With invalid request user", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(Unblock, func(r *http.Request) {}, conn, "/", "/",
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
			testPostHandler(Unblock, func(r *http.Request) {
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
			testPostHandler(Unblock, func(r *http.Request) {
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

		Convey("If user_to is not blocked", func() {
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

			testPostHandler(Unblock, func(r *http.Request) {
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
					So(errResp.Message, ShouldEqual, "User was not blocked")
				})
		})

		Convey("When everything is OK", func() {
			user, token := createRequestUser(conn)
			userTo := NewUser()
			userTo.Username = "testing_to"

			if err := userTo.Save(conn); err != nil {
				panic(err)
			}

			if err := BlockUser(user.ID, userTo.ID, conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(Unblock, func(r *http.Request) {
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
					So(errResp.Message, ShouldEqual, "User unblocked successfully")
				})
		})
	})
}
