package tencent

import (
	"fmt"
	"github.com/winterssy/music-get/pkg/easylog"

	"github.com/winterssy/music-get/common"
)

const (
	SongDownloadURL = "http://dl.stream.qqmusic.qq.com/%s?guid=%s&vkey=%s&fromtag=66"
)

func ExtractMP3List(songs []Song, savePath string) ([]*common.MP3, error) {
	// 测试发现 guid 可以是随机字符串
	guid := "yqq"
	mp3List := make([]*common.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.Extract()
		vkey, filename, err := getVKey(guid, i.Mid, "M500", "mp3")
		if err != nil || vkey == "" {
			easylog.Errorf("get vkey failed: %s", i.Mid)
			continue
		}
		mp3.DownloadURL = fmt.Sprintf(SongDownloadURL, filename, guid, vkey)
		mp3.SavePath = savePath
		mp3List = append(mp3List, mp3)
	}

	return mp3List, nil
}
