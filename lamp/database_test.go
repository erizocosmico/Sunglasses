package lamp

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewDatabaseConn(t *testing.T) {
	Convey("Subject: Establishing a database connection", t, func() {
		Convey("Given a valid config", func() {
			config, err := NewConfig("../config.json")
			if err != nil {
				panic(err)
			}

			Convey("And a valid rethinkdb address", func() {
				conn, err := NewDatabaseConn(config)
				So(conn, ShouldNotEqual, nil)
				So(err, ShouldEqual, nil)
			})

			Convey("And an invalid rethinkdb address", func() {
				config.DatabaseUrl = "localhost:3456"
				conn, err := NewDatabaseConn(config)
				So(conn, ShouldEqual, nil)
				So(err, ShouldNotEqual, nil)
			})
		})
	})
}
