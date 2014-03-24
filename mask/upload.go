package mask

import (
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
)

type UploadOptions struct {
	StorePath string
	MaxHeight int
	MaxWidth  int
}

// StoreUploadedFile stores in disk a file received with the request
func StoreUploadedFile(r *http.Request, key string, options UploadOptions) error {
	r.ParseMultiPartForm(100000)
	if files, ok := r.MultiPartForm.File[key]; ok {
		if len(files) > 0 {
			file, err := files[0].Open()
			defer file.Close()
			if err != nil {
				return err
			}

			fi, err := file.Stat()
			if err != nil {
				return err
			}

			if fi.Size() > 50000 {
				return errors.New("file too large")
			}

			cfg, format, err := image.DecodeConfig(file)
			if err != nil {
				return err
			}

			if format != "image/gif" && format != "image/png" && format != "image/jpg" && format != "image/jpeg" {
				return errors.New("invalid file format")
			}

			if cfg.Width > options.MaxWidth || cfg.Height > options.MaxHeight {
				return errors.New("file dimensions are too large")
			}

			dst, err := os.Create(options.StorePath + NewFileName(format[6:]))
			defer dst.Close()
			if err != nil {
				return err
			}

			if _, err := io.Copy(dst, file); err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("no file was uploaded")
}
