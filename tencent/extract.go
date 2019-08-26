package tencent

import (
	"fmt"

	"github.com/winterssy/music-get/common"
)

const (
	SelfSongDownloadURL = "http://dl.stream.qqmusic.qq.com/M500%s.mp3?guid=%s&vkey=%s&fromtag=1"
)

func ExtractMP3List(songs []Song, savePath string) ([]*common.MP3, error) {
	// 测试发现 guid 可以是随机字符串
	guid := "yqq"
	vKey, err := getVKey(guid)
	if err != nil {
		return nil, err
	}

	mp3List := make([]*common.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.Extract()
		mp3.SavePath = savePath
		mp3.DownloadURL = fmt.Sprintf(SelfSongDownloadURL, i.Mid, guid, vKey)
		mp3List = append(mp3List, mp3)
	}

	return mp3List, nil
}
