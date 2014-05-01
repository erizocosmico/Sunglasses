package tests

import (
	"encoding/json"
	. "github.com/mvader/sunglasses/error"
	. "github.com/mvader/sunglasses/handlers"
	. "github.com/mvader/sunglasses/models"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestCreateComment(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	tokenTmp := new(Token)
	tokenTmp.Type = UserToken
	tokenTmp.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
	tokenTmp.UserID = userTmp.ID
	if err := tokenTmp.Save(conn); err != nil {
		panic(err)
	}

	post := NewPost(PostStatus, user)
	post.Text = "A fancy post"
	post.Privacy = PrivacySettings{Type: PrivacyFollowingOnly}
	if err := post.Save(conn); err != nil {
		panic(err)
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("comments").RemoveAll(nil)
		conn.Db.C("users").RemoveAll(nil)
		conn.Db.C("tokens").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Creating new comments", t, func() {
		Convey("When an invalid post id is passed", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", "")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When a post id that doesn't exist is passed", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", bson.NewObjectId().Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 404)
				So(errResp.Code, ShouldEqual, CodeNotFound)
				So(errResp.Message, ShouldEqual, MsgNotFound)
			})
		})

		Convey("When the post can't be accessed by the user", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeUnauthorized)
				So(errResp.Message, ShouldEqual, MsgUnauthorized)
			})
		})

		Convey("When the comment text is not valid", func() {
			FollowUser(user.ID, userTmp.ID, conn)

			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidCommentText)
				So(errResp.Message, ShouldEqual, MsgInvalidCommentText)
			})
		})

		Convey("When everything is OK", func() {
			testPostHandler(CreateComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
				r.PostForm.Add("comment_text", "My fancy comment")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 201)
			})
		})
	})
}

func TestRemoveComment(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	tokenTmp := new(Token)
	tokenTmp.Type = UserToken
	tokenTmp.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
	tokenTmp.UserID = userTmp.ID
	if err := tokenTmp.Save(conn); err != nil {
		panic(err)
	}

	post := NewPost(PostStatus, user)
	post.Text = "A fancy post"
	post.Privacy = PrivacySettings{}
	if err := post.Save(conn); err != nil {
		panic(err)
	}

	c := NewComment(user.ID, bson.NewObjectId())
	c.Message = "A fancy comment"
	if err := c.Save(conn); err != nil {
		panic(err)
	}

	c2 := NewComment(user.ID, post.ID)
	c2.Message = "A fancy comment"
	if err := c2.Save(conn); err != nil {
		panic(err)
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("comments").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		userTmp.Remove(conn)
		tokenTmp.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Removing comments", t, func() {
		Convey("When an invalid comment id is passed", func() {
			testDeleteHandler(RemoveComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
			}, conn, "/:comment_id", "/a", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When a comment id that doesn't exist is passed", func() {
			testDeleteHandler(RemoveComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:comment_id", "/"+bson.NewObjectId().Hex(), func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 404)
				So(errResp.Code, ShouldEqual, CodeNotFound)
				So(errResp.Message, ShouldEqual, MsgNotFound)
			})
		})

		Convey("When the comment does not belong to the user", func() {
			testDeleteHandler(RemoveComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
			}, conn, "/:comment_id", "/"+c.ID.Hex(), func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeUnauthorized)
				So(errResp.Message, ShouldEqual, MsgUnauthorized)
			})
		})

		Convey("When the post does not exist", func() {
			testDeleteHandler(RemoveComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:comment_id", "/"+c.ID.Hex(), func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 404)
				So(errResp.Code, ShouldEqual, CodeNotFound)
				So(errResp.Message, ShouldEqual, MsgNotFound)
			})
		})

		Convey("When everything is OK", func() {
			testDeleteHandler(RemoveComment, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:comment_id", "/"+c2.ID.Hex(), func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestCommentsForPost(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	tokenTmp := new(Token)
	tokenTmp.Type = UserToken
	tokenTmp.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
	tokenTmp.UserID = userTmp.ID
	if err := tokenTmp.Save(conn); err != nil {
		panic(err)
	}

	post := NewPost(PostStatus, user)
	post.Text = "A fancy post"
	post.Privacy = PrivacySettings{}
	if err := post.Save(conn); err != nil {
		panic(err)
	}

	for i := 0; i < 24; i++ {
		c := NewComment(user.ID, post.ID)
		c.Message = "A fancy comment"
		if err := c.Save(conn); err != nil {
			panic(err)
		}
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("comments").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		userTmp.Remove(conn)
		tokenTmp.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Listing comments for post", t, func() {
		Convey("When post does not exist", func() {
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:post_id", "/"+bson.NewObjectId().Hex(),
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

		Convey("When post can't be acessed by the user", func() {
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
			}, conn, "/:post_id", "/"+post.ID.Hex(),
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
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:post_id", "/"+post.ID.Hex(),
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["comments"].([]interface{})), ShouldEqual, 24)
				})
		})

		Convey("When count param is passed", func() {
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "10")
			}, conn, "/:post_id", "/"+post.ID.Hex(),
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(10))
					So(len(errResp["comments"].([]interface{})), ShouldEqual, 10)
				})
		})

		Convey("When count param and offset are passed", func() {
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "10")
				r.Form.Add("offset", "15")
			}, conn, "/:post_id", "/"+post.ID.Hex(),
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(9))
					So(len(errResp["comments"].([]interface{})), ShouldEqual, 9)
				})
		})

		Convey("When invalid count params are passed", func() {
			testGetHandler(CommentsForPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.Form == nil {
					r.Form = make(url.Values)
				}
				r.Form.Add("count", "2")
			}, conn, "/:post_id", "/"+post.ID.Hex(),
				func(resp *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(resp.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(24))
					So(len(errResp["comments"].([]interface{})), ShouldEqual, 24)
				})
		})
	})
}
