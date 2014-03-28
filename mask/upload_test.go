package mask

import (
	. "github.com/smartystreets/goconvey/convey"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
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

func TestRetrieveUploadedImage(t *testing.T) {
	Convey("Retrievng uploaded images", t, func() {
		Convey("When no file was sent", func() {
			testUploadFileHandler("", "image", "/", func(r *http.Request) {
				f, err := RetrieveUploadedImage(r, "image")

				So(err, ShouldNotEqual, nil)
				So(f, ShouldEqual, nil)
				So(err.Error(), ShouldEqual, "no file was uploaded")
			}, nil, nil, nil)
		})

		Convey("When file is too large", func() {
			f, err := os.Create("../test_assets/large_file.txt")
			if err != nil {
				panic(err)
			}

			f.WriteString(randomString(20000000))
			f.Close()

			testUploadFileHandler("../test_assets/large_file.txt", "image", "/", func(r *http.Request) {
				f, err := RetrieveUploadedImage(r, "image")

				So(err, ShouldNotEqual, nil)
				So(f, ShouldEqual, nil)
				//So(err.Error(), ShouldEqual, "file too large")
			}, nil, nil, nil)

			os.Remove("../test_assets/large_file.txt")
		})

		Convey("When everything is OK", func() {
			testUploadFileHandler("../test_assets/gopher.jpg", "image", "/", func(r *http.Request) {
				f, err := RetrieveUploadedImage(r, "image")

				So(err, ShouldEqual, nil)
				So(f, ShouldNotEqual, nil)
			}, nil, nil, nil)
		})
	})
}
