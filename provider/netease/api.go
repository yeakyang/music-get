package netease

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/winterssy/music-get/conf"
	"github.com/winterssy/music-get/pkg/ecode"
	"github.com/winterssy/music-get/pkg/requests"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
)

const (
	WeAPI       = "https://music.163.com/weapi"
	LoginAPI    = WeAPI + "/login/cellphone"
	SongURLAPI  = WeAPI + "/song/enhance/player/url"
	SongAPI     = WeAPI + "/v3/song/detail"
	ArtistAPI   = WeAPI + "/v1/artist"
	AlbumAPI    = WeAPI + "/v1/album"
	PlaylistAPI = WeAPI + "/v3/playlist/detail"
)

const (
	MaxSongsCount = 1000
)

type SongURLParams struct {
	Ids string `json:"ids"`
	Br  int    `json:"br"`
}

type SongURLResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data []SongURL `json:"data"`
}

type SongURLRequest struct {
	Params   SongURLParams
	Response SongURLResponse
}

func NewSongURLRequest(ids ...int) *SongURLRequest {
	br := conf.DefaultDownloadBr
	switch br {
	case 128, 192, 320:
		br *= 1000
	default:
		br = 999 * 1000
	}
	enc, _ := json.Marshal(ids)
	return &SongURLRequest{Params: SongURLParams{Ids: string(enc), Br: br}}
}

func (s *SongURLRequest) Do() error {
	resp, err := request(SongURLAPI, s.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.SongURLRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.SongURLRequest.Do:json.Unmarshal")
	}
	if s.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.SongURLRequest.Do")
	}

	return nil
}

type SongParams struct {
	C string `json:"c"`
}

type SongResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Songs []Song `json:"songs"`
}

type SongRequest struct {
	Params   SongParams
	Response SongResponse
}

func NewSongRequest(ids ...int) *SongRequest {
	c := make([]map[string]int, 0, len(ids))
	for _, id := range ids {
		c = append(c, map[string]int{"id": id})
	}

	enc, _ := json.Marshal(c)
	return &SongRequest{Params: SongParams{C: string(enc)}}
}

func (s *SongRequest) RequireLogin() bool {
	return !isAuthenticated()
}

func (s *SongRequest) Login() error {
	return login()
}

func (s *SongRequest) Do() error {
	resp, err := request(SongAPI, s.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.SongRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.SongRequest.Do:json.Unmarshal")
	}
	if s.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.SongRequest.Do")
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	return prepare(s.Response.Songs, ".")
}

type ArtistParams struct{}

type ArtistResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	Artist   Artist `json:"artist"`
	HotSongs []Song `json:"hotSongs"`
}

type ArtistRequest struct {
	Id       int
	Params   ArtistParams
	Response ArtistResponse
}

func NewArtistRequest(id int) *ArtistRequest {
	return &ArtistRequest{Id: id, Params: ArtistParams{}}
}

func (a *ArtistRequest) RequireLogin() bool {
	return !isAuthenticated()
}

func (a *ArtistRequest) Login() error {
	return login()
}

func (a *ArtistRequest) Do() error {
	resp, err := request(ArtistAPI+"/"+strconv.Itoa(a.Id), a.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.ArtistRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.ArtistRequest.Do:json.Unmarshal")
	}
	if a.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.ArtistRequest.Do")
	}

	return nil
}

func (a *ArtistRequest) Prepare() ([]*provider.MP3, error) {
	ids := make([]int, 0, len(a.Response.HotSongs))
	for _, i := range a.Response.HotSongs {
		ids = append(ids, i.Id)
	}

	req := NewSongRequest(ids...)
	if err := req.Do(); err != nil {
		return nil, err
	}

	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Artist.Name))
	return prepare(req.Response.Songs, savePath)
}

type AlbumParams struct{}

type AlbumResponse struct {
	SongResponse
	Album Album `json:"album"`
}

type AlbumRequest struct {
	Id       int
	Params   AlbumParams
	Response AlbumResponse
}

func NewAlbumRequest(id int) *AlbumRequest {
	return &AlbumRequest{Id: id, Params: AlbumParams{}}
}

func (a *AlbumRequest) RequireLogin() bool {
	return !isAuthenticated()
}

func (a *AlbumRequest) Login() error {
	return login()
}

func (a *AlbumRequest) Do() error {
	resp, err := request(AlbumAPI+"/"+strconv.Itoa(a.Id), a.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.AlbumRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.AlbumRequest.Do:json.Unmarshal")
	}
	if a.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.AlbumRequest.Do")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Album.Name))
	for i := range a.Response.Songs {
		a.Response.Songs[i].PublishTime = a.Response.Album.PublishTime
	}
	return prepare(a.Response.Songs, savePath)
}

type PlaylistParams struct {
	Id int `json:"id"`
}

type PlaylistResponse struct {
	Code     int      `json:"code"`
	Msg      string   `json:"msg"`
	Playlist Playlist `json:"playlist"`
}

type PlaylistRequest struct {
	Params   PlaylistParams
	Response PlaylistResponse
}

func NewPlaylistRequest(id int) *PlaylistRequest {
	return &PlaylistRequest{Params: PlaylistParams{Id: id}}
}

func (p *PlaylistRequest) RequireLogin() bool {
	return !isAuthenticated()
}

func (p *PlaylistRequest) Login() error {
	return login()
}

func (p *PlaylistRequest) Do() error {
	resp, err := request(PlaylistAPI, p.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.PlaylistRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&p.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.PlaylistRequest.Do:json.Unmarshal")
	}
	if p.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.PlaylistRequest.Do")
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.Response.Playlist.Name))
	ids, songs := make([]int, 0), make([]*provider.MP3, 0, len(p.Response.Playlist.TrackIds))

	count := 0
	for _, i := range p.Response.Playlist.TrackIds {
		count++
		if count > MaxSongsCount {
			req := NewSongRequest(ids...)
			ids = make([]int, 0)
			count = 0
			if err := req.Do(); err != nil {
				return nil, err
			}

			batch, err := prepare(req.Response.Songs, savePath)
			if err != nil {
				return nil, err
			}
			songs = append(songs, batch...)
		}
		ids = append(ids, i.Id)
	}

	if len(ids) > 0 {
		req := NewSongRequest(ids...)
		if err := req.Do(); err != nil {
			return nil, err
		}
		batch, err := prepare(req.Response.Songs, savePath)
		if err != nil {
			return nil, err
		}
		songs = append(songs, batch...)
	}

	return songs, nil
}

type LoginParams struct {
	Phone         string `json:"phone"`
	Password      string `json:"password"`
	RememberLogin bool   `json:"rememberLogin"`
}

type LoginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	LoginType int    `json:"loginType"`
}

type LoginRequest struct {
	Params   LoginParams
	Response LoginResponse
}

func NewLoginRequest(phone, password string) *LoginRequest {
	passwordHash := md5.Sum([]byte(password))
	password = hex.EncodeToString(passwordHash[:])
	return &LoginRequest{Params: LoginParams{Phone: phone, Password: password, RememberLogin: true}}
}

func (l *LoginRequest) Do() error {
	resp, err := request(LoginAPI, l.Params)
	if err != nil {
		return ecode.NewError(ecode.HTTPRequestException, "netease.LoginRequest.Do")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&l.Response); err != nil {
		return ecode.NewError(ecode.APIResponseException, "netease.LoginRequest.Do:json.Unmarshal")
	}
	if l.Response.Code != http.StatusOK {
		return ecode.NewError(ecode.APIResponseException, "netease.LoginRequest.Do")
	}

	conf.Conf.Cookies = resp.Cookies()
	return nil
}

func request(url string, data interface{}) (*http.Response, error) {
	enc, _ := json.Marshal(data)
	params, encSecKey, err := Encrypt(enc)
	if err != nil {
		return nil, err
	}

	return provider.GetRequest().Post(url).
		Form(requests.Values{"params": params, "encSecKey": encSecKey}).
		Headers(provider.RequestHeader[provider.NetEaseMusic]).
		Send().
		Resolve()
}
