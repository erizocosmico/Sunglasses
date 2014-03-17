package mask

import (
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"net/http"
	"net/http/httptest"
	"time"
)

type errorResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
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

func testHandler(methHandler func(*martini.ClassicMartini), middleware martini.Handler, conn *Connection, reqUrl, method string, testFunc func(*httptest.ResponseRecorder)) {
	req, _ := http.NewRequest(method, reqUrl, nil)
	m := martini.Classic()
	m.Map(conn)
	m.Use(render.Renderer())
	store := sessions.NewCookieStore([]byte("secret123"))
	m.Use(sessions.Sessions("my_session", store))
	methHandler(m)
	response := httptest.NewRecorder()
	m.Use(middleware)
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
