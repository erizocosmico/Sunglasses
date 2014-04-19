package tests

import (
	"encoding/json"
	. "github.com/mvader/mask/handlers"
	. "github.com/mvader/mask/models"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShowUserProfile(t *testing.T) {
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

	for i := 0; i < 25; i++ {
		post := NewPost(PostStatus, user)
		post.Text = "A fancy post"

		if i < 12 {
			post.Privacy = PrivacySettings{Type: PrivacyPublic}
		} else {
			post.Privacy = PrivacySettings{}
		}

		if err := post.Save(conn); err != nil {
			panic(err)
		}
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("users").RemoveAll(nil)
		conn.Db.C("tokens").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Getting user profile", t, func() {
		Convey("When the user has access to the profile", func() {
			testGetHandler(ShowUserProfile, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/:username", "/testing", func(res *httptest.ResponseRecorder) {
				var errResp map[string]interface{}
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp["posts_count"].(float64), ShouldEqual, float64(25))
				So(errResp["user"].(map[string]interface{})["username"].(string), ShouldNotEqual, "Protected")
			})
		})

		Convey("When the profile does not exist", func() {
			testGetHandler(ShowUserProfile, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
			}, conn, "/:username", "/testing_not_existent_username", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 404)
			})
		})

		Convey("When the user does not have access to the profile", func() {
			testGetHandler(ShowUserProfile, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
			}, conn, "/:username", "/testing", func(res *httptest.ResponseRecorder) {
				var errResp map[string]interface{}
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp["posts_count"].(float64), ShouldEqual, float64(12))
				So(errResp["user"].(map[string]interface{})["username"].(string), ShouldEqual, "Protected")
			})
		})
	})
}
