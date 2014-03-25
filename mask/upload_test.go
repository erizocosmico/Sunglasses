package mask

import (
	. "github.com/smartystreets/goconvey/convey"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"testing"
)

func loadFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func TestStoreImage(t *testing.T) {
	Convey("Testing the image storage", t, func() {
		Convey("When an invalid file format is given", func() {
			f, err := loadFile("../test_assets/file.md")
			if err != nil {
				panic(err)
			}

			img, thumb, err := StoreImage(f, UploadOptions{
				StorePath:          "../test_assets/",
				ThumbnailStorePath: "../test_assets/",
				MaxHeight:          500,
				MaxWidth:           500,
				ThumbnailHeight:    100,
				ThumbnailWidth:     100,
			})

			So(err, ShouldNotEqual, nil)
			So(img, ShouldEqual, "")
			So(thumb, ShouldEqual, "")
		})

		Convey("When file dimensions are too large", func() {
			f, err := loadFile("../test_assets/gopher.jpg")
			if err != nil {
				panic(err)
			}

			img, thumb, err := StoreImage(f, UploadOptions{
				StorePath:          "../test_assets/",
				ThumbnailStorePath: "../test_assets/",
				MaxHeight:          500,
				MaxWidth:           500,
				ThumbnailHeight:    100,
				ThumbnailWidth:     100,
			})

			So(err, ShouldNotEqual, nil)
			So(img, ShouldEqual, "")
			So(thumb, ShouldEqual, "")
		})

		Convey("When the file is ok", func() {
			f, err := loadFile("../test_assets/gopher.jpg")
			if err != nil {
				panic(err)
			}

			img, thumb, err := StoreImage(f, UploadOptions{
				StorePath:          "../test_assets/",
				ThumbnailStorePath: "../test_assets/",
				MaxHeight:          5000,
				MaxWidth:           5000,
				ThumbnailHeight:    100,
				ThumbnailWidth:     100,
			})

			So(err, ShouldEqual, nil)
			So(img, ShouldNotEqual, "")
			So(thumb, ShouldNotEqual, "")

			os.Remove(img)
			os.Remove(thumb)
		})
	})
}
