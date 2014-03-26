package mask

import (
	"code.google.com/p/graphics-go/graphics"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
)

type UploadOptions struct {
	StorePath, ThumbnailStorePath                        string
	MaxHeight, MaxWidth, ThumbnailHeight, ThumbnailWidth int
}

// RetrieveUploadedImage returns the uploaded file at the given key
func RetrieveUploadedImage(r *http.Request, key string) (io.ReadCloser, error) {
	_, fh, err := r.FormFile(key)
	if err == nil && fh != nil {
		file, err := fh.Open()
		if err != nil {
			return nil, err
		}

		fi, err := file.(*os.File).Stat()
		if err != nil {
			return nil, err
		}

		if fi.Size() > 50000 {
			return nil, errors.New("file too large")
		}

		return file, nil
	}

	return nil, errors.New("no file was uploaded")
}

// StoreImage stores in disk a file received with the request
func StoreImage(file io.ReadCloser, options UploadOptions) (string, string, error) {
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", err
	}

	if format != "gif" && format != "png" && format != "jpg" && format != "jpeg" {
		return "", "", errors.New("invalid file format")
	}

	if img.Bounds().Max.X > options.MaxWidth || img.Bounds().Max.Y > options.MaxHeight {
		return "", "", errors.New("file dimensions are too large")
	}

	imagePath := options.StorePath + NewFileName(format)
	dst, err := os.Create(imagePath)
	defer dst.Close()
	if err != nil {
		return "", "", err
	}

	if err := writeFile(dst, img, format); err != nil {
		return "", "", err
	}

	thumbnail, err := generateThumbnail(img, options)
	if err != nil {
		return "", "", err
	}

	thumbnailPath := options.ThumbnailStorePath + NewFileName(format)
	thumbDst, err := os.Create(thumbnailPath)
	if err != nil {
		return "", "", err
	}

	if err := writeFile(thumbDst, thumbnail, format); err != nil {
		return "", "", nil
	}

	thumbDst.Close()

	return imagePath, thumbnailPath, nil
}

func generateThumbnail(src image.Image, options UploadOptions) (image.Image, error) {
	dst := image.NewRGBA(image.Rect(0, 0, options.ThumbnailWidth, options.ThumbnailHeight))
	if err := graphics.Thumbnail(dst, src); err != nil {
		return nil, err
	}

	return dst, nil
}

func writeFile(w io.Writer, i image.Image, format string) error {
	var err error

	switch format {
	case "png":
		err = png.Encode(w, i)
		break

	case "jpeg":
		err = jpeg.Encode(w, i, &jpeg.Options{Quality: 100})
		break

	case "gif":
		err = gif.Encode(w, i, nil)
		break
	}

	return err
}
