package mask

import (
	"encoding/json"
	"fmt"
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

func TestListBlocks(t *testing.T) {
	conn := getConnection()
	users := make([]bson.ObjectId, 0, 24)
	user, token := createRequestUser(conn)

	for i := 0; i < 24; i++ {
		u := NewUser()
		u.Username = fmt.Sprintf("test_%i", i)
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
		if err := BlockUser(user.ID, uid, conn); err != nil {
			panic(err)
		}
	}

	Convey("Listing blocked users", t, func() {
		Convey("When invalid user is provided", func() {
			testGetHandler(ListBlocks, func(r *http.Request) {}, conn, "/", "/",
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
			testGetHandler(ListBlocks, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["blocks"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(ListBlocks, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
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
					So(len(errResp["blocks"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(ListBlocks, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
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
					So(len(errResp["blocks"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count param are passed", func() {
			testGetHandler(ListBlocks, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.ID.Hex())
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
					So(len(errResp["blocks"].([]interface{})), ShouldEqual, 24)
				})
		})
	})

	for _, u := range users {
		if err := UnblockUser(user.ID, u, conn); err != nil {
			panic(err)
		}
	}

	user.Remove(conn)
	token.Remove(conn)
	if _, err := conn.Db.C("users").RemoveAll(bson.M{"_id": bson.M{"$in": users}}); err != nil {
		panic(err)
	}
}
