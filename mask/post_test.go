package mask

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPostStatus(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
	}()

	Convey("Posting a status", t, func() {
		Convey("When no user is passed", func() {
			testPostHandler(CreatePost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When the status text is invalid", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", randomString(3000))
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidStatusText)
				So(errResp.Message, ShouldEqual, MsgInvalidStatusText)
			})
		})

		Convey("When everything is OK", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", "A test status")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Status posted successfully")
			})
		})
	})
}
