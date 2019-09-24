package handler

import (
	"reflect"
	"testing"

	"github.com/winterssy/music-get/provider/netease"
	"github.com/winterssy/music-get/provider/qq"
)

func TestParse(t *testing.T) {
	tests := []struct {
		url  string
		want reflect.Type
	}{
		{
			url:  "https://music.163.com/#/song?id=553310243",
			want: reflect.TypeOf(&netease.SongRequest{}),
		},
		{
			url:  "https://music.163.com/#/artist?id=13193",
			want: reflect.TypeOf(&netease.ArtistRequest{}),
		},
		{
			url:  "https://music.163.com/#/album?id=38373053",
			want: reflect.TypeOf(&netease.AlbumRequest{}),
		},
		{
			url:  "https://music.163.com/#/playlist?id=156934569",
			want: reflect.TypeOf(&netease.PlaylistRequest{}),
		},
		{
			url:  "https://y.qq.com/n/yqq/song/002Zkt5S2z8JZx.html",
			want: reflect.TypeOf(&qq.SongRequest{}),
		},
		{
			url:  "https://y.qq.com/n/yqq/singer/000Sp0Bz4JXH0o.html",
			want: reflect.TypeOf(&qq.SingerRequest{}),
		},
		{
			url:  "https://y.qq.com/n/yqq/album/002fRO0N4FftzY.html",
			want: reflect.TypeOf(&qq.AlbumRequest{}),
		},
		{
			url:  "https://y.qq.com/n/yqq/playsquare/5474239760.html",
			want: reflect.TypeOf(&qq.PlaylistRequest{}),
		},
		{
			url:  "https://y.qq.com/n/yqq/playlist/5474239760.html",
			want: reflect.TypeOf(&qq.PlaylistRequest{}),
		},
	}

	for _, test := range tests {
		req, _ := Parse(test.url)
		if got := reflect.TypeOf(req); got != test.want {
			t.Errorf("Parse %q got: %v, want: %v", test.url, got, test.want)
		}
	}
}
