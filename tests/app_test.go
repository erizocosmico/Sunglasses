package tests

import (
	"github.com/mvader/sunglasses/app"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	a, err := app.NewApp("../config.sample.json")
	if err != nil {
		panic(err)
	}

	defer func() {
		a.Connection.Session.Close()
	}()

	go a.Martini.Run()

	response, err := http.Get("http://localhost:3000/api/auth/access_token")

	Convey("Testing App", t, func() {
		So(response.StatusCode, ShouldEqual, 200)
	})

	time.Sleep(2 * time.Second)
}
