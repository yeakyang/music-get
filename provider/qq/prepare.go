package qq

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/winterssy/music-get/provider"
)

const (
	SongDownloadURL = "http://aqqmusic.tc.qq.com/amobile.music.tc.qq.com/M500%s.mp3?guid=%s&vkey=%s&uin=0&fromtag=8"
)

func prepare(songs []Song, savePath string) ([]*provider.MP3, error) {
	n := len(songs)
	// mids := make([]string, 0, n)
	// for _, i := range songs {
	// 	mids = append(mids, i.Mid)
	// }
	if n == 0 {
		return nil, errors.New("empty song list")
	}

	guid := "7332953645"
	req := NewSongURLRequest(guid, songs[0].Mid)
	if err := req.Do(); err != nil {
		return nil, err
	}

	re := regexp.MustCompile("vkey=(\\w+)")
	matched, ok := re.FindStringSubmatch(req.Response.Req0.Data.TestFile2g), re.MatchString(req.Response.Req0.Data.TestFile2g)
	if !ok {
		return nil, errors.New("get vkey failed")
	}
	defaultVkey := matched[1]

	// urlMap := make(map[string]string, n)
	// for _, i := range req.Response.Req0.Data.MidURLInfo {
	// 	if i.Vkey == "" {
	// 		urlMap[i.SongMid] = defaultVkey
	// 	} else {
	// 		urlMap[i.SongMid] = i.Vkey
	// 	}
	// }

	mp3List := make([]*provider.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.resolve()
		// if urlMap[i.Mid] == "" {
		// 	easylog.Errorf("get vkey failed: %s", i.Mid)
		// 	continue
		// }
		mp3.DownloadURL = fmt.Sprintf(SongDownloadURL, i.Mid, guid, defaultVkey)
		mp3.SavePath = savePath
		mp3List = append(mp3List, mp3)
	}
	return mp3List, nil
}
