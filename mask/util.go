package mask

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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

// NewFileName returns an unique name for the file
func NewFileName(extension string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9]")
	return r.ReplaceAllString(NewRandomHash(), "") + "." + extension
}

// Hash hashes a string
func Hash(h string) string {
	hasher := sha512.New()
	hasher.Write([]byte(h))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func strlen(s string) int {
	return utf8.RuneCountInString(s)
}

func getBoolean(r *http.Request, key string) bool {
	if v := r.FormValue(key); v != "" {
		if strings.ToLower(v) == "true" || v == "1" {
			return true
		}
	}

	return false
}

func isValidURL(URL string) bool {
	r := regexp.MustCompile("https?://[-A-Za-z0-9+&@]+\\.[a-zA-Z0-9\\.-]+([/#\\?&\\.-_a-zA-Z0-9%=,:;$\\(\\)]+)?")
	return r.MatchString(URL)
}

func responseTitle(resp *http.Response) string {
	var title string

	r := regexp.MustCompile("<title>(.*)</title>")
	defer resp.Body.Close()

	matches := r.FindStringSubmatch(string(ioutil.ReadAll(resp.Body)))
	if len(matches) > 1 {
		title = matches[1]
	} else {
		title = "Untitled"
	}

	return title
}

func isValidLink(URL string) (bool, string, string) {
	var (
		valid bool
		title string
	)

	resp, err := http.Get(URL)
	if err != nil || resp.StatusCode != 200 {
		return false, "", ""
	}

	title = responseTitle(resp)

	return valid, URL, title
}
