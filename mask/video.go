package mask

import (
	"net/http"
	"regexp"
)

type VideoService int

const (
	// Service code
	VideoServiceYoutube = 1
	VideoServiceVimeo   = 2

	// Services regexps
	VideoServiceYoutubeRegexp = "youtube.com/watch?(.*)v=(.*)"
	VideoServiceVimeoRegexp   = "vimeo.com/([0-9]+)"
)

func isValidVideo(URL string) (bool, string, VideoService, string) {
	var (
		valid     bool
		ID, title string
		service   VideoService
	)

	if valid, ID, service = isValidVideoURL(URL); !valid {
		return false, "", 0, ""
	}

	// Youtube does not reply with a not found status :-)
	if service == VideoServiceYoutube {
		URL = "http://gdata.youtube.com/feeds/api/videos/" + ID
	}

	resp, err := http.Get(URL)
	if err != nil || resp.StatusCode != 200 {
		return false, "", 0, ""
	}

	title = responseTitle(resp)

	return valid, ID, service, title
}

func isValidVideoURL(URL string) (bool, string, VideoService) {
	if ID := extractIDFromVideoService(VideoServiceYoutubeRegexp, URL, 2, 3); ID != "" {
		return true, ID, VideoServiceYoutube
	}

	if ID := extractIDFromVideoService(VideoServiceVimeoRegexp, URL, 1, 2); ID != "" {
		return true, ID, VideoServiceVimeo
	}

	return false, "", 0
}

func extractIDFromVideoService(service, URL string, idIndex, numMatches int) string {
	r := regexp.MustCompile(service)
	if r.MatchString(URL) {
		matches := r.FindStringSubmatch(URL)
		if len(matches) == numMatches {
			return matches[idIndex]
		}
	}

	return ""
}
