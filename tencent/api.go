package tencent

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/winterssy/music-get/common"
	"github.com/winterssy/music-get/utils"
)

const (
	SongAPI     = "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg"
	ArtistAPI   = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_singer_track_cp.fcg"
	AlbumAPI    = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_detail_cp.fcg"
	PlaylistAPI = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_playlist_cp.fcg"
)

type SongResponse struct {
	Code int    `json:"code"`
	Data []Song `json:"data"`
}

type SongRequest struct {
	Params   map[string]string
	Response SongResponse
}

func NewSongRequest(mid string) *SongRequest {
	query := map[string]string{
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
	resp, err := common.Request("GET", SongAPI, s.Params, nil, common.TencentMusic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return err
	}
	if s.Response.Code != 0 {
		return fmt.Errorf("%s %s error: %d", resp.Request.Method, resp.Request.URL.String(), s.Response.Code)
	}

	return nil
}

func (s *SongRequest) Extract() ([]*common.MP3, error) {
	return ExtractMP3List(s.Response.Data, ".")
}

type SingerResponse struct {
	Code int `json:"code"`
	Data struct {
		List []struct {
			MusicData Song `json:"musicData"`
		} `json:"list"`
		SingerId   string `json:"singer_id"`
		SingerMid  string `json:"singer_mid"`
		SingerName string `json:"singer_name"`
		Total      int    `json:"total"`
	} `json:"data"`
}

type SingerRequest struct {
	Params   map[string]string
	Response SingerResponse
}

func NewSingerRequest(mid string) *SingerRequest {
	query := map[string]string{
		"singermid": mid,
		"begin":     "0",
		"num":       "50",
		"order":     "listen",
		"newsong":   "1",
		"platform":  "yqq",
	}
	return &SingerRequest{Params: query}
}

func (s *SingerRequest) RequireLogin() bool {
	return false
}

func (s *SingerRequest) Login() error {
	panic("implement me")
}

func (s *SingerRequest) Do() error {
	resp, err := common.Request("GET", ArtistAPI, s.Params, nil, common.TencentMusic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return err
	}
	if s.Response.Code != 0 {
		return fmt.Errorf("%s %s error: %d", resp.Request.Method, resp.Request.URL.String(), s.Response.Code)
	}

	return nil
}

func (s *SingerRequest) Extract() ([]*common.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(s.Response.Data.SingerName))
	var songs []Song
	for _, i := range s.Response.Data.List {
		songs = append(songs, i.MusicData)
	}
	return ExtractMP3List(songs, savePath)
}

type AlbumResponse struct {
	Code int
	Data struct {
		GetAlbumInfo GetAlbumInfo `json:"getAlbumInfo"`
		GetSongInfo  []Song       `json:"getSongInfo"`
	} `json:"data"`
}

type AlbumRequest struct {
	Params   map[string]string
	Response AlbumResponse
}

func NewAlbumRequest(mid string) *AlbumRequest {
	query := map[string]string{
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
	resp, err := common.Request("GET", AlbumAPI, a.Params, nil, common.TencentMusic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return err
	}
	if a.Response.Code != 0 {
		return fmt.Errorf("%s %s error: %d", resp.Request.Method, resp.Request.URL.String(), a.Response.Code)
	}

	return nil
}

func (a *AlbumRequest) Extract() ([]*common.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Data.GetAlbumInfo.FAlbumName))
	return ExtractMP3List(a.Response.Data.GetSongInfo, savePath)
}

type PlaylistResponse struct {
	Code int `json:"code"`
	Data struct {
		CDList []CD `json:"cdlist"`
	} `json:"data"`
}

type PlaylistRequest struct {
	Params   map[string]string
	Response PlaylistResponse
}

func NewPlaylistRequest(id string) *PlaylistRequest {
	query := map[string]string{
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
	resp, err := common.Request("GET", PlaylistAPI, p.Params, nil, common.TencentMusic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&p.Response); err != nil {
		return err
	}
	if p.Response.Code != 0 {
		return fmt.Errorf("%s %s error: %d", resp.Request.Method, resp.Request.URL.String(), p.Response.Code)
	}

	return nil
}

func (p *PlaylistRequest) Extract() ([]*common.MP3, error) {
	var res []*common.MP3
	for _, i := range p.Response.Data.CDList {
		savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(i.DissName))
		mp3List, err := ExtractMP3List(i.SongList, savePath)
		if err != nil {
			continue
		}
		res = append(res, mp3List...)
	}

	return res, nil
}
