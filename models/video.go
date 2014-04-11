package models

type VideoService int

const (
	// Service code
	VideoServiceYoutube = 1
	VideoServiceVimeo   = 2

	// Services regexps
	VideoServiceYoutubeRegexp = "youtube.com/watch?(.*)v=(.*)"
	VideoServiceVimeoRegexp   = "vimeo.com/([0-9]+)"
)
