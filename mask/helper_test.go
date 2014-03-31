package mask

import (
	"bytes"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type errorResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

func testMartini() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Static("public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

func testGetHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Get(handlerUrl, handler)
	}, middleware, conn, reqUrl, "GET", testFunc)
}

func testPostHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Post(handlerUrl, handler)
	}, middleware, conn, reqUrl, "POST", testFunc)
}

func testPutHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Put(handlerUrl, handler)
	}, middleware, conn, reqUrl, "PUT", testFunc)
}

func testDeleteHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Delete(handlerUrl, handler)
	}, middleware, conn, reqUrl, "DELETE", testFunc)
}

func uploadFile(file, key, url string) (*http.Request, error) {
	var b bytes.Buffer
	contentType := "application/octet-stream"

	if file != "" {
		w := multipart.NewWriter(&b)

		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		fw, err := w.CreateFormFile(key, file)
		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(fw, f); err != nil {
			return nil, err
		}

		if fw, err = w.CreateFormFile(key, file); err != nil {
			return nil, err
		}

		w.Close()

		contentType = w.FormDataContentType()
	}

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprint(req.ContentLength))

	return req, nil
}

func testUploadFileHandler(file, key, url string, handler martini.Handler, conn *Connection, middleware func(*http.Request), testFunc func(*httptest.ResponseRecorder)) {
	config, err := NewConfig("../config.sample.json")
	if err != nil {
		panic(err)
	}

	req, err := uploadFile(file, key, url)
	if err != nil {
		panic(err)
	}

	config.StorePath = "../test_assets/"
	config.ThumbnailStorePath = "../test_assets/"

	m := testMartini()
	m.Map(conn)
	m.Map(config)
	m.Use(render.Renderer())
	store := sessions.NewCookieStore([]byte("secret123"))
	store.Options(sessions.Options{
		MaxAge:   0,
		Secure:   false,
		HttpOnly: true,
	})
	m.Use(sessions.Sessions("my_session", store))
	if middleware != nil {
		m.Use(middleware)
	}
	m.Use(CreateContext)
	m.Post(url, handler)
	response := httptest.NewRecorder()
	m.ServeHTTP(response, req)
	if testFunc != nil {
		testFunc(response)
	}

}

func testHandler(methHandler func(*martini.ClassicMartini), middleware martini.Handler, conn *Connection, reqUrl, method string, testFunc func(*httptest.ResponseRecorder)) {
	config, err := NewConfig("../config.sample.json")
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(method, reqUrl, nil)
	m := testMartini()
	m.Map(conn)
	m.Map(config)
	m.Use(render.Renderer())
	store := sessions.NewCookieStore([]byte("secret123"))
	store.Options(sessions.Options{
		MaxAge:   0,
		Secure:   false,
		HttpOnly: true,
	})
	m.Use(sessions.Sessions("my_session", store))
	if middleware != nil {
		m.Use(middleware)
	}
	m.Use(CreateContext)
	methHandler(m)
	response := httptest.NewRecorder()
	m.ServeHTTP(response, req)
	testFunc(response)
}

func getConnection() *Connection {
	config, err := NewConfig("../config.sample.json")
	if err != nil {
		panic(err)
	}
	conn, err := NewDatabaseConn(config)
	if err != nil {
		panic(err)
	}

	return conn
}

func createRequestUser(conn *Connection) (*User, *Token) {
	user := NewUser()
	user.Username = "testing"

	if err := user.Save(conn); err != nil {
		panic(err)
	}

	token := new(Token)
	token.Type = UserToken
	token.Expires = float64(time.Now().Unix() + int64(3600*time.Second))
	token.UserID = user.ID
	if err := token.Save(conn); err != nil {
		panic(err)
	}

	return user, token
}
