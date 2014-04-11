package tests

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
	. "github.com/mvader/mask/handlers"
	. "github.com/mvader/mask/models"
	. "github.com/mvader/mask/error"
	"github.com/mvader/mask/util"
)

func TestPostStatus(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
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
				r.PostForm.Add("post_text", util.RandomString(3000))
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
				So(res.Code, ShouldEqual, 201)
				So(errResp.Message, ShouldEqual, "Status posted successfully")
			})
		})
	})
}

/*
Wercker does not like this test :-(

func TestPostCrazyness(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	var wg sync.WaitGroup

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
	}()

	for j := 1; j < 100; j++ {
		wg.Add(100)
		for i := 0; i < 100; i++ {

			go func() {
				cn := new(Connection)
				cn.Session = conn.Session.Copy()
				cn.Db = cn.Session.DB("mask_test")

				testPostHandler(CreatePost, func(r *http.Request) {
						if r.PostForm == nil {
							r.PostForm = make(url.Values)
						}
						r.Header.Add("X-User-Token", token.Hash)
						r.PostForm.Add("post_text", "A test status")
					}, cn, "/", "/", func(res *httptest.ResponseRecorder) {
						cn.Session.Close()
						wg.Done()
					})
			}()
		}
		wg.Wait()
	}
}
*/

func TestPostVideo(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Posting a video", t, func() {
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
				r.PostForm.Add("post_type", "video")
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", util.RandomString(3000))
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

		Convey("When the link is not valid (youtube)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://youtube.com/watch?v=notfoundvideo")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidVideoURL)
				So(errResp.Message, ShouldEqual, MsgInvalidVideoURL)
			})
		})

		Convey("When the link is not valid (vimeo)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://vimeo.com/00000")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidVideoURL)
				So(errResp.Message, ShouldEqual, MsgInvalidVideoURL)
			})
		})

		Convey("When everything is OK (vimeo)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://vimeo.com/89856635")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 201)
				So(errResp.Message, ShouldEqual, "Video posted successfully")
			})
		})

		Convey("When everything is OK (youtube)", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "video")
				r.PostForm.Add("video_url", "http://www.youtube.com/watch?v=9bZkp7q19f0")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 201)
				So(errResp.Message, ShouldEqual, "Video posted successfully")
			})
		})
	})
}

func TestPostLink(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Posting a link", t, func() {
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
				r.PostForm.Add("post_type", "link")
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", util.RandomString(3000))
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

		Convey("When the link is not valid", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "link")
				r.PostForm.Add("link_url", "http://alargedomainnamethatdoesnotexist.com")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidLinkURL)
				So(errResp.Message, ShouldEqual, MsgInvalidLinkURL)
			})
		})

		Convey("When everything is OK", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "link")
				r.PostForm.Add("link_url", "http://google.es")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 201)
				So(errResp.Message, ShouldEqual, "Link posted successfully")
			})
		})
	})
}

func TestPostPhoto(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
		filepath.Walk("../test_assets/", func(path string, _ os.FileInfo, _ error) error {
			if path[util.Strlen(path)-4:] == "jpeg" {
				os.Remove(path)
			}
			return nil
		})
	}()

	Convey("Posting a photo", t, func() {
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
			testUploadFileHandler("../test_assets/gopher.jpg", "post_picture", "/", CreatePost, conn, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "photo")
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", util.RandomString(3000))
			}, func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
			})
		})

		Convey("When no file is uploaded", func() {
			testPostHandler(CreatePost, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "photo")
				r.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
			})
		})

		Convey("When everything is OK", func() {
			testUploadFileHandler("../test_assets/gopher.jpg", "post_picture", "/", CreatePost, conn, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_type", "photo")
				r.Header.Add("X-User-Token", token.Hash)
				r.PostForm.Add("post_text", "Fancy pic")
			}, func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 201)
			})
		})
	})
}

func TestDeletePost(t *testing.T) {
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

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		userTmp.Remove(conn)
		tokenTmp.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Deleting a post", t, func() {
		Convey("When no user is passed", func() {
			testDeleteHandler(DeletePost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When an invalid post id is passed", func() {
			testDeleteHandler(DeletePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", "")
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
			testDeleteHandler(DeletePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", bson.NewObjectId().Hex())
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

		Convey("When the post does not belong to the user", func() {
			testDeleteHandler(DeletePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
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

		Convey("When everything is OK", func() {
			testDeleteHandler(DeletePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestLikePost(t *testing.T) {
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

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("likes").RemoveAll(nil)
		user.Remove(conn)
		token.Remove(conn)
		userTmp.Remove(conn)
		tokenTmp.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Liking a post", t, func() {
		Convey("When no user is passed", func() {
			testPostHandler(LikePost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When an invalid post id is passed", func() {
			testPostHandler(LikePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", "")
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
			testPostHandler(LikePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", bson.NewObjectId().Hex())
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
			testPostHandler(LikePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", tokenTmp.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
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

		Convey("When everything is OK (like)", func() {
			testPostHandler(LikePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Post liked successfully")
			})
		})

		Convey("When everything is OK (unlike)", func() {
			testPostHandler(LikePost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "Post unliked successfully")
			})
		})
	})
}

func TestShowPost(t *testing.T) {
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

	uids := make([]bson.ObjectId, 0, 10)
	for i := 0; i < 10; i++ {
		u := NewUser()
		u.Username = "testing_user_" + fmt.Sprint(i)
		u.PrivateName = "Super secret"
		if err := u.Save(conn); err != nil {
			panic(err)
		}

		uids = append(uids, u.ID)
	}

	for i := 0; i < 6; i++ {
		FollowUser(user.ID, uids[i], conn)
	}

	post := NewPost(PostStatus, user)
	post.Text = "A fancy post"
	post.Privacy = PrivacySettings{}
	if err := post.Save(conn); err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		c := NewComment(uids[i], post.ID)
		c.Message = "Fancy comment"
		if err := c.Save(conn); err != nil {
			panic(err)
		}
	}

	defer func() {
		conn.Db.C("posts").RemoveAll(nil)
		conn.Db.C("comments").RemoveAll(nil)
		conn.Db.C("users").RemoveAll(nil)
		conn.Db.C("tokens").RemoveAll(nil)
		conn.Session.Close()
	}()

	Convey("Showing a post", t, func() {
		Convey("When no user is passed", func() {
			testGetHandler(ShowPost, nil, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidData)
				So(errResp.Message, ShouldEqual, MsgInvalidData)
			})
		})

		Convey("When an invalid post id is passed", func() {
			testGetHandler(ShowPost, func(r *http.Request) {
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
			testGetHandler(ShowPost, func(r *http.Request) {
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
			testGetHandler(ShowPost, func(r *http.Request) {
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

		Convey("When everything is OK", func() {
			testGetHandler(ShowPost, func(r *http.Request) {
				r.Header.Add("X-User-Token", token.Hash)
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.PostForm.Add("post_id", post.ID.Hex())
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp map[string]interface{}
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(len(errResp["post"].(map[string]interface{})["comments"].([]interface{})), ShouldEqual, 10)

				count := 0
				for _, v := range errResp["post"].(map[string]interface{})["comments"].([]interface{}) {
					if v.(map[string]interface{})["user"].(map[string]interface{})["private_name"] != "" {
						count++
					}
				}

				So(count, ShouldEqual, 6)
			})
		})
	})
}
