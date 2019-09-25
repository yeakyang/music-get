package netease

import (
	"fmt"
	"strings"

	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

type Artist struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	PicURL      string `json:"picURL"`
	PublishTime int64  `json:"publishTime"`
}

type SongURL struct {
	Id   int    `json:"id"`
	Code int    `json:"code"`
	URL  string `json:"url"`
}

type Song struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Artist      []Artist `json:"ar"`
	Album       Album    `json:"al"`
	Position    int      `json:"no"`
	PublishTime int64    `json:"publishTime"`
}

type TrackId struct {
	Id int `json:"id"`
}

type Playlist struct {
	Id       int       `json:"id"`
	Name     string    `json:"name"`
	TrackIds []TrackId `json:"trackIds"`
}

func (s *Song) resolve() *provider.MP3 {
	title := strings.TrimSpace(s.Name)

	artists := make([]string, 0, len(s.Artist))
	for _, ar := range s.Artist {
		artists = append(artists, strings.TrimSpace(ar.Name))
	}

	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.mp3", strings.Join(artists, " "), title))
	return &provider.MP3{
		FileName: fileName,
		Provider: provider.NetEaseMusic,
	}
}
