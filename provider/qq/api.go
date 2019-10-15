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
	GetSong     = "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg?platform=yqq&format=json"
	GetArtist   = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_singer_track_cp.fcg?begin=0&num=50&order=listen&newsong=1&platform=yqq&format=json"
	GetAlbum    = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_detail_cp.fcg?newsong=1&platform=yqq&format=json"
	GetPlaylist = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_playlist_cp.fcg?newsong=1&platform=yqq&format=json"
)

type (
	SongURLResponse struct {
		Code int `json:"code"`
		Req0 struct {
			Data struct {
				MidURLInfo []struct {
					FileName string `json:"filename"`
					PURL     string `json:"purl"`
					SongMid  string `json:"songmid"`
					Vkey     string `json:"vkey"`
				} `json:"midurlinfo"`
				Sip        []string `json:"sip"`
				TestFile2g string   `json:"testfile2g"`
			} `json:"data"`
		} `json:"req0"`
	}

	SongURLRequest struct {
		Params   sreq.Params
		Response SongURLResponse
	}

	SongResponse struct {
		Code int     `json:"code"`
		Data []*Song `json:"data"`
	}

	SongRequest struct {
		Params   sreq.Params
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
		Params   sreq.Params
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
		Params   sreq.Params
		Response AlbumResponse
	}

	PlaylistResponse struct {
		Code int `json:"code"`
		Data struct {
			CDList []CD `json:"cdlist"`
		} `json:"data"`
	}

	PlaylistRequest struct {
		Params   sreq.Params
		Response PlaylistResponse
	}
)

func NewSongURLRequest(guid string, songMids ...string) *SongURLRequest {
	param := map[string]interface{}{
		"guid":      guid,
		"loginflag": 1,
		"songmid":   songMids,
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
	params := sreq.Params{
		"data": string(enc),
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

	if s.Response.Code != 0 {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %d", s.Response.Code)
	}
	if len(s.Response.Req0.Data.Sip) == 0 {
		return errors.New("SongURLRequest: no sip")
	}

	return nil
}

func NewSongRequest(songMid string) *SongRequest {
	params := sreq.Params{
		"songmid": songMid,
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

	if s.Response.Code != 0 {
		return fmt.Errorf("SongRequest: GetSong api status error: %d", s.Response.Code)
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	return prepare(s.Response.Data, ".")
}

func NewArtistRequest(singerMid string) *ArtistRequest {
	params := sreq.Params{
		"singermid": singerMid,
	}
	return &ArtistRequest{Params: params}
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
		sreq.WithQuery(a.Params),
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

func NewAlbumRequest(albumMid string) *AlbumRequest {
	params := sreq.Params{
		"albummid": albumMid,
	}
	return &AlbumRequest{Params: params}
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
		sreq.WithQuery(a.Params),
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
	params := sreq.Params{
		"id": id,
	}
	return &PlaylistRequest{Params: params}
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
		sreq.WithQuery(p.Params),
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
	res := make([]*provider.MP3, 0, len(p.Response.Data.CDList))
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
