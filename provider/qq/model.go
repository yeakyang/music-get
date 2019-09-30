package qq

import (
	"fmt"
	"strings"

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

func (s *Song) resolve() *provider.MP3 {
	title := strings.TrimSpace(s.Title)
	// playable := s.Action.Switch != 65537
	playable := true

	artists := make([]string, 0, len(s.Singer))
	for _, ar := range s.Singer {
		artists = append(artists, strings.TrimSpace(ar.Name))
	}

	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.mp3", strings.Join(artists, " "), title))
	return &provider.MP3{
		FileName: fileName,
		Playable: playable,
		Provider: provider.QQMusic,
	}
}
