package tests

import (
	"encoding/json"
	"github.com/go-martini/martini"
	. "github.com/mvader/sunglasses/handlers"
	"github.com/mvader/sunglasses/middleware"
	. "github.com/mvader/sunglasses/models"
	"github.com/mvader/sunglasses/modules/timeline"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPropagatePostsOnCreation(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when a new post is created", t, func() {
		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			So(errResp.Message, ShouldEqual, "Status posted successfully")
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnUserFollow(t *testing.T) {
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

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when an user is followed", t, func() {
		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			So(errResp.Message, ShouldEqual, "Status posted successfully")
		}, true)

		time.Sleep(500 * time.Millisecond)

		FollowUser(userTmp.ID, user.ID, conn)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 0)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", func(c middleware.Context) {
				timeline.PropagatePostsOnUserFollow(c, user.ID)
			})
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", tokenTmp.Hash)
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {}, true)

		time.Sleep(500 * time.Millisecond)

		count, err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnPrivacyChange(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)
	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when the privacy settings of a post change", t, func() {
		var post string

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			post = errResp["post"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Put("/:id", ChangePostPrivacy)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("privacy_type", "4")
		}, conn, "/"+post, "PUT", func(res *httptest.ResponseRecorder) {
			So(res.Code, ShouldEqual, 200)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 0)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnUserUnfollow(t *testing.T) {
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

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when an user is unfollowed", t, func() {
		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			So(errResp.Message, ShouldEqual, "Status posted successfully")
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", func(c middleware.Context) {
				timeline.PropagatePostsOnUserUnfollow(c, user.ID)
			})
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", tokenTmp.Hash)
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {}, true)

		time.Sleep(500 * time.Millisecond)

		count, err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 0)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnLike(t *testing.T) {
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

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when a post is liked", t, func() {
		var (
			t    TimelineEntry
			post string
		)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			post = errResp["post"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(t.Liked, ShouldEqual, false)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Put("/:id", LikePost)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", tokenTmp.Hash)
		}, conn, "/"+post, "PUT", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 200)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(t.Liked, ShouldEqual, true)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnDeletion(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when a post is deleted", t, func() {
		var post string

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			post = errResp["post"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Delete("/:id", DeletePost)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
		}, conn, "/"+post, "DELETE", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 200)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 0)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnUserDeletion(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when an user is deleted", t, func() {
		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			So(res.Code, ShouldEqual, 201)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 1)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Delete("/", DestroyAccount)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.PostForm.Add("confirmed", "true")
		}, conn, "/", "DELETE", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 200)
		}, true)

		time.Sleep(500 * time.Millisecond)

		count, err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).Count()
		So(count, ShouldEqual, 0)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnNewComment(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("comments").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when a new comment is added to a post", t, func() {
		var (
			t    TimelineEntry
			post string
		)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			post = errResp["post"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(len(t.Comments), ShouldEqual, 0)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreateComment)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.PostForm.Add("post_id", post)
			r.PostForm.Add("comment_text", "Hey! I'm a comment!")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			So(res.Code, ShouldEqual, 201)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(len(t.Comments), ShouldEqual, 1)
		So(err, ShouldEqual, nil)
	})
}

func TestPropagatePostsOnCommentDeleted(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	userTmp := NewUser()
	userTmp.Username = "testing_very_hard"
	if err := userTmp.Save(conn); err != nil {
		panic(err)
	}

	FollowUser(userTmp.ID, user.ID, conn)

	conn.C("timelines").RemoveAll(nil)

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("comments").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Propagating posts to timelines when a comment is deleted", t, func() {
		var (
			t         TimelineEntry
			cmt, post string
		)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreatePost)
		}, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
			r.PostForm.Add("post_text", "A test status")
			r.PostForm.Add("privacy_type", "1")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			post = errResp["post"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err := conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(len(t.Comments), ShouldEqual, 0)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Post("/", CreateComment)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.PostForm.Add("post_id", post)
			r.PostForm.Add("comment_text", "Hey! I'm a comment!")
		}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
			var errResp map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 201)
			cmt = errResp["comment"].(map[string]interface{})["id"].(string)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(len(t.Comments), ShouldEqual, 1)
		So(err, ShouldEqual, nil)

		testHandler(func(m *martini.ClassicMartini) {
			m.Delete("/:comment_id", RemoveComment)
		}, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.PostForm.Add("confirmed", "true")
		}, conn, "/"+cmt, "DELETE", func(res *httptest.ResponseRecorder) {
			So(res.Code, ShouldEqual, 200)
		}, true)

		time.Sleep(500 * time.Millisecond)

		err = conn.C("timelines").Find(bson.M{"user_id": userTmp.ID}).One(&t)
		So(len(t.Comments), ShouldEqual, 0)
		So(err, ShouldEqual, nil)
	})
}
