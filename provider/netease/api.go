package netease

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/conf"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/sreq"
)

const (
	WeAPI       = "https://music.163.com/weapi"
	Login       = WeAPI + "/login/cellphone"
	GetSongURL  = WeAPI + "/song/enhance/player/url"
	GetSong     = WeAPI + "/v3/song/detail"
	GetArtist   = WeAPI + "/v1/artist"
	GetAlbum    = WeAPI + "/v1/album"
	GetPlaylist = WeAPI + "/v3/playlist/detail"

	BatchSongsCount = 1000
)

type (
	SongURLParams struct {
		Ids string `json:"ids"`
		Br  int    `json:"br"`
	}

	SongURLResponse struct {
		Code int       `json:"code"`
		Msg  string    `json:"msg"`
		Data []SongURL `json:"data"`
	}

	SongURLRequest struct {
		Params   SongURLParams
		Response SongURLResponse
	}

	SongParams struct {
		C string `json:"c"`
	}

	SongResponse struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Songs []Song `json:"songs"`
	}

	SongRequest struct {
		Params   SongParams
		Response SongResponse
	}

	ArtistParams struct{}

	ArtistResponse struct {
		Code     int    `json:"code"`
		Msg      string `json:"msg"`
		Artist   Artist `json:"artist"`
		HotSongs []Song `json:"hotSongs"`
	}

	ArtistRequest struct {
		Id       int
		Params   ArtistParams
		Response ArtistResponse
	}

	AlbumParams struct{}

	AlbumResponse struct {
		SongResponse
		Album Album `json:"album"`
	}

	AlbumRequest struct {
		Id       int
		Params   AlbumParams
		Response AlbumResponse
	}

	PlaylistParams struct {
		Id int `json:"id"`
	}

	PlaylistResponse struct {
		Code     int      `json:"code"`
		Msg      string   `json:"msg"`
		Playlist Playlist `json:"playlist"`
	}

	PlaylistRequest struct {
		Params   PlaylistParams
		Response PlaylistResponse
	}

	LoginParams struct {
		Phone         string `json:"phone"`
		Password      string `json:"password"`
		RememberLogin bool   `json:"rememberLogin"`
	}

	LoginResponse struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		LoginType int    `json:"loginType"`
	}

	LoginRequest struct {
		Params   LoginParams
		Response LoginResponse
	}
)

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
	easylog.Debug("SongURLRequest: send GetSongURL api request")
	err := request(GetSongURL, s.Params).
		JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongURLRequest: GetSongURL api request error: %w", err)
	}

	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %d: %s",
			s.Response.Code, s.Response.Msg)
	}

	return nil
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
	easylog.Debug("SongRequest: send GetSong api request")
	err := request(GetSong, s.Params).
		JSON(&s.Response)
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
	return prepare(s.Response.Songs, ".")
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
	easylog.Debugf("ArtistRequest: send GetArtist api request: %d", a.Id)
	err := request(GetArtist+"/"+strconv.Itoa(a.Id), a.Params).
		JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("ArtistRequest: GetArtist api request error: %w", err)
	}

	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("ArtistRequest: GetArtist api status error: %d: %s",
			a.Response.Code, a.Response.Msg)
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
	easylog.Debugf("AlbumRequest: send GetAlbum api request: %d", a.Id)
	err := request(GetAlbum+"/"+strconv.Itoa(a.Id), a.Params).
		JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("AlbumRequest: GetAlbum api request error: %w", err)
	}

	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("AlbumRequest: GetAlbum api status error: %d: %s",
			a.Response.Code, a.Response.Msg)
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
	easylog.Debugf("PlaylistRequest: send GetPlaylist api request: %d", p.Params.Id)
	err := request(GetPlaylist, p.Params).
		JSON(&p.Response)
	if err != nil {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api request error: %w", err)
	}

	if p.Response.Code != http.StatusOK {
		return fmt.Errorf("PlaylistRequest: GetPlaylist api status error: %d: %s",
			p.Response.Code, p.Response.Msg)
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.Response.Playlist.Name))
	n := len(p.Response.Playlist.TrackIds)
	mp3List := make([]*provider.MP3, 0, n)

	for i := 0; i < n; i += BatchSongsCount {
		j := i + BatchSongsCount
		if j > n {
			j = n
		}

		ids := make([]int, 0, j-i)
		for k := i; k < j; k++ {
			ids = append(ids, p.Response.Playlist.TrackIds[k].Id)
		}

		req := NewSongRequest(ids...)
		if err := req.Do(); err != nil {
			return nil, err
		}

		batch, err := prepare(req.Response.Songs, savePath)
		if err != nil {
			return nil, err
		}
		mp3List = append(mp3List, batch...)
	}

	return mp3List, nil
}

func NewLoginRequest(phone, password string) *LoginRequest {
	passwordHash := md5.Sum([]byte(password))
	password = hex.EncodeToString(passwordHash[:])
	return &LoginRequest{Params: LoginParams{Phone: phone, Password: password, RememberLogin: true}}
}

func (l *LoginRequest) Do() error {
	easylog.Debug("LoginRequest: send Login api request")
	resp := request(Login, l.Params)
	if err := resp.JSON(&l.Response); err != nil {
		return fmt.Errorf("LoginRequest: Login api request error: %w", err)
	}

	if l.Response.Code != http.StatusOK {
		return fmt.Errorf("LoginRequest: Login api status error: %d: %s",
			l.Response.Code, l.Response.Msg)
	}

	conf.Conf.Cookies = resp.R.Cookies()
	return nil
}

func request(url string, data interface{}) *sreq.Response {
	enc, _ := json.Marshal(data)
	params, encSecKey, err := Encrypt(enc)
	if err != nil {
		return &sreq.Response{
			Err: err,
		}
	}

	return provider.Client(provider.NetEaseMusic).
		Post(url,
			sreq.WithForm(sreq.Value{"params": params, "encSecKey": encSecKey}),
		).EnsureStatusOk()
}
