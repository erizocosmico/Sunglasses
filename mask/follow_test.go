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
	"time"
)

func TestFollowUser(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

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
	defer conn.Session.Close()

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
	defer conn.Session.Close()

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

		Convey("When 'user_to' does not exist", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(SendFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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
				r.Header.Add("X-User-Token", token.Hash)
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

func TestReplyFollowRequest(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Replying follow requests", t, func() {
		Convey("With invalid request user", func() {
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {}, conn, "/", "/",
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

		Convey("With invalid request id", func() {
			user, token := createRequestUser(conn)

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {
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

		Convey("With a valid request id which does not exist", func() {
			user, token := createRequestUser(conn)

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("request_id", bson.NewObjectId().Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 404)
					So(errResp.Code, ShouldEqual, CodeFollowRequestDoesNotExist)
					So(errResp.Message, ShouldEqual, MsgFollowRequestDoesNotExist)
				})
		})

		Convey("With a valid request id that does not belong to the user", func() {
			user, token := createRequestUser(conn)

			req := new(FollowRequest)
			req.From = bson.NewObjectId()
			req.To = bson.NewObjectId()
			if err := req.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
				req.Remove(conn)
			}()
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("request_id", req.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 404)
					So(errResp.Code, ShouldEqual, CodeUnauthorized)
					So(errResp.Message, ShouldEqual, MsgUnauthorized)
				})
		})

		Convey("With a valid request id and 'accept' != 'yes'", func() {
			user, token := createRequestUser(conn)

			req := new(FollowRequest)
			req.From = bson.NewObjectId()
			req.To = user.ID
			if err := req.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("request_id", req.ID.Hex())
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp.Error, ShouldEqual, false)
					So(errResp.Message, ShouldEqual, "Successfully replied to follow request")
				})
		})

		Convey("With a valid request id and 'accept' = 'yes'", func() {
			user, token := createRequestUser(conn)

			req := new(FollowRequest)
			req.From = bson.NewObjectId()
			req.To = user.ID
			if err := req.Save(conn); err != nil {
				panic(err)
			}

			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(ReplyFollowRequest, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("request_id", req.ID.Hex())
				r.PostForm.Add("accept", "yes")
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp errorResponse
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp.Error, ShouldEqual, false)
					So(errResp.Message, ShouldEqual, "Successfully replied to follow request")
				})
		})
	})
}

func TestUnfollow(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Unfollowing an user", t, func() {

		Convey("With invalid request user", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(Unfollow, func(r *http.Request) {}, conn, "/", "/",
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
			testPostHandler(Unfollow, func(r *http.Request) {
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

		Convey("When 'user_to' does not exist", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()
			testPostHandler(Unfollow, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
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

		Convey("When the user does not follow 'user_to'", func() {
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

			testPostHandler(Unfollow, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
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
					So(errResp.Message, ShouldEqual, "You can't unfollow that user")
				})
		})

		Convey("When the user follows 'user_to'", func() {
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
				user.Remove(conn)
				token.Remove(conn)
				userTo.Remove(conn)
			}()

			testPostHandler(Unfollow, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
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
					So(errResp.Message, ShouldEqual, "User unfollowed successfully")
				})
		})
	})
}

func TestListFollowers(t *testing.T) {
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
		if err := FollowUser(uid, user.ID, conn); err != nil {
			panic(err)
		}
	}

	Convey("Listing followers", t, func() {
		Convey("When invalid user is provided", func() {
			testGetHandler(ListFollowers, func(r *http.Request) {}, conn, "/", "/",
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
			testGetHandler(ListFollowers, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["followers"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(ListFollowers, func(r *http.Request) {
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
					So(len(errResp["followers"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(ListFollowers, func(r *http.Request) {
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
					So(len(errResp["followers"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count params are passed", func() {
			testGetHandler(ListFollowers, func(r *http.Request) {
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
					So(len(errResp["followers"].([]interface{})), ShouldEqual, 24)
				})
		})
	})

	for _, u := range users {
		if err := UnfollowUser(u, user.ID, conn); err != nil {
			panic(err)
		}
	}

	user.Remove(conn)
	token.Remove(conn)
	if _, err := conn.Db.C("users").RemoveAll(bson.M{"_id": bson.M{"$in": users}}); err != nil {
		panic(err)
	}
}

func TestListFollowing(t *testing.T) {
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
		if err := FollowUser(user.ID, uid, conn); err != nil {
			panic(err)
		}
	}

	Convey("Listing followings", t, func() {
		Convey("When invalid user is provided", func() {
			testGetHandler(ListFollowing, func(r *http.Request) {}, conn, "/", "/",
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
			testGetHandler(ListFollowing, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["followings"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(ListFollowing, func(r *http.Request) {
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
					So(len(errResp["followings"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(ListFollowing, func(r *http.Request) {
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
					So(len(errResp["followings"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count params are passed", func() {
			testGetHandler(ListFollowing, func(r *http.Request) {
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
					So(len(errResp["followings"].([]interface{})), ShouldEqual, 24)
				})
		})
	})

	for _, u := range users {
		if err := UnfollowUser(user.ID, u, conn); err != nil {
			panic(err)
		}
	}

	user.Remove(conn)
	token.Remove(conn)
	if _, err := conn.Db.C("users").RemoveAll(bson.M{"_id": bson.M{"$in": users}}); err != nil {
		panic(err)
	}
}

func TestListFollowRequests(t *testing.T) {
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
		fr := new(FollowRequest)
		fr.From = uid
		fr.To = user.ID
		fr.Msg = "Message"
		fr.Time = float64(time.Now().Unix())

		if err := fr.Save(conn); err != nil {
			panic(err)
		}
	}

	Convey("Listing follow requests", t, func() {
		Convey("When invalid user is provided", func() {
			testGetHandler(ListFollowRequests, func(r *http.Request) {}, conn, "/", "/",
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
			testGetHandler(ListFollowRequests, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/",
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["follow_requests"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(ListFollowRequests, func(r *http.Request) {
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
					So(len(errResp["follow_requests"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(ListFollowRequests, func(r *http.Request) {
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
					So(len(errResp["follow_requests"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count params are passed", func() {
			testGetHandler(ListFollowRequests, func(r *http.Request) {
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
					So(len(errResp["follow_requests"].([]interface{})), ShouldEqual, 24)
				})
		})
	})

	if _, err := conn.Db.C("requests").RemoveAll(bson.M{"user_to": user.ID}); err != nil {
		panic(err)
	}

	user.Remove(conn)
	token.Remove(conn)
	if _, err := conn.Db.C("users").RemoveAll(bson.M{"_id": bson.M{"$in": users}}); err != nil {
		panic(err)
	}
}
