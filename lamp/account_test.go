package lamp

import (
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCreateUser(t *testing.T) {
	Convey("Subject: Creating a new user", t, func() {
		conn := getConnection()

		Convey("When the recovery method is not valid it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "john_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testin")
				req.PostForm.Add("recovery_method", "45")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
			})
		})

		Convey("When passwords don't match it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "john_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testin")
				req.PostForm.Add("recovery_method", "0")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
			})
		})

		Convey("When username is not valid it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "john doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testing")
				req.PostForm.Add("recovery_method", "0")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
			})
		})

		Convey("When recovery method is set to email and the email is not valid it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "john_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testing")
				req.PostForm.Add("recovery_method", "1")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
			})
		})

		Convey("When recovery method is set to security question and the either the answer or the question are empty it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "john_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testing")
				req.PostForm.Add("recovery_method", "2")
				req.PostForm.Add("recovery_answer", "How are you?")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
			})
		})

		Convey("When all the data is correct it should not fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "liam_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testing")
				req.PostForm.Add("recovery_method", "0")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldEqual, 200)
			})
		})

		Convey("When the user already exists it should fail", func() {
			testPostHandler(CreateUser, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "liam_doe")
				req.PostForm.Add("password", "testing")
				req.PostForm.Add("password_repeat", "testing")
				req.PostForm.Add("recovery_method", "0")
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldNotEqual, 200)
				conn.Db.C("users").RemoveAll(bson.M{"username": "liam_doe"})
			})
		})
	})
}
