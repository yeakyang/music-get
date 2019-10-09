package kugou

import (
	"crypto/md5"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/sreq"
)

const (
	GetSongURL       = "http://trackercdn.kugou.com/i/v2/?pid=2&behavior=play&cmd=25"
	GetSong          = "http://m.kugou.com/api/v1/song/get_song_info?cmd=playInfo"
	GetArtistInfo    = "http://mobilecdn.kugou.com/api/v3/singer/info"
	GetArtistSongs   = "http://mobilecdn.kugou.com/api/v3/singer/song?page=1&pagesize=50"
	GetAlbumInfo     = "http://mobilecdn.kugou.com/api/v3/album/info"
	GetAlbumSongs    = "http://mobilecdn.kugou.com/api/v3/album/song?page=1&pagesize=-1"
	GetPlaylistInfo  = "http://mobilecdn.kugou.com/api/v3/special/info"
	GetPlaylistSongs = "http://mobilecdn.kugou.com/api/v3/special/song?page=1&pagesize=-1"
)

type (
	SongURLResponse struct {
		BitRate int      `json:"bitRate"`
		ExtName string   `json:"extName"`
		URL     []string `json:"url"`
		Status  int      `json:"status"`
	}

	SongURLRequest struct {
		Params   sreq.Value
		Response SongURLResponse
	}

	SongResponse struct {
		SongName   string `json:"songName"`
		SingerName string `json:"singerName"`
		FileName   string `json:"fileName"`
		ExtName    string `json:"extName"`
		Hash       string `json:"hash"`
		Extra      struct {
			SQHash string `json:"sqhash"`
			PQHash string `json:"128hash"`
			HQHash string `json:"320hash"`
		}
		URL    string `json:"url"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	SongRequest struct {
		Params   sreq.Value
		Response SongResponse
	}

	ArtistResponse struct {
		Data struct {
			Info []*Song `json:"info"`
		} `json:"data"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	ArtistRequest struct {
		SingerId   string
		SingerName string
		Params     sreq.Value
		Response   ArtistResponse
	}

	AlbumResponse struct {
		Data struct {
			Info []*Song `json:"info"`
		} `json:"data"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	AlbumRequest struct {
		AlbumId   string
		AlbumName string
		Params    sreq.Value
		Response  AlbumResponse
	}

	PlaylistResponse struct {
		Data struct {
			Info []*Song `json:"info"`
		} `json:"data"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	PlaylistRequest struct {
		SpecialId   string
		SpecialName string
		Params      sreq.Value
		Response    PlaylistResponse
	}
)

func NewSongURLRequest(hash string) *SongURLRequest {
	data := []byte(hash + "kgcloudv2")
	key := fmt.Sprintf("%x", md5.Sum(data))
	params := sreq.Value{
		"hash": hash,
		"key":  key,
	}
	return &SongURLRequest{Params: params}
}

func (s *SongURLRequest) Do() error {
	easylog.Debug("SongURLRequest: send GetSongURL api request")
	err := request(GetSongURL,
		sreq.WithParams(s.Params),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://trackercdn.kugou.com",
			"Referer": "http://trackercdn.kugou.com",
		}),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongURLRequest: GetSongURL api request error: %w", err)
	}

	if s.Response.Status != 1 {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %d", s.Response.Status)
	}

	if len(s.Response.URL) == 0 {
		return errors.New("SongURLRequest: song url unavailable")
	}

	return nil
}

func NewSongRequest(hash string) *SongRequest {
	params := sreq.Value{
		"hash": hash,
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
		sreq.WithParams(s.Params),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://m.kugou.com",
			"Referer": "http://m.kugou.com",
		}),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongRequest: GetSong api request error: %w", err)
	}

	if s.Response.Status != 1 {
		return fmt.Errorf("SongRequest: GetSong api status error: %d: %s",
			s.Response.Status, s.Response.Error)
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	songs := []*Song{
		{
			FileName: s.Response.FileName,
			ExtName:  s.Response.ExtName,
			Hash:     s.Response.Hash,
		},
	}
	return prepare(songs, ".")
}

func NewArtistRequest(singerId string) *ArtistRequest {
	params := sreq.Value{
		"singerid": singerId,
	}
	return &ArtistRequest{
		SingerId: singerId,
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
		Data   Artist `json:"data"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	easylog.Debug("ArtistRequest: send GetArtistInfo api request")
	err := request(GetArtistInfo,
		sreq.WithParams(sreq.Value{
			"singerid": a.SingerId,
		}),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://mobilecdn.kugou.com",
			"Referer": "http://mobilecdn.kugou.com",
		}),
	).JSON(&data)
	if err != nil {
		return err
	}

	if data.Status != 1 {
		return fmt.Errorf("ArtistRequest: GetArtistInfo api status error: %d: %s",
			data.Status, data.Error)
	}

	a.SingerName = data.Data.SingerName

	easylog.Debug("ArtistRequest: send GetArtistSongs api request")
	err = request(GetArtistSongs,
		sreq.WithParams(a.Params),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://mobilecdn.kugou.com",
			"Referer": "http://mobilecdn.kugou.com",
		}),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api request error: %w", err)
	}

	if a.Response.Status != 1 {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api status error: %d: %s",
			a.Response.Status, a.Response.Error)
	}

	if len(a.Response.Data.Info) == 0 {
		return errors.New("ArtistRequest: empty artist data")
	}

	return nil
}

func (a *ArtistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.SingerName))
	return prepare(a.Response.Data.Info, savePath)
}

func NewAlbumRequest(albumId string) *AlbumRequest {
	params := sreq.Value{
		"albumid": albumId,
	}
	return &AlbumRequest{
		AlbumId: albumId,
		Params:  params,
	}
}

func (a *AlbumRequest) RequireLogin() bool {
	return false
}

func (a *AlbumRequest) Login() error {
	panic("implement me")
}

func (a *AlbumRequest) Do() error {
	var data struct {
		Data   Album  `json:"data"`
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	easylog.Debug("AlbumRequest: send GetAlbumInfo api request")
	err := request(GetAlbumInfo,
		sreq.WithParams(sreq.Value{
			"albumid": a.AlbumId,
		}),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://mobilecdn.kugou.com",
			"Referer": "http://mobilecdn.kugou.com",
		}),
	).JSON(&data)
	if err != nil {
		return err
	}

	if data.Status != 1 {
		return fmt.Errorf("AlbumRequest: GetAlbumInfo api status error: %d: %s",
			data.Status, data.Error)
	}

	a.AlbumName = data.Data.AlbumName

	easylog.Debug("AlbumRequest: send GetAlbumSongs api request")
	err = request(GetAlbumSongs,
		sreq.WithParams(a.Params),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://mobilecdn.kugou.com",
			"Referer": "http://mobilecdn.kugou.com",
		}),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("AlbumRequest: GetAlbumSongs api request error: %w", err)
	}

	if a.Response.Status != 1 {
		return fmt.Errorf("AlbumRequest: GetAlbumSongs api status error: %d: %s",
			a.Response.Status, a.Response.Error)
	}

	if len(a.Response.Data.Info) == 0 {
		return errors.New("AlbumRequest: empty album data")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.AlbumName))
	return prepare(a.Response.Data.Info, savePath)
}

func NewPlaylistRequest(specialId string) *PlaylistRequest {
	params := sreq.Value{
		"specialid": specialId,
	}
	return &PlaylistRequest{
		SpecialId: specialId,
		Params:    params,
	}
}

func (p *PlaylistRequest) RequireLogin() bool {
	return false
}

func (p *PlaylistRequest) Login() error {
	panic("implement me")
}

func (p *PlaylistRequest) Do() error {
	var data struct {
		Data   Playlist `json:"data"`
		Status int      `json:"status"`
		Error  string   `json:"error"`
	}

	easylog.Debug("PlaylistRequest: send GetPlaylistInfo api request")
	err := request(GetPlaylistInfo,
		sreq.WithParams(sreq.Value{
			"specialid": p.SpecialId,
		}),
		sreq.WithHeaders(sreq.Value{
			"Origin":  "http://mobilecdn.kugou.com",
			"Referer": "http://mobilecdn.kugou.com",
		}),
	).JSON(&data)
	if err != nil {
		return err
	}

	if data.Status != 1 {
		return fmt.Errorf("PlaylistRequest: GetPlaylistInfo api status error: %d: %s",
			data.Status, data.Error)
	}

	p.SpecialName = data.Data.SpecialName

	easylog.Debug("PlaylistRequest: send GetPlaylistSongs api request")
	err = request(GetPlaylistSongs,
		sreq.WithParams(p.Params),
	).JSON(&p.Response)
	if err != nil {
		return fmt.Errorf("PlaylistRequest: GetPlaylistSongs api request error: %w", err)
	}

	if p.Response.Status != 1 {
		return fmt.Errorf("PlaylistRequest: GetPlaylistSongs api status error: %d: %s",
			p.Response.Status, p.Response.Error)
	}

	if len(p.Response.Data.Info) == 0 {
		return errors.New("PlaylistRequest: empty playlist data")
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.SpecialName))
	return prepare(p.Response.Data.Info, savePath)
}

func request(url string, opts ...sreq.RequestOption) *sreq.Response {
	return provider.Client(provider.KugouMusic).
		Get(url, opts...).
		EnsureStatusOk()
}
