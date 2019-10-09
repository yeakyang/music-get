package kuwo

import (
	"fmt"

	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

type (
	Song struct {
		MusicRId    string `json:"musicrid"`
		RId         int    `json:"rid"`
		Name        string `json:"name"`
		Artist      string `json:"artist"`
		IsListenFee bool   `json:"isListenFee"`
	}

	Artist struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (s *Song) resolve() *provider.MP3 {
	fileName := utils.TrimInvalidFilePathChars(fmt.Sprintf("%s - %s.mp3", s.Artist, s.Name))
	return &provider.MP3{
		FileName: fileName,
		Playable: true,
		Provider: provider.KuwoMusic,
	}
}
