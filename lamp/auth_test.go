package lamp

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAccessTokenValidation(t *testing.T) {
	Convey("Subject: Testing access token validation", t, func() {
		config, err := NewConfig("../config.sample.json")
		if err != nil {
			panic(err)
		}
		conn, err := NewDatabaseConn(config)
		if err != nil {
			panic(err)
		}

		Convey("When an invalid access token is given", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(request *http.Request) {
				request.Header.Add("X-Access-Token", "")
			})
			m.Get("/", ValidateAccessToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 403", func() {
				So(response.Code, ShouldEqual, 403)
			})
		})

		Convey("When a valid access token is given", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			token := new(Token)
			token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
			token.Type = AccessToken

			success, err := token.Save(conn)
			if err != nil || !success {
				panic(err)
			}

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(request *http.Request) {
				request.Header.Add("X-Access-Token", token.ID)
			})
			m.Get("/", ValidateAccessToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 200", func() {
				So(response.Code, ShouldEqual, 200)

				token.Expires = float64(time.Now().Add(-(AccessTokenExpirationHours + 1) * time.Hour).Unix())
				success, err = token.Save(conn)
				if err != nil || !success {
					panic(err)
				}

				m.ServeHTTP(response, req)

				Convey("But when the token is expired the response status will be 403", func() {
					So(response.Code, ShouldEqual, 403)
				})
			})

		})
	})
}

func TestUserTokenValidation(t *testing.T) {
	Convey("Subject: Testing user token validation", t, func() {
		config, err := NewConfig("../config.sample.json")
		if err != nil {
			panic(err)
		}
		conn, err := NewDatabaseConn(config)
		if err != nil {
			panic(err)
		}

		Convey("When an invalid user token is given", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(request *http.Request) {
				request.Header.Add("X-User-Token", "")
			})
			m.Get("/", ValidateUserToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 403", func() {
				So(response.Code, ShouldEqual, 403)
			})
		})

		Convey("When a valid user token is given", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			token := new(Token)
			token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
			token.Type = UserToken

			success, err := token.Save(conn)
			if err != nil || !success {
				panic(err)
			}

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(request *http.Request) {
				request.Header.Add("X-User-Token", token.ID)
			})
			m.Get("/", ValidateUserToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 200", func() {
				So(response.Code, ShouldEqual, 200)

				token.Expires = float64(time.Now().Add(-(AccessTokenExpirationHours + 1) * time.Hour).Unix())
				success, err = token.Save(conn)
				if err != nil || !success {
					panic(err)
				}

				m.ServeHTTP(response, req)

				Convey("But when the token is expired the response status will be 403", func() {
					So(response.Code, ShouldEqual, 403)
				})
			})

		})
	})
}

func TestGetAccessToken(t *testing.T) {
	Convey("Subject: Getting access token", t, func() {
		config, err := NewConfig("../config.sample.json")
		if err != nil {
			panic(err)
		}
		conn, err := NewDatabaseConn(config)
		if err != nil {
			panic(err)
		}

		Convey("When we request an access token", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			m.Map(conn)
			m.Use(render.Renderer())
			m.Get("/", GetAccessToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 200 and we must receive a token", func() {
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
	Convey("Subject: Getting user token", t, func() {
		config, err := NewConfig("../config.sample.json")
		if err != nil {
			panic(err)
		}
		conn, err := NewDatabaseConn(config)
		if err != nil {
			panic(err)
		}

		user := new(User)
		user.Username = "Jane Doe"
		err = user.SetPassword("testing")
		if err != nil {
			panic(err)
		}
		user.Role = RoleUser
		user.Active = true
		success, err := user.Save(conn)
		if err != nil || !success {
			panic(err)
		}

		Convey("When we request an user token with valid data", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("POST", "/", nil)

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(req *http.Request) {
				req.PostForm.Add("username", "Jane Doe")
				req.PostForm.Add("password", "testing")
			})
			m.Post("/", GetUserToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 200 and we must receive a token", func() {
				var resultBody map[string]interface{}
				So(response.Code, ShouldEqual, 200)
				fmt.Println(response.Body)
				err := json.Unmarshal(response.Body.Bytes(), &resultBody)
				if err != nil {
					panic(err)
				}
				So(resultBody["user_token"].(string), ShouldNotEqual, "")
			})
		})

		Convey("When we request an user token with invalid data", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("POST", "/", nil)

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(req *http.Request) {
				req.PostForm.Add("username", "Jana Doe")
				req.PostForm.Add("password", "testing")
			})
			m.Post("/", GetUserToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 400", func() {
				So(response.Code, ShouldEqual, 400)
				_, _ = user.Remove(conn)
			})
		})
	})
}

func TestDestroyUserToken(t *testing.T) {
	Convey("Subject: Testing user token destruction", t, func() {
		config, err := NewConfig("../config.sample.json")
		if err != nil {
			panic(err)
		}
		conn, err := NewDatabaseConn(config)
		if err != nil {
			panic(err)
		}

		Convey("When a valid user token is given", func() {
			response := httptest.NewRecorder()
			m := martini.Classic()
			req, _ := http.NewRequest("GET", "/", nil)

			token := new(Token)
			token.Expires = float64(time.Now().Add(AccessTokenExpirationHours * time.Hour).Unix())
			token.Type = UserToken

			success, err := token.Save(conn)
			if err != nil || !success {
				panic(err)
			}

			m.Map(conn)
			m.Use(render.Renderer())
			m.Use(func(request *http.Request) {
				request.Header.Add("X-User-Token", token.ID)
			})
			m.Get("/", DestroyUserToken)

			m.ServeHTTP(response, req)

			Convey("The response status will be 200", func() {
				So(response.Code, ShouldEqual, 200)
			})

		})
	})
}
