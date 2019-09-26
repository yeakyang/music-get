package qq

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/winterssy/music-get/pkg/ecode"
	"github.com/winterssy/music-get/pkg/requests"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

const (
	SongURLAPI  = "https://u.y.qq.com/cgi-bin/musicu.fcg"
	SongAPI     = "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg"
	ArtistAPI   = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_singer_track_cp.fcg"
	AlbumAPI    = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_detail_cp.fcg"
	PlaylistAPI = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_playlist_cp.fcg"
)

type SongURLResponse struct {
	Code int `json:"code"`
	Req0 struct {
		Data struct {
			MidURLInfo []struct {
				SongMid string `json:"songmid"`
				Vkey    string `json:"vkey"`
			} `json:"midurlinfo"`
			TestFile2g string `json:"testfile2g"`
		} `json:"data"`
	} `json:"req0"`
}

type SongURLRequest struct {
	Params   requests.Values
	Response SongURLResponse
}

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
	query := requests.Values{
		"data": string(enc),
	}

	return &SongURLRequest{Params: query}
}

func (s *SongURLRequest) Do() error {
	resp, err := request(SongURLAPI, s.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "qq.SongURLRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "qq.SongURLRequest.Do:json.Unmarshal")
	}
	if s.Response.Code != 0 {
		return ecode.NewError(ecode.APIResponseException, "qq.SongURLRequest.Do")
	}

	return nil
}

type SongResponse struct {
	Code int    `json:"code"`
	Data []Song `json:"data"`
}

type SongRequest struct {
	Params   requests.Values
	Response SongResponse
}

func NewSongRequest(mid string) *SongRequest {
	query := requests.Values{
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
	resp, err := request(SongAPI, s.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "qq.SongRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "qq.SongRequest.Do:json.Unmarshal")
	}
	if s.Response.Code != 0 {
		return ecode.NewError(ecode.APIResponseException, "qq.SongRequest.Do")
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	return prepare(s.Response.Data, ".")
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
	Params   requests.Values
	Response SingerResponse
}

func NewSingerRequest(mid string) *SingerRequest {
	query := requests.Values{
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
	resp, err := request(ArtistAPI, s.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "qq.SingerRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "qq.SingerRequest.Do:json.Unmarshal")
	}
	if s.Response.Code != 0 {
		return ecode.NewError(ecode.APIResponseException, "qq.SingerRequest.Do")
	}

	return nil
}

func (s *SingerRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(s.Response.Data.SingerName))
	var songs []Song
	for _, i := range s.Response.Data.List {
		songs = append(songs, i.MusicData)
	}
	return prepare(songs, savePath)
}

type AlbumResponse struct {
	Code int
	Data struct {
		GetAlbumInfo GetAlbumInfo `json:"getAlbumInfo"`
		GetSongInfo  []Song       `json:"getSongInfo"`
	} `json:"data"`
}

type AlbumRequest struct {
	Params   requests.Values
	Response AlbumResponse
}

func NewAlbumRequest(mid string) *AlbumRequest {
	query := requests.Values{
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
	resp, err := request(AlbumAPI, a.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "qq.AlbumRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "qq.AlbumRequest.Do:json.Unmarshal")
	}
	if a.Response.Code != 0 {
		return ecode.NewError(ecode.APIResponseException, "qq.AlbumRequest.Do")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Data.GetAlbumInfo.FAlbumName))
	return prepare(a.Response.Data.GetSongInfo, savePath)
}

type PlaylistResponse struct {
	Code int `json:"code"`
	Data struct {
		CDList []CD `json:"cdlist"`
	} `json:"data"`
}

type PlaylistRequest struct {
	Params   requests.Values
	Response PlaylistResponse
}

func NewPlaylistRequest(id string) *PlaylistRequest {
	query := requests.Values{
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
	resp, err := request(PlaylistAPI, p.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "qq.PlaylistRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&p.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "qq.PlaylistRequest.Do:json.Unmarshal")
	}
	if p.Response.Code != 0 {
		return ecode.NewError(ecode.APIResponseException, "qq.PlaylistRequest.Do")
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

func request(url string, params requests.Values) (*http.Response, error) {
	return provider.GetRequest().Get(url).
		Params(params).
		Headers(provider.RequestHeader[provider.QQMusic]).
		Send().
		Resolve()
}
