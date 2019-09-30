package qq

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
)

const (
	SongDownloadURL = "http://aqqmusic.tc.qq.com/amobile.music.tc.qq.com/M500%s.mp3?guid=%s&vkey=%s&uin=0&fromtag=8"
	BatchSongsCount = 10
)

func prepare(songs []Song, savePath string) ([]*provider.MP3, error) {
	n := len(songs)
	midMap := make(map[string]string, n)
	vkeyMap := make(map[string]string, n)

	guid := "7332953645"
	re := regexp.MustCompile("vkey=(\\w+)")
	var defaultKey string
	for i := 0; i < n; i += BatchSongsCount {
		j := i + BatchSongsCount
		if j > n {
			j = n
		}

		mids := make([]string, 0, j-i)
		for k := i; k < j; k++ {
			mids = append(mids, songs[k].Mid)
		}

		req := NewSongURLRequest(guid, mids...)
		if err := req.Do(); err != nil {
			return nil, err
		}

		if defaultKey == "" {
			matched, ok := re.FindStringSubmatch(req.Response.Req0.Data.TestFile2g), re.MatchString(req.Response.Req0.Data.TestFile2g)
			if ok {
				defaultKey = matched[1]
			}
		}

		for _, i := range req.Response.Req0.Data.MidURLInfo {
			if len(i.FileName) > 4 {
				midMap[i.SongMid] = i.FileName[4 : len(i.FileName)-len(filepath.Ext(i.FileName))]
			} else {
				midMap[i.SongMid] = i.SongMid
			}
			vkeyMap[i.SongMid] = i.Vkey
		}
	}

	for k, v := range vkeyMap {
		if v == "" {
			easylog.Debugf("Vkey not found, use default value: %s", k)
			vkeyMap[k] = defaultKey
		}
	}

	mp3List := make([]*provider.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.resolve()
		if vkeyMap[i.Mid] == "" {
			easylog.Errorf("get vkey failed: %s [%s]", mp3.FileName, i.Mid)
			continue
		}
		mp3.DownloadURL = fmt.Sprintf(SongDownloadURL, midMap[i.Mid], guid, vkeyMap[i.Mid])
		mp3.SavePath = savePath
		mp3List = append(mp3List, mp3)
	}
	return mp3List, nil
}
