package tests

import (
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/mvader/sunglasses/handlers"
	. "github.com/mvader/sunglasses/models"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGetUserTimeline(t *testing.T) {
	var (
		users        = make([]*User, 3)
		tokens       = make([]*Token, 3)
		posts        = make([][]string, 3)
		resultCounts = []float64{16.0, 18.0, 11.0}
	)

	conn := getConnection()

	for i := 0; i < 3; i++ {
		users[i] = NewUser()
		users[i].Username = fmt.Sprintf("fancy_user_%d", i)

		if err := users[i].Save(conn); err != nil {
			panic(err)
		}

		token := new(Token)
		token.Type = UserToken
		token.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
		token.UserID = users[i].ID
		if err := token.Save(conn); err != nil {
			panic(err)
		}
		tokens[i] = token
	}

	follow := func(a, b int) {
		FollowUser(users[a].ID, users[b].ID, conn)
	}
	follow(0, 1)
	follow(0, 2)
	follow(1, 0)
	follow(1, 2)
	follow(2, 0)

	for i := 0; i < 3; i++ {
		posts[i] = make([]string, 8)

		for j := 1; j <= 8; j++ {
			var pUser string

			if i == 0 {
				switch j {
				case 5:
					pUser = users[2].ID.Hex()
					break
				case 6:
					pUser = users[2].ID.Hex()
					break
				case 7:
					pUser = users[2].ID.Hex()
					break
				case 8:
					pUser = users[1].ID.Hex()
					break
				}
			} else if i == 1 {
				switch j {
				case 5:
					pUser = users[0].ID.Hex()
					break
				case 6:
					pUser = users[0].ID.Hex()
					break
				case 7:
					pUser = users[0].ID.Hex()
					break
				case 8:
					pUser = users[2].ID.Hex()
					break
				}
			} else {
				switch j {
				case 5:
					pUser = users[1].ID.Hex()
					break
				case 6:
					pUser = users[1].ID.Hex()
					break
				case 7:
					pUser = users[0].ID.Hex()
					break
				case 8:
					pUser = users[1].ID.Hex()
					break
				}
			}

			var post string

			testHandler(func(m *martini.ClassicMartini) {
				m.Post("/", handlers.CreatePost)
			}, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", tokens[i].Hash)
				r.PostForm.Add("post_text", fmt.Sprintf("A test status number %d for user %d", j, i))
				r.PostForm.Add("privacy_type", fmt.Sprintf("%d", j))

				if pUser != "" {
					r.PostForm.Add("privacy_users", pUser)
				}
			}, conn, "/", "POST", func(res *httptest.ResponseRecorder) {
				var errResp map[string]interface{}
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				post = errResp["post"].(map[string]interface{})["id"].(string)
			}, true)

			posts[i][j-1] = post
		}
	}

	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.Session.Close()
	}()

	// Take a nap while timelines are being propagated
	time.Sleep(4 * time.Second)

	for i := 0; i < 3; i++ {
		// For async's sake!
		j := i
		Convey(fmt.Sprintf("Getting timeline of user %d", i+1), t, func() {
			testGetHandler(handlers.GetUserTimeline, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokens[j].Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var result map[string]interface{}
				if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(result["count"].(float64), ShouldEqual, resultCounts[j])
				// We should test that the posts are actually the ones the user should see
				// for now you'll have to trust me, I checked it and they're right
				// TODO: do that
			})
		})
	}
}
