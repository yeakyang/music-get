package handler

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/provider/migu"
	"github.com/winterssy/music-get/provider/netease"
	"github.com/winterssy/music-get/provider/qq"
)

const (
	URLPattern     = "music.163.com|y.qq.com|music.migu.cn"
	NetEasePattern = "/(song|artist|album|playlist)\\?id=(\\d+)"
	QQPattern      = "/(song|singer|album|playsquare|playlist)/(\\w+)\\.html"
	MiguPattern    = "/v3/music/(song|artist|album|playlist)/(\\d+)"
)

func Parse(url string) (req provider.MusicRequest, err error) {
	re := regexp.MustCompile(URLPattern)
	matched, ok := re.FindString(url), re.MatchString(url)
	if !ok {
		err = errors.New("unsupported music address")
		return
	}

	switch matched {
	case "music.163.com":
		req, err = parseNetEase(url)
	case "y.qq.com":
		req, err = parseQQ(url)
	case "music.migu.cn":
		req, err = parseMigu(url)
	}

	return
}

func parseNetEase(url string) (req provider.MusicRequest, err error) {
	easylog.Debug("Use netease music parser")
	re := regexp.MustCompile(NetEasePattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok {
		err = errors.New("invalid netease music address")
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
	easylog.Debug("Use qq music parser")
	re := regexp.MustCompile(QQPattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok || len(matched) < 3 {
		err = errors.New("invalid qq music address")
		return
	}

	switch matched[1] {
	case "song":
		req = qq.NewSongRequest(matched[2])
	case "singer":
		req = qq.NewArtistRequest(matched[2])
	case "album":
		req = qq.NewAlbumRequest(matched[2])
	case "playsquare", "playlist":
		req = qq.NewPlaylistRequest(matched[2])
	}

	return
}

func parseMigu(url string) (req provider.MusicRequest, err error) {
	easylog.Debug("Use migu music parser")
	re := regexp.MustCompile(MiguPattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok || len(matched) < 3 {
		err = errors.New("invalid migu music address")
		return
	}

	switch matched[1] {
	case "song":
		req = migu.NewSongRequest(matched[2])
	case "artist":
		req = migu.NewArtistRequest(matched[2])
	case "album":
		req = migu.NewAlbumRequest(matched[2])
	case "playlist":
		req = migu.NewPlaylistRequest(matched[2])
	}

	return
}
