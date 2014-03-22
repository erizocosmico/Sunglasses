package mask

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/martini-contrib/render"
	"net/http"
	"strconv"
	"time"
)

// RenderError renders an error message
func RenderError(resp render.Render, code, status int, message string) {
	resp.JSON(status, map[string]interface{}{
		"error":   true,
		"single":  true,
		"code":    code,
		"message": message,
	})
}

// RenderErrors renders a json response with an array of errors
func RenderErrors(resp render.Render, status int, codes []int, messages []string) {
	resp.JSON(status, map[string]interface{}{
		"error":    true,
		"single":   false,
		"messages": messages,
		"codes":    codes,
	})
}

// ListCountParams returns the count and offset parameters from the request
func ListCountParams(r *http.Request) (int, int) {
	var (
		count, offset int64
		err           error
	)

	if count, err = strconv.ParseInt(r.FormValue("count"), 10, 8); err != nil {
		count = 25
	}

	if count > 100 || count < 5 {
		count = 25
	}

	if offset, err = strconv.ParseInt(r.FormValue("offset"), 10, 8); err != nil {
		offset = 0
	}

	return int(count), int(offset)
}

func crypt(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 10)
	if err != nil {
		return "", err
	}

	return string(bytes[:]), nil
}

func randomString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)

	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}

	return string(bytes)
}

// NewRandomHash returns a random hash
func NewRandomHash() string {
	s := randomString(25) + fmt.Sprint(time.Now().UnixNano())
	hasher := sha512.New()
	hasher.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// Hash hashes a string
func Hash(h string) string {
	hasher := sha512.New()
	hasher.Write([]byte(h))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
