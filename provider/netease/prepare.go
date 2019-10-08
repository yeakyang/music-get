package netease

import (
	"github.com/winterssy/music-get/provider"
)

func prepare(songs []*Song, savePath string) ([]*provider.MP3, error) {
	n := len(songs)
	ids := make([]int, 0, n)
	for _, i := range songs {
		ids = append(ids, i.Id)
	}

	req := NewSongURLRequest(ids...)
	if err := req.Do(); err != nil {
		return nil, err
	}

	codeMap, urlMap := make(map[int]int, n), make(map[int]string, n)
	for _, i := range req.Response.Data {
		codeMap[i.Id] = i.Code
		urlMap[i.Id] = i.URL
	}

	mp3List := make([]*provider.MP3, 0, n)
	for _, i := range songs {
		mp3 := i.resolve()
		mp3.SavePath = savePath
		mp3.Playable = codeMap[i.Id] == 200
		mp3.DownloadURL = urlMap[i.Id]
		mp3List = append(mp3List, mp3)
	}

	return mp3List, nil
}
