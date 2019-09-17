package qq

import (
	"fmt"
	"strings"
	"time"

	"github.com/winterssy/music-get/provider"

	"github.com/winterssy/music-get/utils"
)

const (
	AlbumPicURL = "https://y.gtimg.cn/music/photo_new/T002R300x300M000%s.jpg"
)

type Singer struct {
	Id   int    `json:"id"`
	Mid  string `json:"mid"`
	Name string `json:"name"`
}

type Album struct {
	Id   int    `json:"id"`
	Mid  string `json:"mid"`
	Name string `json:"name"`
}

type GetAlbumInfo struct {
	FAlbumId   string `json:"Falbum_id"`
	FAlbumMid  string `json:"Falbum_mid"`
	FAlbumName string `json:"Falbum_name"`
}

type Song struct {
	Id         int      `json:"id"`
	Mid        string   `json:"mid"`
	Title      string   `json:"title"`
	Singer     []Singer `json:"singer"`
	Album      Album    `json:"album"`
	IndexAlbum int      `json:"index_album"`
	TimePublic string   `json:"time_public"`
	Action     struct {
		Switch int `json:"switch"`
	} `json:"action"`
}

type CD struct {
	DissTid  string `json:"disstid"`
	DissName string `json:"dissname"`
	SongList []Song `json:"songlist"`
}

func (s *Song) Extract() *provider.MP3 {
	title, album := strings.TrimSpace(s.Title), strings.TrimSpace(s.Album.Name)
	playable := s.Action.Switch != 65537
	publishTime, _ := time.Parse("2006-01-02", s.TimePublic)
	year, track := fmt.Sprintf("%d", publishTime.Year()), fmt.Sprintf("%d", s.IndexAlbum)
	coverImage := fmt.Sprintf(AlbumPicURL, s.Album.Mid)

	artistList := make([]string, 0, len(s.Singer))
	for _, ar := range s.Singer {
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
		Playable: playable,
		Tag:      tag,
		Provider: provider.QQMusic,
	}
}
