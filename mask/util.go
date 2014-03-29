package mask

import (
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

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

func isValidURL(URL string) bool {
	r := regexp.MustCompile("https?://[-A-Za-z0-9+&@]+\\.[a-zA-Z0-9\\.-]+([/#\\?&\\.-_a-zA-Z0-9%=,:;$\\(\\)]+)?")
	return r.MatchString(URL)
}

func responseTitle(resp *http.Response) string {
	var title string

	r := regexp.MustCompile("<title>(.*)</title>")
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	matches := r.FindStringSubmatch(string(bytes))
	if len(matches) > 1 {
		title = matches[1]
	} else {
		title = "Untitled"
	}

	return title
}

func isValidLink(URL string) (bool, string, string) {
	resp, err := http.Get(URL)
	if err != nil || resp.StatusCode != 200 {
		return false, "", ""
	}

	title := responseTitle(resp)

	return true, URL, title
}

func toLocalImagePath(url string, config *Config) string {
	return strings.Replace(url, config.WebStorePath, config.StorePath, -1)
}

func toLocalThumbnailPath(url string, config *Config) string {
	return strings.Replace(url, config.WebThumbnailStorePath, config.ThumbnailStorePath, -1)
}
