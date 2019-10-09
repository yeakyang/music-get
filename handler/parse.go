package handler

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/provider/kugou"
	"github.com/winterssy/music-get/provider/kuwo"
	"github.com/winterssy/music-get/provider/migu"
	"github.com/winterssy/music-get/provider/netease"
	"github.com/winterssy/music-get/provider/qq"
)

const (
	URLPattern     = "music.163.com|y.qq.com|music.migu.cn|www.kugou.com|www.kuwo.cn"
	NetEasePattern = "/(song|artist|album|playlist)\\?id=(\\d+)"
	QQPattern      = "/(song|singer|album|playsquare|playlist)/(\\w+)\\.html"
	MiguPattern    = "/v3/music/(song|artist|album|playlist)/(\\d+)"
	KugouPattern   = "/(song|singer|yy/album/single|yy/special/single)/(#hash=(\\w+)|(\\d+).html)"
	KuwoPattern    = "/(play_detail|singer_detail|album_detail|playlist_detail)/(\\d+)"
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
	case "www.kugou.com":
		req, err = parseKugou(url)
	case "www.kuwo.cn":
		req, err = parseKuwo(url)
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
	if !ok {
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
	if !ok {
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

func parseKugou(url string) (req provider.MusicRequest, err error) {
	easylog.Debug("Use kugou music parser")
	re := regexp.MustCompile(KugouPattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok {
		err = errors.New("invalid kugou music address")
		return
	}

	switch matched[1] {
	case "song":
		req = kugou.NewSongRequest(matched[3])
	case "singer":
		req = kugou.NewArtistRequest(matched[4])
	case "yy/album/single":
		req = kugou.NewAlbumRequest(matched[4])
	case "yy/special/single":
		req = kugou.NewPlaylistRequest(matched[4])
	}

	return
}

func parseKuwo(url string) (req provider.MusicRequest, err error) {
	easylog.Debug("Use kuwo music parser")
	re := regexp.MustCompile(KuwoPattern)
	matched, ok := re.FindStringSubmatch(url), re.MatchString(url)
	if !ok {
		err = errors.New("invalid kuwo music address")
		return
	}

	switch matched[1] {
	case "play_detail":
		req = kuwo.NewSongRequest(matched[2])
	case "singer_detail":
		req = kuwo.NewArtistRequest(matched[2])
	case "album_detail":
		req = kuwo.NewAlbumRequest(matched[2])
	case "playlist_detail":
		req = kuwo.NewPlaylistRequest(matched[2])
	}

	return
}
