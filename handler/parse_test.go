package handler

import (
	"reflect"
	"testing"

	"github.com/winterssy/music-get/provider/kugou"
	"github.com/winterssy/music-get/provider/migu"
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
			want: reflect.TypeOf(&qq.ArtistRequest{}),
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
		{
			url:  "http://music.migu.cn/v3/music/song/63273402938",
			want: reflect.TypeOf(&migu.SongRequest{}),
		},

		{
			url:  "http://music.migu.cn/v3/music/artist/208",
			want: reflect.TypeOf(&migu.ArtistRequest{}),
		},
		{
			url:  "http://music.migu.cn/v3/music/album/1121438701",
			want: reflect.TypeOf(&migu.AlbumRequest{}),
		},
		{
			url:  "http://music.migu.cn/v3/music/playlist/159248239",
			want: reflect.TypeOf(&migu.PlaylistRequest{}),
		},

		{
			url:  "https://www.kugou.com/song/#hash=1571941D82D63AD614E35EAD9DB6A6A2&album_id=976965",
			want: reflect.TypeOf(&kugou.SongRequest{}),
		},
		{
			url:  "https://www.kugou.com/singer/8965.html",
			want: reflect.TypeOf(&kugou.ArtistRequest{}),
		},
		{
			url:  "https://www.kugou.com/yy/album/single/976965.html",
			want: reflect.TypeOf(&kugou.AlbumRequest{}),
		},
		{
			url:  "https://www.kugou.com/yy/special/single/547134.html",
			want: reflect.TypeOf(&kugou.PlaylistRequest{}),
		},
	}

	for _, test := range tests {
		req, _ := Parse(test.url)
		if got := reflect.TypeOf(req); got != test.want {
			t.Errorf("Parse(%q) got: %v, want: %v", test.url, got, test.want)
		}
	}
}
