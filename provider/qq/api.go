package qq

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/sreq"
)

const (
	GetSongURL  = "https://u.y.qq.com/cgi-bin/musicu.fcg"
	GetSong     = "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg"
	GetArtist   = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_singer_track_cp.fcg"
	GetAlbum    = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_detail_cp.fcg"
	GetPlaylist = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_playlist_cp.fcg"
)

type (
	SongURLResponse struct {
		Code int `json:"code"`
		Req0 struct {
			Data struct {
				MidURLInfo []struct {
					FileName string `json:"filename"`
					SongMid  string `json:"songmid"`
					Vkey     string `json:"vkey"`
				} `json:"midurlinfo"`
				TestFile2g string `json:"testfile2g"`
			} `json:"data"`
		} `json:"req0"`
	}

	SongURLRequest struct {
		Params   sreq.Value
		Response SongURLResponse
	}

	SongResponse struct {
		Code int     `json:"code"`
		Data []*Song `json:"data"`
	}

	SongRequest struct {
		Params   sreq.Value
		Response SongResponse
	}

	SingerResponse struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				MusicData *Song `json:"musicData"`
			} `json:"list"`
			SingerId   string `json:"singer_id"`
			SingerMid  string `json:"singer_mid"`
			SingerName string `json:"singer_name"`
			Total      int    `json:"total"`
		} `json:"data"`
	}

	ArtistRequest struct {
		Params   sreq.Value
		Response SingerResponse
	}

	AlbumResponse struct {
		Code int
		Data struct {
			GetAlbumInfo GetAlbumInfo `json:"getAlbumInfo"`
			GetSongInfo  []*Song      `json:"getSongInfo"`
		} `json:"data"`
	}

	AlbumRequest struct {
		Params   sreq.Value
		Response AlbumResponse
	}

	PlaylistResponse struct {
		Code int `json:"code"`
		Data struct {
			CDList []CD `json:"cdlist"`
		} `json:"data"`
	}

	PlaylistRequest struct {
		Params   sreq.Value
		Response PlaylistResponse
	}
)

func NewSongURLRequest(guid string, mids ...string) *SongURLRequest {
	param := map[string]interface{}{
		"guid":      guid,
		"loginflag": 1,
		"songmid":   mids,
		"uin":       "0",
		"platform":  "20",
	}
	req0 := map[string]interface{}{
		"module": "vkey.GetVkeyServer",
		"method": "CgiGetVkey",
		"param":  param,
	}
	data := map[string]interface{}{
		"req0": req0,
	}

	enc, _ := json.Marshal(data)
	params := sreq.Value{
		"data": string(enc),
	}

	return &SongURLRequest{Params: params}
}

func (s *SongURLRequest) Do() error {
	easylog.Debug("SongURLRequest: send GetSongURL api request")
	err := request(GetSongURL,
		sreq.WithParams(s.Params),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongURLRequest: GetSongURL api request error: %w", err)
	}

	if s.Response.Code != 0 {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %d", s.Response.Code)
	}

	return nil
}

func NewSongRequest(mid string) *SongRequest {
	query := sreq.Value{
		"songmid":  mid,
		"platform": "yqq",
		"format":   "json",
	}
	return &SongRequest{Params: query}
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
		sreq.WithParams(s.Params),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongRequest: GetSong api request error: %w", err)
	}

	if s.Response.Code != 0 {
		return fmt.Errorf("SongRequest: GetSong api status error: %d", s.Response.Code)
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	return prepare(s.Response.Data, ".")
}

func NewSingerRequest(mid string) *ArtistRequest {
	query := sreq.Value{
		"singermid": mid,
		"begin":     "0",
		"num":       "50",
		"order":     "listen",
		"newsong":   "1",
		"platform":  "yqq",
	}
	return &ArtistRequest{Params: query}
}

func (a *ArtistRequest) RequireLogin() bool {
	return false
}

func (a *ArtistRequest) Login() error {
	panic("implement me")
}

func (a *ArtistRequest) Do() error {
	easylog.Debug("ArtistRequest: send GetArtist api request")
	err := request(GetArtist,
		sreq.WithParams(a.Params),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("ArtistRequest: GetArtist api request error: %w", err)
	}

	if a.Response.Code != 0 {
		return fmt.Errorf("ArtistRequest: GetArtist api status error: %d", a.Response.Code)
	}

	if len(a.Response.Data.List) == 0 {
		return errors.New("ArtistRequest: empty artist data")
	}

	return nil
}

func (a *ArtistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Data.SingerName))
	songs := make([]*Song, len(a.Response.Data.List))
	for i, s := range a.Response.Data.List {
		songs[i] = s.MusicData
	}
	return prepare(songs, savePath)
}

func NewAlbumRequest(mid string) *AlbumRequest {
	query := sreq.Value{
		"albummid": mid,
		"newsong":  "1",
		"platform": "yqq",
		"format":   "json",
	}
	return &AlbumRequest{Params: query}
}

func (a *AlbumRequest) RequireLogin() bool {
	return false
}

func (a *AlbumRequest) Login() error {
	panic("implement me")
}

func (a *AlbumRequest) Do() error {
	easylog.Debug("AlbumRequest: send album api request")
	err := request(GetAlbum,
		sreq.WithParams(a.Params),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("AlbumRequest: GetAlbum api request error: %w", err)
	}

	if a.Response.Code != 0 {
		return fmt.Errorf("AlbumRequest: GetAlbum api status error: %d", a.Response.Code)
	}

	if len(a.Response.Data.GetSongInfo) == 0 {
		return errors.New("AlbumRequest: empty album data")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Data.GetAlbumInfo.FAlbumName))
	return prepare(a.Response.Data.GetSongInfo, savePath)
}

func NewPlaylistRequest(id string) *PlaylistRequest {
	query := sreq.Value{
		"id":       id,
		"newsong":  "1",
		"platform": "yqq",
		"format":   "json",
	}
	return &PlaylistRequest{Params: query}
}

func (p *PlaylistRequest) RequireLogin() bool {
	return false
}

func (p *PlaylistRequest) Login() error {
	panic("implement me")
}

func (p *PlaylistRequest) Do() error {
	easylog.Debug("PlaylistRequest: send playlist api request")
	err := request(GetPlaylist,
		sreq.WithParams(p.Params),
	).JSON(&p.Response)
	if err != nil {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api request error: %w", err)
	}

	if p.Response.Code != 0 {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api status error: %d", p.Response.Code)
	}

	if len(p.Response.Data.CDList) == 0 {
		return errors.New("PlaylistRequest: empty playlist data")
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	var res []*provider.MP3
	for _, i := range p.Response.Data.CDList {
		savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(i.DissName))
		mp3List, err := prepare(i.SongList, savePath)
		if err != nil {
			continue
		}
		res = append(res, mp3List...)
	}

	return res, nil
}

func request(url string, opts ...sreq.RequestOption) *sreq.Response {
	return provider.Client(provider.QQMusic).
		Get(url, opts...).
		EnsureStatusOk()
}
