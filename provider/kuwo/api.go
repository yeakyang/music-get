package kuwo

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/sreq"
)

const (
	GetSongURL     = "http://www.kuwo.cn/url?format=mp3&response=url&type=convert_url3&br=128kmp3"
	GetSong        = "http://www.kuwo.cn/api/www/music/musicInfo"
	GetArtistInfo  = "http://www.kuwo.cn/api/www/artist/artist"
	GetArtistSongs = "http://www.kuwo.cn/api/www/artist/artistMusic?pn=1&rn=50"
	GetAlbum       = "http://www.kuwo.cn/api/www/album/albumInfo?pn=1&rn=9999"
	GetPlaylist    = "http://www.kuwo.cn/api/www/playlist/playListInfo?pn=1&rn=9999"
)

type (
	SongURLResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		URL  string `json:"url"`
	}

	SongURLRequest struct {
		Params   sreq.Params
		Response SongURLResponse
	}

	SongResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *Song  `json:"data"`
	}

	SongRequest struct {
		Params   sreq.Params
		Response SongResponse
	}

	ArtistResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			List []*Song `json:"list"`
		} `json:"data"`
	}

	ArtistRequest struct {
		artistId   string
		artistName string
		Params     sreq.Params
		Response   ArtistResponse
	}

	AlbumResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AlbumId   int     `json:"albumId"`
			Album     string  `json:"album"`
			MusicList []*Song `json:"musicList"`
		} `json:"data"`
	}

	AlbumRequest struct {
		Params   sreq.Params
		Response AlbumResponse
	}

	PlaylistResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Id        int     `json:"id"`
			Name      string  `json:"name"`
			MusicList []*Song `json:"musicList"`
		} `json:"data"`
	}

	PlaylistRequest struct {
		Params   sreq.Params
		Response PlaylistResponse
	}
)

func NewSongURLRequest(rid string) *SongURLRequest {
	params := sreq.Params{
		"rid": rid,
	}
	return &SongURLRequest{Params: params}
}

func (s *SongURLRequest) Do() error {
	easylog.Debug("SongURLRequest: send GetSongURL api request")
	err := request(GetSongURL,
		sreq.WithQuery(s.Params),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongURLRequest: GetSongURL api request error: %w", err)
	}

	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %d, %s",
			s.Response.Code, s.Response.Msg)
	}

	if s.Response.URL == "" {
		return errors.New("SongURLRequest: song url unavailable")
	}

	return nil
}

func NewSongRequest(mid string) *SongRequest {
	params := sreq.Params{
		"mid": mid,
	}
	return &SongRequest{Params: params}
}

func (s *SongRequest) RequireLogin() bool {
	return false
}

func (s *SongRequest) Login() error {
	panic("implement me")
}

func (s *SongRequest) Do() error {
	easylog.Debug("SongRequest: send GetSong api request")
	err := request(GetSong,
		sreq.WithQuery(s.Params),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongRequest: GetSong api request error: %w", err)
	}

	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("SongRequest: GetSong api status error: %d: %s",
			s.Response.Code, s.Response.Msg)
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	songs := []*Song{
		s.Response.Data,
	}
	return prepare(songs, ".")
}

func NewArtistRequest(artistId string) *ArtistRequest {
	params := sreq.Params{
		"artistid": artistId,
	}
	return &ArtistRequest{
		artistId: artistId,
		Params:   params,
	}
}

func (a *ArtistRequest) RequireLogin() bool {
	return false
}

func (a *ArtistRequest) Login() error {
	panic("implement me")
}

func (a *ArtistRequest) Do() error {
	var data struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data Artist `json:"data"`
	}

	easylog.Debug("ArtistRequest: send GetArtistInfo api request")
	err := request(GetArtistInfo,
		sreq.WithQuery(sreq.Params{
			"artistid": a.artistId,
		}),
	).JSON(&data)
	if err != nil {
		return err
	}

	if data.Code != http.StatusOK {
		return fmt.Errorf("ArtistRequest: GetArtistInfo api status error: %d: %s",
			data.Code, data.Msg)
	}

	a.artistName = data.Data.Name

	easylog.Debug("ArtistRequest: send GetArtistSongs api request")
	err = request(GetArtistSongs,
		sreq.WithQuery(a.Params),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api request error: %w", err)
	}

	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api status error: %d: %s",
			a.Response.Code, a.Response.Msg)
	}

	if len(a.Response.Data.List) == 0 {
		return errors.New("ArtistRequest: empty artist data")
	}

	return nil
}

func (a *ArtistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.artistName))
	return prepare(a.Response.Data.List, savePath)
}

func NewAlbumRequest(albumId string) *AlbumRequest {
	params := sreq.Params{
		"albumId": albumId,
	}
	return &AlbumRequest{
		Params: params,
	}
}

func (a *AlbumRequest) RequireLogin() bool {
	return false
}

func (a *AlbumRequest) Login() error {
	panic("implement me")
}

func (a *AlbumRequest) Do() error {
	easylog.Debug("AlbumRequest: send GetAlbum api request")
	err := request(GetAlbum,
		sreq.WithQuery(a.Params),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("AlbumRequest: GetAlbum api request error: %w", err)
	}

	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("AlbumRequest: GetAlbum api status error: %d: %s",
			a.Response.Code, a.Response.Msg)
	}

	if len(a.Response.Data.MusicList) == 0 {
		return errors.New("AlbumRequest: empty album data")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Data.Album))
	return prepare(a.Response.Data.MusicList, savePath)
}

func NewPlaylistRequest(pid string) *PlaylistRequest {
	params := sreq.Params{
		"pid": pid,
	}
	return &PlaylistRequest{
		Params: params,
	}
}

func (p *PlaylistRequest) RequireLogin() bool {
	return false
}

func (p *PlaylistRequest) Login() error {
	panic("implement me")
}

func (p *PlaylistRequest) Do() error {
	easylog.Debug("PlaylistRequest: send GetPlaylist api request")
	err := request(GetPlaylist,
		sreq.WithQuery(p.Params),
	).JSON(&p.Response)
	if err != nil {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api request error: %w", err)
	}

	if p.Response.Code != http.StatusOK {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api status error: %d: %s",
			p.Response.Code, p.Response.Msg)
	}

	if len(p.Response.Data.MusicList) == 0 {
		return errors.New("PlaylistRequest: empty playlist data")
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.Response.Data.Name))
	return prepare(p.Response.Data.MusicList, savePath)
}

func request(url string, opts ...sreq.RequestOption) *sreq.Response {
	return provider.Client(provider.KuwoMusic).
		Get(url, opts...).
		EnsureStatusOk()
}
