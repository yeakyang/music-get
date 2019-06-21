package tencent

import (
	"fmt"

	"github.com/winterssy/music-get/common"
	"github.com/winterssy/music-get/config"
)

const (
	SelfSongDownloadURL  = "http://dl.stream.qqmusic.qq.com/M500%s.mp3?guid=%s&vkey=%s&fromtag=1"
	ThirdSongDownloadAPI = "https://v1.itooi.cn/tencent/url?id=%s&quality=%d"
)

func ExtractMP3List(songs []Song, savePath string) ([]*common.MP3, error) {
	// 测试发现 guid 可以是随机字符串
	guid := "yqq"
	vKey, err := getVKey(guid)
	if err != nil {
		return nil, err
	}

	br := config.MP3DownloadBr
	mp3List := make([]*common.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.Extract()
		mp3.SavePath = savePath
		switch br {
		case 192, 320:
			if mp3.Playable {
				mp3.DownloadURL = fmt.Sprintf(ThirdSongDownloadAPI, i.Mid, br)
			} else {
				mp3.DownloadURL = fmt.Sprintf(SelfSongDownloadURL, i.Mid, guid, vKey)
				mp3.Playable = true
			}
		default:
			mp3.DownloadURL = fmt.Sprintf(SelfSongDownloadURL, i.Mid, guid, vKey)
		}
		mp3List = append(mp3List, mp3)
	}

	return mp3List, nil
}
