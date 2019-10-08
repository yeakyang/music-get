package migu

import (
	"fmt"
	"strings"

	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

type (
	Song struct {
		ResourceType string `json:"resourceType"`
		ContentId    string `json:"contentId"`
		CopyrightId  string `json:"copyrightId"`
		SongId       string `json:"songId"`
		SongName     string `json:"songName"`
		SingerId     string `json:"singerId"`
		Singer       string `json:"singer"`
		AlbumId      string `json:"albumId"`
		Album        string `json:"album"`
	}

	Album struct {
		ResourceType string `json:"resourceType"`
		AlbumId      string `json:"albumId"`
		Title        string `json:"title"`
		SongItems    []*Song
	}

	Playlist struct {
		ResourceType string `json:"resourceType"`
		MusicListId  string `json:"musicListId"`
		Title        string `json:"title"`
		SongItems    []*Song
	}

	Artist struct {
		ResourceType string `json:"resourceType"`
		SingerId     string `json:"singerId"`
		Singer       string `json:"singer"`
	}
)

func (s *Song) resolve() *provider.MP3 {
	title := strings.TrimSpace(s.SongName)
	artist := strings.ReplaceAll(s.Singer, "|", " ")
	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.mp3", artist, title))
	return &provider.MP3{
		FileName: fileName,
		Playable: true,
		Provider: provider.QQMusic,
	}
}
