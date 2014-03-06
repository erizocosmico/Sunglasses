package lamp

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewConfig(t *testing.T) {
	Convey("Subject: Loading the configuration", t, func() {
		Convey("Given a non-existent file", func() {
			var path = "/non/existent/file.json"

			Convey("The function must return an error", func() {
				_, err := NewConfig(path)
				So(err, ShouldNotEqual, nil)
			})
		})

		Convey("Given a correct file", func() {
			var path = "../config.sample.json"

			Convey("The function must not return an error", func() {
				config, err := NewConfig(path)
				So(err, ShouldEqual, nil)
				So(config.URL, ShouldEqual, "localhost:3000")
				So(config.Port, ShouldEqual, ":3000")
				So(config.Debug, ShouldEqual, true)
				So(config.DatabaseName, ShouldEqual, "lamp_test")
				So(config.DatabaseUrl, ShouldEqual, "127.0.0.1:27017")
				So(config.RedisAddress, ShouldEqual, ":6379")
				So(config.StaticContentPath, ShouldEqual, "/path/to/static/content")
				So(config.SecretKey, ShouldEqual, "my fancy secret key")
			})
		})
	})
}
