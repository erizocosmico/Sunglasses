package lamp

import (
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"net/http/httptest"
)

func testGetHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Get(handlerUrl, handler)
	}, handler, middleware, conn, reqUrl, handlerUrl, testFunc)
}

func testPostHandler(handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	testHandler(func(m *martini.ClassicMartini) {
		m.Post(handlerUrl, handler)
	}, handler, middleware, conn, reqUrl, handlerUrl, testFunc)
}

func testHandler(methHandler func(*martini.ClassicMartini), handler, middleware martini.Handler, conn *Connection, reqUrl, handlerUrl string, testFunc func(*httptest.ResponseRecorder)) {
	req, _ := http.NewRequest("POST", reqUrl, nil)
	m := martini.Classic()
	m.Map(conn)
	m.Use(render.Renderer())
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
