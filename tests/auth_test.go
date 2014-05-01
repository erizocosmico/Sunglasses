package tests

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	. "github.com/mvader/mask/handlers"
	. "github.com/mvader/mask/middleware"
	. "github.com/mvader/mask/models"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGetAccessToken(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Subject: Getting access token", t, func() {

		Convey("When we request an access token the response status will be 200 and we must receive a token", func() {
			testGetHandler(GetAccessToken, func(_ *http.Request) {}, conn, "/", "/",
				func(response *httptest.ResponseRecorder) {
					var resultBody map[string]interface{}
					So(response.Code, ShouldEqual, 200)
					err := json.Unmarshal(response.Body.Bytes(), &resultBody)
					if err != nil {
						panic(err)
					}
					So(resultBody["access_token"].(string), ShouldNotEqual, "")
				})
		})
	})
}

func TestGetUserToken(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	user := new(User)
	user.Username = "Jane Doe"
	err := user.SetPassword("testing")
	if err != nil {
		panic(err)
	}
	user.Role = RoleUser
	user.Active = true
	if err := user.Save(conn); err != nil {
		panic(err)
	}

	Convey("Subject: Getting user token", t, func() {
		Convey("When we request an user token with valid data the response code will be 200", func() {
			testPostHandler(GetUserToken, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jane Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 200)
			})
		})

		Convey("When we request an user token with invalid data the response code will be 400", func() {
			testPostHandler(GetUserToken, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jana Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 400)
				if err = user.Remove(conn); err != nil {
					panic(err)
				}
			})
		})
	})
}

func TestLogin(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	user := new(User)
	user.Username = "Jane Doe"
	err := user.SetPassword("testing")
	if err != nil {
		panic(err)
	}
	user.Role = RoleUser
	user.Active = true
	if err := user.Save(conn); err != nil {
		panic(err)
	}

	Convey("Subject: Logging the user in", t, func() {

		Convey("When we provide valid data the user will be logged in", func() {
			var sess *sessions.Session
			testPostHandler(func(c Context) {
				sess = c.Session
				Login(c)
			}, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jane Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 200)
				So(sess.Values["user_token"], ShouldNotEqual, nil)
			})
		})

		Convey("When we request an user token with invalid data the response code will be 400", func() {
			var sess *sessions.Session
			testPostHandler(func(c Context) {
				sess = c.Session
				Login(c)
			}, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jana Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 400)
				So(sess.Values["user_token"], ShouldEqual, nil)
				if err := user.Remove(conn); err != nil {
					panic(err)
				}
			})
		})
	})
}

func TestDestroyUserToken(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Subject: Testing user token destruction", t, func() {
		token := new(Token)
		token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
		token.Type = UserToken

		if err := token.Save(conn); err != nil {
			panic(err)
		}

		Convey("When a valid user token is given the response status will be 200", func() {
			testDeleteHandler(DestroyUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 200)
			})
		})
	})
}
