package mask

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"sync"
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
				So(res.Code, ShouldEqual, 200)
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
				So(res.Code, ShouldEqual, 200)
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
				So(res.Code, ShouldEqual, 200)
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
			if path[strlen(path)-4:] == "jpeg" {
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
				r.PostForm.Add("post_text", randomString(3000))
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
				r.PostForm.Add("post_type", "link")
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
				So(res.Code, ShouldEqual, 200)
			})
		})
	})
}
