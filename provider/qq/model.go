package qq

import (
	"fmt"
	"strings"

	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

type (
	Singer struct {
		Id   int    `json:"id"`
		Mid  string `json:"mid"`
		Name string `json:"name"`
	}

	Album struct {
		Id   int    `json:"id"`
		Mid  string `json:"mid"`
		Name string `json:"name"`
	}

	GetAlbumInfo struct {
		FAlbumId   string `json:"Falbum_id"`
		FAlbumMid  string `json:"Falbum_mid"`
		FAlbumName string `json:"Falbum_name"`
	}

	Song struct {
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

	CD struct {
		DissTid  string  `json:"disstid"`
		DissName string  `json:"dissname"`
		SongList []*Song `json:"songlist"`
	}
)

func (s *Song) resolve() *provider.MP3 {
	title := strings.TrimSpace(s.Title)
	// playable := s.Action.Switch != 65537

	artists := make([]string, 0, len(s.Singer))
	for _, ar := range s.Singer {
		artists = append(artists, strings.TrimSpace(ar.Name))
	}

	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.m4a", strings.Join(artists, " "), title))
	return &provider.MP3{
		FileName: fileName,
		Playable: true,
		Provider: provider.MiguMusic,
	}
}
