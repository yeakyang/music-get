package tencent

import (
	"fmt"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/common"
)

const (
	SongDownloadURL = "http://aqqmusic.tc.qq.com/amobile.music.tc.qq.com/M500%s.mp3?guid=%s&vkey=%s&uin=0&fromtag=8"
)

func ExtractMP3List(songs []Song, savePath string) ([]*common.MP3, error) {
	guid := "7332953645"
	mp3List := make([]*common.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.Extract()
		vkey, err := getVkey(guid, i.Mid)
		if err != nil {
			easylog.Errorf("get vkey failed: %s", i.Mid)
			continue
		}
		mp3.DownloadURL = fmt.Sprintf(SongDownloadURL, i.Mid, guid, vkey)
		mp3.SavePath = savePath
		mp3List = append(mp3List, mp3)
	}
	return mp3List, nil
}
