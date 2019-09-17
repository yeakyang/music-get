package handler

import (
	"regexp"
	"strconv"

	"github.com/winterssy/music-get/pkg/ecode"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/provider/netease"
	"github.com/winterssy/music-get/provider/qq"
)

const (
	URLPattern     = "music.163.com|y.qq.com"
	NetEasePattern = "/(song|artist|album|playlist)\\?id=(\\d+)"
	QQPattern      = "/(song|singer|album|playsquare|playlist)/(\\w+)\\.html"
)

func Parse(url string) (req provider.MusicRequest, err error) {
	re := regexp.MustCompile(URLPattern)
	matched, ok := re.FindString(url), re.MatchString(url)
	if !ok {
		err = ecode.NewError(ecode.ParseURLException, "handler.Parse")
		return
	}

	switch matched {
	case "music.163.com":
		req, err = parseNetEase(url)
	case "y.qq.com":
		req, err = parseQQ(url)
	}

	return
}

func parseNetEase(url string) (req provider.MusicRequest, err error) {
	re := regexp.MustCompile(NetEasePattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok {
		err = ecode.NewError(ecode.ParseURLException, "netease.Parse")
		return
	}

	id, err := strconv.Atoi(matched[2])
	if err != nil {
		return
	}

	switch matched[1] {
	case "song":
		req = netease.NewSongRequest(id)
	case "artist":
		req = netease.NewArtistRequest(id)
	case "album":
		req = netease.NewAlbumRequest(id)
	case "playlist":
		req = netease.NewPlaylistRequest(id)
	}

	return
}

func parseQQ(url string) (req provider.MusicRequest, err error) {
	re := regexp.MustCompile(QQPattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok || len(matched) < 3 {
		err = ecode.NewError(ecode.ParseURLException, "handler.parseQQ")
		return
	}

	switch matched[1] {
	case "song":
		req = qq.NewSongRequest(matched[2])
	case "singer":
		req = qq.NewSingerRequest(matched[2])
	case "album":
		req = qq.NewAlbumRequest(matched[2])
	case "playsquare", "playlist":
		req = qq.NewPlaylistRequest(matched[2])
	}

	return
}
