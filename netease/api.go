package netease

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/winterssy/music-get/common"
	"github.com/winterssy/music-get/config"
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
	br := config.MP3DownloadBr
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
	resp, err := post(SongURLAPI, s.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return err
	}
	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), s.Response.Code, s.Response.Msg)
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
	resp, err := post(SongAPI, s.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s.Response); err != nil {
		return err
	}
	if s.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), s.Response.Code, s.Response.Msg)
	}

	return nil
}

func (s *SongRequest) Extract() ([]*common.MP3, error) {
	return ExtractMP3List(s.Response.Songs, ".")
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
	resp, err := post(ArtistAPI+fmt.Sprintf("/%d", a.Id), a.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return err
	}
	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), a.Response.Code, a.Response.Msg)
	}

	return nil
}

func (a *ArtistRequest) Extract() ([]*common.MP3, error) {
	ids := make([]int, 0, len(a.Response.HotSongs))
	for _, i := range a.Response.HotSongs {
		ids = append(ids, i.Id)
	}

	req := NewSongRequest(ids...)
	if err := req.Do(); err != nil {
		return nil, err
	}

	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Artist.Name))
	return ExtractMP3List(req.Response.Songs, savePath)
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
	resp, err := post(AlbumAPI+fmt.Sprintf("/%d", a.Id), a.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a.Response); err != nil {
		return err
	}
	if a.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), a.Response.Code, a.Response.Msg)
	}

	return nil
}

func (a *AlbumRequest) Extract() ([]*common.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Album.Name))
	for i := range a.Response.Songs {
		a.Response.Songs[i].PublishTime = a.Response.Album.PublishTime
	}
	return ExtractMP3List(a.Response.Songs, savePath)
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
	resp, err := post(PlaylistAPI, p.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&p.Response); err != nil {
		return err
	}
	if p.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), p.Response.Code, p.Response.Msg)
	}

	return nil
}

func (p *PlaylistRequest) Extract() ([]*common.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.Response.Playlist.Name))
	ids, mp3List := make([]int, 0), make([]*common.MP3, 0, len(p.Response.Playlist.TrackIds))

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

			batch, err := ExtractMP3List(req.Response.Songs, savePath)
			if err != nil {
				return nil, err
			}
			mp3List = append(mp3List, batch...)
		}
		ids = append(ids, i.Id)
	}

	if len(ids) > 0 {
		req := NewSongRequest(ids...)
		if err := req.Do(); err != nil {
			return nil, err
		}
		batch, err := ExtractMP3List(req.Response.Songs, savePath)
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
	resp, err := post(LoginAPI, l.Params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&l.Response); err != nil {
		return err
	}
	if l.Response.Code != http.StatusOK {
		return fmt.Errorf("%s %s error: %d %s", resp.Request.Method, resp.Request.URL.String(), l.Response.Code, l.Response.Msg)
	}

	config.M.Cookies = resp.Cookies()
	return nil
}
