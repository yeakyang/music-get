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
	LoginAPI    = WeAPI + "/login/cellphone"
	SongURLAPI  = WeAPI + "/song/enhance/player/url"
	SongAPI     = WeAPI + "/v3/song/detail"
	ArtistAPI   = WeAPI + "/v1/artist"
	AlbumAPI    = WeAPI + "/v1/album"
	PlaylistAPI = WeAPI + "/v3/playlist/detail"
)

const (
	BatchSongsCount = 1000
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
	easylog.Debug("Send song url api request")
	resp, err := request(SongURLAPI, s.Params)
	if err != nil {
		return fmt.Errorf("song url api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return fmt.Errorf("song url api response error: %w", err)
	}
	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("song url api status error: %d", s.Response.Code)
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
	easylog.Debug("Send song api request")
	resp, err := request(SongAPI, s.Params)
	if err != nil {
		return fmt.Errorf("song api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return fmt.Errorf("song api response error: %w", err)
	}
	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("song api response status got: %d", s.Response.Code)
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
	easylog.Debugf("Send artist api request: %d", a.Id)
	resp, err := request(ArtistAPI+"/"+strconv.Itoa(a.Id), a.Params)
	if err != nil {
		return fmt.Errorf("artist api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return fmt.Errorf("artist api response error: %w", err)
	}
	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("artist api status error: %d", a.Response.Code)
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
	easylog.Debugf("Send album api request: %d", a.Id)
	resp, err := request(AlbumAPI+"/"+strconv.Itoa(a.Id), a.Params)
	if err != nil {
		return fmt.Errorf("album api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return fmt.Errorf("album api response error: %w", err)
	}
	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("album api status error: %d", a.Response.Code)
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
	easylog.Debugf("Send playlist api request: %d", p.Params.Id)
	resp, err := request(PlaylistAPI, p.Params)
	if err != nil {
		return fmt.Errorf("playlist api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&p.Response); err != nil {
		return fmt.Errorf("playlist api response error: %w", err)
	}
	if p.Response.Code != http.StatusOK {
		return fmt.Errorf("playlist api status error: %d", p.Response.Code)
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
	easylog.Debug("Send login api request")
	resp, err := request(LoginAPI, l.Params)
	if err != nil {
		return fmt.Errorf("login api request error: %w", err)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&l.Response); err != nil {
		return fmt.Errorf("login api response error: %w", err)
	}
	if l.Response.Code != http.StatusOK {
		return fmt.Errorf("login api status error: %d", l.Response.Code)
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

	return provider.Client(provider.NetEaseMusic).
		Post(url,
			sreq.WithForm(sreq.Value{"params": params, "encSecKey": encSecKey}),
		).
		Resolve()
}
