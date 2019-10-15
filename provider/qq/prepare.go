package qq

import (
	"github.com/winterssy/music-get/provider"
)

const (
	SongDownloadURL = "http://aqqmusic.tc.qq.com/amobile.music.tc.qq.com/M500%s.mp3?guid=%s&vkey=%s&uin=0&fromtag=8"
	BatchSongsCount = 10
)

func prepare(songs []*Song, savePath string) ([]*provider.MP3, error) {
	n := len(songs)
	urlMap := make(map[string]string, n)

	guid := "7332953645"
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

		sip := req.Response.Req0.Data.Sip
		for _, i := range req.Response.Req0.Data.MidURLInfo {
			if i.PURL != "" {
				urlMap[i.SongMid] = sip[0] + i.PURL
			}
		}
	}

	mp3List := make([]*provider.MP3, 0, len(songs))
	for _, i := range songs {
		mp3 := i.resolve()
		mp3.DownloadURL = urlMap[i.Mid]
		mp3.SavePath = savePath
		mp3List = append(mp3List, mp3)
	}
	return mp3List, nil
}
