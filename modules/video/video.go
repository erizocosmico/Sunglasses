package video

import (
	"net/http"
	"regexp"
	"github.com/mvader/mask/util"
	"github.com/mvader/mask/models"
)

// IsValidVideo determines if the video is valid or not
func IsValidVideo(URL string) (bool, string, models.VideoService, string) {
	var (
		valid     bool
		ID, title string
		service   models.VideoService
	)

	if valid, ID, service = isValidVideoURL(URL); !valid {
		return false, "", 0, ""
	}

	// Youtube does not reply with a not found status :-)
	if service == models.VideoServiceYoutube {
		URL = "http://gdata.youtube.com/feeds/api/videos/" + ID
	}

	resp, err := http.Get(URL)
	if err != nil || resp.StatusCode != 200 {
		return false, "", 0, ""
	}

	title = util.ResponseTitle(resp)

	return valid, ID, service, title
}

func isValidVideoURL(URL string) (bool, string, models.VideoService) {
	if ID := extractIDFromVideoService(models.VideoServiceYoutubeRegexp, URL, 2, 3); ID != "" {
		return true, ID, models.VideoServiceYoutube
	}

	if ID := extractIDFromVideoService(models.VideoServiceVimeoRegexp, URL, 1, 2); ID != "" {
		return true, ID, models.VideoServiceVimeo
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
