package kugou

import (
	"fmt"

	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

type (
	Song struct {
		FileName string `json:"filename"`
		ExtName  string `json:"extname"`
		Hash     string `json:"hash"`
	}

	Artist struct {
		SingerId   int    `json:"singerid"`
		SingerName string `json:"singername"`
	}

	Album struct {
		AlbumId   int    `json:"albumid"`
		AlbumName string `json:"albumname"`
	}

	Playlist struct {
		SpecialId   int    `json:"specialid"`
		SpecialName string `json:"specialname"`
	}
)

func (s *Song) resolve() *provider.MP3 {
	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s.%s", s.FileName, s.ExtName))
	return &provider.MP3{
		FileName: fileName,
		Playable: true,
		Provider: provider.KugouMusic,
	}
}
