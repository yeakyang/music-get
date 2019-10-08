package kugou

import (
	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/pkg/concurrency"
	"github.com/winterssy/music-get/provider"
)

func prepare(songs []*Song, savePath string) ([]*provider.MP3, error) {
	mp3List := make([]*provider.MP3, len(songs))
	c := concurrency.New(16)
	for i, s := range songs {
		c.Add(1)
		go func(i int, song *Song) {
			defer c.Done()
			mp3 := song.resolve()
			mp3.SavePath = savePath
			req := NewSongURLRequest(song.Hash)
			if err := req.Do(); err != nil {
				mp3.Playable = false
				easylog.Errorf("Get song download url failed: %s: %s", song.Hash, err.Error())
			} else {
				mp3.Playable = req.Response.Status == 1
				mp3.DownloadURL = req.Response.URL[0]
			}
			mp3List[i] = mp3
		}(i, s)
	}
	c.Wait()
	return mp3List, nil
}
