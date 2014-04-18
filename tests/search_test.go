package tests

import (
	"encoding/json"
	"fmt"
	. "github.com/mvader/mask/handlers"
	"github.com/mvader/mask/models"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSearch(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)

	ids := make([]bson.ObjectId, 0, 24)
	for i := 0; i < 24; i++ {
		u := models.NewUser()
		u.Username = fmt.Sprintf("fancy_testing_user_%d", i)
		if err := u.Save(conn); err != nil {
			panic(err)
		}

		u.Settings.Invisible = false
		if err := u.Save(conn); err != nil {
			panic(err)
		}

		ids = append(ids, u.ID)

		if i < 12 {
			models.FollowUser(u.ID, user.ID, conn)
		} else {
			models.FollowUser(user.ID, u.ID, conn)
		}
	}

	defer func() {
		conn.C("users").RemoveAll(nil)
		conn.C("tokens").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Testing search", t, func() {
		Convey("Searching for all users", func() {
			Convey("When no count params passed", func() {
				testGetHandler(Search, func(r *http.Request) {
					r.Header.Add("X-User-Token", token.Hash)

						if r.Form == nil {
							r.Form = make(url.Values)
						}

					r.Form.Add("q", "test")
				}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(res.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(25))
				})
			})

			Convey("When count = 10 and offset = 20", func() {
				testGetHandler(Search, func(r *http.Request) {
					r.Header.Add("X-User-Token", token.Hash)
					if r.Form == nil {
						r.Form = make(url.Values)
					}

					r.Form.Add("count", "10")
					r.Form.Add("offset", "20")
					r.Form.Add("q", "test")
				}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
					var errResp map[string]interface{}
					if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
						panic(err)
					}
					So(res.Code, ShouldEqual, 200)
					So(errResp["count"].(float64), ShouldEqual, float64(5))
				})
			})
		})

		Convey("Searching for just the followings", func() {
				Convey("When no count params passed", func() {
						testGetHandler(Search, func(r *http.Request) {
								r.Header.Add("X-User-Token", token.Hash)

								if r.Form == nil {
									r.Form = make(url.Values)
								}

								r.Form.Add("q", "test")
								r.Form.Add("just_followings", "true")
							}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
								var errResp map[string]interface{}
								if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
									panic(err)
								}
								So(res.Code, ShouldEqual, 200)
								So(errResp["count"].(float64), ShouldEqual, float64(12))
							})
					})

				Convey("When count = 10 and offset = 10", func() {
						testGetHandler(Search, func(r *http.Request) {
								r.Header.Add("X-User-Token", token.Hash)
								if r.Form == nil {
									r.Form = make(url.Values)
								}

								r.Form.Add("count", "10")
								r.Form.Add("offset", "20")
								r.Form.Add("q", "test")
								r.Form.Add("just_followings", "true")
							}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
								var errResp map[string]interface{}
								if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
									panic(err)
								}
								So(res.Code, ShouldEqual, 200)
								So(errResp["count"].(float64), ShouldEqual, float64(5))
							})
					})
		})
	})
}
