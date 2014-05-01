package tests

import (
	"encoding/json"
	. "github.com/mvader/sunglasses/error"
	. "github.com/mvader/sunglasses/handlers"
	. "github.com/mvader/sunglasses/models"
	"github.com/mvader/sunglasses/util"
	. "github.com/smartystreets/goconvey/convey"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	// TODO needs revisit
	conn := getConnection()
	defer conn.Session.Close()
	Convey("Subject: Creating a new user", t, func() {

		Convey("When the recovery method is not valid it should fail", func() {
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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
			testPostHandler(CreateAccount, func(req *http.Request) {
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

func TestGetAccountInfo(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Testing the retrieval of account info", t, func() {
		Convey("When valid user is passed", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()

			testGetHandler(GetAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestGetAccountSettings(t *testing.T) {
	conn := getConnection()
	defer conn.Session.Close()

	Convey("Testing the retrieval of account settings", t, func() {
		Convey("When valid user is passed", func() {
			user, token := createRequestUser(conn)
			defer func() {
				user.Remove(conn)
				token.Remove(conn)
			}()

			testGetHandler(GetAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				So(res.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestUpdateAccountInfo(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Testing the update of account info", t, func() {
		Convey("When any of the fields is more than 500 characters long", func() {
			testPutHandler(UpdateAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("work", util.RandomString(501))
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidInfoLength)
				So(errResp.Message, ShouldEqual, MsgInvalidInfoLength)
			})
		})

		Convey("When invalid URLs are given", func() {
			testPutHandler(UpdateAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("websites", "http://google.es")
				req.PostForm.Add("websites", "this is not an url")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidWebsites)
				So(errResp.Message, ShouldEqual, MsgInvalidWebsites)
			})
		})

		Convey("When invalid gender is given", func() {
			testPutHandler(UpdateAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("gender", "male")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidGender)
				So(errResp.Message, ShouldEqual, MsgInvalidGender)
			})
		})

		Convey("When invalid status is given", func() {
			testPutHandler(UpdateAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("status", "married")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidStatus)
				So(errResp.Message, ShouldEqual, MsgInvalidStatus)
			})
		})

		Convey("When everything is OK", func() {
			testPutHandler(UpdateAccountInfo, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("websites", "http://google.es")
				req.PostForm.Add("work", "20th Century Fox")
				req.PostForm.Add("education", "Harvard")
				req.PostForm.Add("gender", "0")
				req.PostForm.Add("status", "0")
				req.PostForm.Add("tv", "Bones, Game of Thrones")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "User info updated successfully")
			})
		})
	})
}

func TestUpdateAccountSettings(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
	}()

	Convey("Testing the update of account settings", t, func() {
		Convey("When override default privacy is true and there is an error with privacy settings", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "true")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidPrivacySettings)
				So(errResp.Message, ShouldEqual, MsgInvalidPrivacySettings)
			})
		})

		Convey("When override default privacy is true and everything is OK", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "true")
				req.PostForm.Add("privacy_status_type", "1")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "User settings updated successfully")
			})
		})

		Convey("When override default privacy is true and users are required but not provided", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "true")
				req.PostForm.Add("privacy_status_type", "5")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidPrivacySettings)
				So(errResp.Message, ShouldEqual, MsgInvalidPrivacySettings)
			})
		})

		Convey("When override default privacy is false and there is an error with privacy settings", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "false")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidPrivacySettings)
				So(errResp.Message, ShouldEqual, MsgInvalidPrivacySettings)
			})
		})

		Convey("When override default privacy is false and everything is OK", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "false")
				req.PostForm.Add("privacy_status_type", "1")
				req.PostForm.Add("privacy_video_type", "1")
				req.PostForm.Add("privacy_link_type", "1")
				req.PostForm.Add("privacy_photo_type", "1")
				req.PostForm.Add("privacy_album_type", "1")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 200)
				So(errResp.Message, ShouldEqual, "User settings updated successfully")
			})
		})

		Convey("When recovery question or recovery answer are empty", func() {
			testPutHandler(UpdateAccountSettings, func(req *http.Request) {
				req.Header.Add("X-User-Token", token.Hash)
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("override_default_privacy", "true")
				req.PostForm.Add("privacy_status_type", "1")
				req.PostForm.Add("recovery_method", "2")
			}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(res.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidRecoveryQuestion)
				So(errResp.Message, ShouldEqual, MsgInvalidRecoveryQuestion)
			})
		})
	})
}

func TestUpdateProfilePicture(t *testing.T) {
	conn := getConnection()
	user, token := createRequestUser(conn)
	defer func() {
		user.Remove(conn)
		token.Remove(conn)
		conn.Session.Close()
		filepath.Walk("../test_assets/", func(path string, _ os.FileInfo, _ error) error {
			if path[util.Strlen(path)-3:] == "png" && path[util.Strlen(path)-10:] != "avatar.png" {
				os.Remove(path)
			}
			return nil
		})
	}()

	Convey("Updating a profile picture", t, func() {
		Convey("When no file is uploaded", func() {
			testPostHandler(UpdateProfilePicture, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
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
			testUploadFileHandler("../test_assets/gopher_avatar.png", "account_picture", "/", UpdateProfilePicture, conn, func(r *http.Request) {
				if r.PostForm == nil {
					r.PostForm = make(url.Values)
				}
				r.Header.Add("X-User-Token", token.Hash)
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

func TestDestroyAccount(t *testing.T) {
	conn := getConnection()

	// Make sure there are no residual records
	conn.C("tokens").RemoveAll(nil)
	conn.C("users").RemoveAll(nil)
	conn.C("posts").RemoveAll(nil)
	conn.C("notifications").RemoveAll(nil)
	conn.C("follows").RemoveAll(nil)
	conn.C("blocks").RemoveAll(nil)
	conn.C("likes").RemoveAll(nil)
	conn.C("timelines").RemoveAll(nil)
	conn.C("comments").RemoveAll(nil)

	user, token := createRequestUser(conn)
	defer func() {
		conn.C("tokens").RemoveAll(nil)
		conn.C("users").RemoveAll(nil)
		conn.C("posts").RemoveAll(nil)
		conn.C("notifications").RemoveAll(nil)
		conn.C("follows").RemoveAll(nil)
		conn.C("blocks").RemoveAll(nil)
		conn.C("likes").RemoveAll(nil)
		conn.C("timelines").RemoveAll(nil)
		conn.C("comments").RemoveAll(nil)
	}()

	Convey("Destroying an user account", t, func() {
		// Upload profile pictures
		testUploadFileHandler("../test_assets/gopher_avatar.png", "account_picture", "/", UpdateProfilePicture, conn, func(r *http.Request) {
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.Header.Add("X-User-Token", token.Hash)
		}, func(res *httptest.ResponseRecorder) {
			So(res.Code, ShouldEqual, 200)
		})

		// Uploading a post with a photo
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

		FollowUser(user.ID, bson.NewObjectId(), conn)
		FollowUser(bson.NewObjectId(), user.ID, conn)

		BlockUser(user.ID, bson.NewObjectId(), conn)
		BlockUser(bson.NewObjectId(), user.ID, conn)

		cmt := NewComment(user.ID, bson.NewObjectId())
		err := cmt.Save(conn)
		So(err, ShouldEqual, nil)

		post := NewPost(PostStatus, user)
		post.Text = "A fancy post"
		post.Privacy = PrivacySettings{Type: PrivacyPublic}

		err = post.Save(conn)
		So(err, ShouldEqual, nil)

		testPostHandler(LikePost, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
		}, conn, "/:id", "/"+post.ID.Hex(), func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 200)
			So(errResp.Message, ShouldEqual, "Post liked successfully")
		})

		err = SendNotification(NotificationPostCommented, user, bson.NewObjectId(), bson.NewObjectId(), conn)
		So(err, ShouldEqual, nil)

		testDeleteHandler(DestroyAccount, func(r *http.Request) {
			r.Header.Add("X-User-Token", token.Hash)
			if r.PostForm == nil {
				r.PostForm = make(url.Values)
			}
			r.PostForm.Add("confirmed", "true")
		}, conn, "/", "/", func(res *httptest.ResponseRecorder) {
			var errResp errorResponse
			if err := json.Unmarshal(res.Body.Bytes(), &errResp); err != nil {
				panic(err)
			}
			So(res.Code, ShouldEqual, 200)
			So(errResp.Message, ShouldEqual, "User account has been successfully destroyed")
		})

		files := 0
		filepath.Walk("../test_assets/", func(path string, fi os.FileInfo, _ error) error {
			if fi.Name() != ".DS_Store" && fi.Name() != "test_assets" {
				files++
			}
			return nil
		})

		So(files, ShouldEqual, 3)

		for _, col := range []string{"tokens", "users", "comments", "timelines", "posts", "follows", "blocks", "notifications", "likes"} {
			count, err := conn.C(col).Count()
			So(count, ShouldEqual, 0)
			So(err, ShouldEqual, nil)
		}
	})
}
