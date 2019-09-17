package netease

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	album := strings.TrimSpace(s.Album.Name)
	publishTime := time.Unix(s.PublishTime/1000, s.PublishTime%1000*1000*1000)
	year := strconv.Itoa(publishTime.Year())
	track := strconv.Itoa(s.Position)
	coverImage := s.Album.PicURL

	artistList := make([]string, 0, len(s.Artist))
	for _, ar := range s.Artist {
		artistList = append(artistList, strings.TrimSpace(ar.Name))
	}
	artist := strings.Join(artistList, "/")

	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.mp3", strings.Join(artistList, " "), title))
	tag := provider.Tag{
		Title:         title,
		Artist:        artist,
		Album:         album,
		Year:          year,
		Track:         track,
		CoverImageURL: coverImage,
	}

	return &provider.MP3{
		FileName: fileName,
		Tag:      tag,
		Provider: provider.NetEaseMusic,
	}
}
