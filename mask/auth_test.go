package mask

import (
	"encoding/json"
	"fmt"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestAccessTokenValidation(t *testing.T) {
	var (
		publicKey  = ""
		privateKey = ""
		timestamp  = fmt.Sprint(time.Now().Unix())
	)

	conn := getConnection()
	token := new(Token)
	token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
	token.Type = AccessToken

	if err := token.Save(conn); err != nil {
		panic(err)
	}

	Convey("Subject: Testing access token validation", t, func() {
		Convey("When an invalid signature is given", func() {
			testGetHandler(ValidateAccessToken, func(request *http.Request) {
				request.Header.Add("X-Access-Token", "")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/invalid"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid timestamp is given", func() {
			testGetHandler(ValidateAccessToken, func(request *http.Request) {
				request.Header.Add("X-Access-Token", "")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				timestamp := fmt.Sprint(time.Now().Unix() - 50000)
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid access token is given", func() {
			testGetHandler(ValidateAccessToken, func(request *http.Request) {
				request.Header.Add("X-Access-Token", "")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidAccessToken)
				So(errResp.Message, ShouldEqual, MsgInvalidAccessToken)
			})
		})

		Convey("When a valid access token is given", func() {
			testGetHandler(ValidateAccessToken, func(request *http.Request) {
				request.Header.Add("X-Access-Token", token.Hash)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldEqual, 200)
			})

		})

		Convey("But when the token is expired the response status will be 403", func() {
			token.Expires = float64(time.Now().Add(-(AccessTokenExpirationHours + 1) * time.Hour).Unix())
			if err := token.Save(conn); err != nil {
				panic(err)
			}

			testGetHandler(ValidateAccessToken, func(request *http.Request) {
				request.Header.Add("X-Access-Token", token.Hash)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidAccessToken)
				So(errResp.Message, ShouldEqual, MsgInvalidAccessToken)
			})
		})
	})
}

func TestAPIUserTokenValidation(t *testing.T) {
	var (
		publicKey  = ""
		privateKey = ""
		timestamp  = fmt.Sprint(time.Now().Unix())
	)

	conn := getConnection()
	token := new(Token)
	token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
	token.Type = UserToken

	if err := token.Save(conn); err != nil {
		panic(err)
	}
	Convey("Subject: Testing user token validation", t, func() {
		Convey("When an invalid signature is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", "")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/invalid"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid timestamp is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", "")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				timestamp := fmt.Sprint(time.Now().Unix() - 50000)
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid user token is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", "invalid token")
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidUserToken)
				So(errResp.Message, ShouldEqual, MsgInvalidUserToken)
			})
		})

		Convey("When a valid user token is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", token.Hash)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				So(resp.Code, ShouldEqual, 200)
			})

		})

		Convey("But when the token is expired the response status will be 403", func() {
			token.Expires = float64(time.Now().Add(-(AccessTokenExpirationHours + 1) * time.Hour).Unix())
			if err := token.Save(conn); err != nil {
				panic(err)
			}

			testGetHandler(ValidateUserToken, func(request *http.Request) {
				request.Header.Add("X-User-Token", token.Hash)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+privateKey+timestamp))
				request.Form.Add("api_key", publicKey)
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidUserToken)
				So(errResp.Message, ShouldEqual, MsgInvalidUserToken)
			})
		})
	})
}

func TestWebUserTokenValidation(t *testing.T) {
	var (
		csrfKey   = "1234"
		timestamp = fmt.Sprint(time.Now().Unix())
	)

	conn := getConnection()
	token := new(Token)
	token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
	token.Type = SessionToken

	if err := token.Save(conn); err != nil {
		panic(err)
	}
	Convey("Subject: Testing user token validation", t, func() {
		Convey("When an invalid signature is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request, s sessions.Session) {
				s.Set("user_token", token.Hash)
				s.Set("csrf_key", csrfKey)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/invalid"+csrfKey+timestamp))
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid timestamp is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request, s sessions.Session) {
				s.Set("user_token", token.Hash)
				s.Set("csrf_key", csrfKey)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				timestamp := fmt.Sprint(time.Now().Unix() - 50000)
				request.Form.Add("signature", Hash("/"+csrfKey+timestamp))
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 400)
				So(errResp.Code, ShouldEqual, CodeInvalidSignature)
				So(errResp.Message, ShouldEqual, MsgInvalidSignature)
			})
		})

		Convey("When an invalid user token is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request, s sessions.Session) {
				s.Set("user_token", "sjdhaksjdhsa")
				s.Set("csrf_key", csrfKey)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+csrfKey+timestamp))
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidUserToken)
				So(errResp.Message, ShouldEqual, MsgInvalidUserToken)
			})
		})

		Convey("When a valid user token is given", func() {
			testGetHandler(ValidateUserToken, func(request *http.Request, s sessions.Session) {
				s.Set("user_token", token.Hash)
				s.Set("csrf_key", csrfKey)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+csrfKey+timestamp))
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				fmt.Println(resp.Body)
				So(resp.Code, ShouldEqual, 200)
			})
		})

		Convey("But when the token is expired the response status will be 403", func() {
			token.Expires = float64(time.Now().Add(-(AccessTokenExpirationHours + 1) * time.Hour).Unix())
			if err := token.Save(conn); err != nil {
				panic(err)
			}

			testGetHandler(ValidateUserToken, func(request *http.Request, s sessions.Session) {
				s.Set("user_token", token.Hash)
				s.Set("csrf_key", csrfKey)
				if request.Form == nil {
					request.Form = make(url.Values)
				}
				request.Form.Add("signature", Hash("/"+csrfKey+timestamp))
				request.Form.Add("timestamp", timestamp)
			}, conn, "/", "/", func(resp *httptest.ResponseRecorder) {
				var errResp errorResponse
				if err := json.Unmarshal(resp.Body.Bytes(), &errResp); err != nil {
					panic(err)
				}
				So(resp.Code, ShouldEqual, 403)
				So(errResp.Code, ShouldEqual, CodeInvalidUserToken)
				So(errResp.Message, ShouldEqual, MsgInvalidUserToken)
			})
		})
	})
}

func TestGetAccessToken(t *testing.T) {
	Convey("Subject: Getting access token", t, func() {
		conn := getConnection()

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

		Convey("When we request an user token with valid data the response code will be 200 and we will receive a token", func() {
			testPostHandler(GetUserToken, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jane Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				var resultBody map[string]interface{}
				So(response.Code, ShouldEqual, 200)
				err := json.Unmarshal(response.Body.Bytes(), &resultBody)
				if err != nil {
					panic(err)
				}
				So(resultBody["user_token"].(string), ShouldNotEqual, "")
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
			var sess sessions.Session
			testPostHandler(func(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
				sess = s
				Login(req, conn, resp, s)
			}, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jane Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 200)
				So(sess.Get("user_token"), ShouldNotEqual, nil)
				sess.Set("user_token", nil)
			})
		})

		Convey("When we request an user token with invalid data the response code will be 400", func() {
			var sess sessions.Session
			testPostHandler(func(req *http.Request, conn *Connection, resp render.Render, s sessions.Session) {
				sess = s
				Login(req, conn, resp, s)
			}, func(req *http.Request) {
				if req.PostForm == nil {
					req.PostForm = make(url.Values)
				}
				req.PostForm.Add("username", "Jana Doe")
				req.PostForm.Add("password", "testing")
			}, conn, "/", "/", func(response *httptest.ResponseRecorder) {
				So(response.Code, ShouldEqual, 400)
				So(sess.Get("user_token"), ShouldEqual, nil)
				if err := user.Remove(conn); err != nil {
					panic(err)
				}
			})
		})
	})
}

func TestDestroyUserToken(t *testing.T) {
	Convey("Subject: Testing user token destruction", t, func() {
		conn := getConnection()

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
