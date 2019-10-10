package migu

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/provider"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/sreq"
)

const (
	GetSongURL          = "https://app.c.nf.migu.cn/MIGUM2.0/v2.0/content/listen-url?netType=01&toneFlag=HQ"
	GetSongId           = "http://music.migu.cn/v3/api/music/audioPlayer/songs?type=1"
	GetSong             = "https://app.c.nf.migu.cn/MIGUM2.0/v2.0/content/querySongBySongId.do?contentId=0"
	GetArtistResource   = "https://app.c.nf.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do?needSimple=01&resourceType=2002"
	GetAlbumResource    = "https://app.c.nf.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do?needSimple=01&resourceType=2003"
	GetPlaylistResource = "https://app.c.nf.migu.cn/MIGUM2.0/v1.0/content/resourceinfo.do?needSimple=01&resourceType=2021"
	GetArtistSongs      = "https://app.c.nf.migu.cn/MIGUM3.0/v1.0/template/singerSongs/release?pageNo=1&pageSize=50&templateVersion=2"
)

type (
	SongURLResponse struct {
		Code string `json:"code"`
		Info string `json:"info"`
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	SongURLRequest struct {
		Params   sreq.Params
		Response SongURLResponse
	}

	SongResponse struct {
		Code     string  `json:"code"`
		Info     string  `json:"info"`
		Resource []*Song `json:"resource"`
	}

	SongRequest struct {
		Params   sreq.Params
		Response SongResponse
	}

	ArtistResponse struct {
		Code string `json:"code"`
		Info string `json:"info"`
		Data struct {
			ContentItemList []struct {
				ItemList []struct {
					Song *Song `json:"song"`
				} `json:"itemList"`
			} `json:"contentItemList"`
		} `json:"data"`
	}

	ArtistRequest struct {
		SingerId string
		Singer   string
		Params   sreq.Params
		Response ArtistResponse
	}

	AlbumResponse struct {
		Code     string  `json:"code"`
		Info     string  `json:"info"`
		Resource []Album `json:"resource"`
	}

	AlbumRequest struct {
		Params   sreq.Params
		Response AlbumResponse
	}

	PlaylistResponse struct {
		Code     string     `json:"code"`
		Info     string     `json:"info"`
		Resource []Playlist `json:"resource"`
	}

	PlaylistRequest struct {
		Params   sreq.Params
		Response PlaylistResponse
	}
)

func NewSongURLRequest(albumId, contentId, copyrightId, resourceType string) *SongURLRequest {
	params := sreq.Params{
		"albumId":               albumId,
		"contentId":             contentId,
		"copyrightId":           copyrightId,
		"lowerQualityContentId": contentId,
		"resourceType":          resourceType,
	}
	return &SongURLRequest{Params: params}
}

func (s *SongURLRequest) Do() error {
	easylog.Debug("SongURLRequest: send GetSongURL api request")
	err := request(GetSongURL,
		sreq.WithQuery(s.Params),
		sreq.WithHeaders(sreq.Headers{
			"channel": "0146832",
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongURLRequest: GetSongURL api request error: %w", err)
	}

	if s.Response.Code != "000000" {
		return fmt.Errorf("SongURLRequest: GetSongURL api status error: %s: %s",
			s.Response.Code, s.Response.Info)
	}

	return nil
}

func NewSongRequest(copyrightId string) *SongRequest {
	params := sreq.Params{
		"copyrightId": copyrightId,
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
	var data struct {
		ReturnCode string `json:"returnCode"`
		Msg        string `json:"msg"`
		Items      []struct {
			SongId string `json:"songId"`
		} `json:"items"`
	}

	easylog.Debug("SongRequest: send GetSongId api request")
	err := request(GetSongId,
		sreq.WithQuery(s.Params),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "http://music.migu.cn",
			"Referer": "http://music.migu.cn",
		}),
	).JSON(&data)
	if err != nil {
		return fmt.Errorf("SongRequest: GetSongId api request error: %w", err)
	}

	if data.ReturnCode != "000000" {
		return fmt.Errorf("SongRequest: GetSongId api status error: %s: %s",
			data.ReturnCode, data.Msg)
	}

	if len(data.Items) == 0 || data.Items[0].SongId == "" {
		return errors.New("SongRequest: empty song id")
	}

	easylog.Debug("SongRequest: send GetSong api request")
	err = request(GetSong,
		sreq.WithQuery(sreq.Params{
			"songId": data.Items[0].SongId,
		}),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&s.Response)
	if err != nil {
		return fmt.Errorf("SongRequest: GetSong api request error: %w", err)
	}

	if s.Response.Code != "000000" {
		return fmt.Errorf("SongRequest: GetSong api status error: %s: %s",
			s.Response.Code, s.Response.Info)
	}

	return nil
}

func (s *SongRequest) Prepare() ([]*provider.MP3, error) {
	return prepare(s.Response.Resource, ".")
}

func NewArtistRequest(singerId string) *ArtistRequest {
	params := sreq.Params{
		"singerId": singerId,
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
		Code     string   `json:"code"`
		Info     string   `json:"info"`
		Resource []Artist `json:"resource"`
	}

	easylog.Debug("ArtistRequest: send GetArtistResource api request")
	err := request(GetArtistResource,
		sreq.WithQuery(sreq.Params{
			"resourceId": a.SingerId,
		}),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&data)
	if err != nil {
		return err
	}

	if data.Code != "000000" {
		return fmt.Errorf("ArtistRequest: GetArtistResource api status error: %s: %s",
			data.Code, data.Info)
	}
	if len(data.Resource) == 0 {
		return errors.New("ArtistRequest: empty artist resource")
	}

	a.Singer = data.Resource[0].Singer

	easylog.Debug("ArtistRequest: send GetArtistSongs api request")
	err = request(GetArtistSongs,
		sreq.WithQuery(a.Params),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api request error: %w", err)
	}

	if a.Response.Code != "000000" {
		return fmt.Errorf("ArtistRequest: GetArtistSongs api status error: %s: %s",
			a.Response.Code, a.Response.Info)
	}

	contentItemList := a.Response.Data.ContentItemList
	if len(contentItemList) == 0 || len(contentItemList[0].ItemList) == 0 {
		return errors.New("ArtistRequest: empty artist data")
	}

	return nil
}

func (a *ArtistRequest) Prepare() ([]*provider.MP3, error) {
	itemList := a.Response.Data.ContentItemList[0].ItemList
	n := len(itemList)
	songs := make([]*Song, 0, n/2)
	for i := 0; i < n; i += 2 {
		songs = append(songs, itemList[i].Song)
	}

	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Singer))
	return prepare(songs, savePath)
}

func NewAlbumRequest(albumId string) *AlbumRequest {
	params := sreq.Params{
		"resourceId": albumId,
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
	easylog.Debug("AlbumRequest: send GetAlbumResource api request")
	err := request(GetAlbumResource,
		sreq.WithQuery(a.Params),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&a.Response)
	if err != nil {
		return fmt.Errorf("AlbumRequest: GetAlbumResource api request error: %w", err)
	}

	if a.Response.Code != "000000" {
		return fmt.Errorf("AlbumRequest: GetAlbumResource api status error: %s: %s",
			a.Response.Code, a.Response.Info)
	}

	if len(a.Response.Resource) == 0 {
		return errors.New("AlbumRequest: empty album data")
	}

	return nil
}

func (a *AlbumRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(a.Response.Resource[0].Title))
	return prepare(a.Response.Resource[0].SongItems, savePath)
}

func NewPlaylistRequest(playlistId string) *PlaylistRequest {
	params := sreq.Params{
		"resourceId": playlistId,
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
	easylog.Debug("PlaylistRequest: send GetPlaylistResource api request")
	err := request(GetPlaylistResource,
		sreq.WithQuery(p.Params),
		sreq.WithHeaders(sreq.Headers{
			"Origin":  "https://app.c.nf.migu.cn",
			"Referer": "https://app.c.nf.migu.cn",
		}),
	).JSON(&p.Response)

	if err != nil {
		return fmt.Errorf("PlaylistRequest: GetPlaylistResource api request error: %w", err)
	}

	if p.Response.Code != "000000" {
		return fmt.Errorf("PlaylistRequest: GetPlaylistResource api status error: %s: %s",
			p.Response.Code, p.Response.Info)
	}

	if len(p.Response.Resource) == 0 {
		return errors.New("PlaylistRequest: empty playlist data")
	}

	return nil
}

func (p *PlaylistRequest) Prepare() ([]*provider.MP3, error) {
	savePath := filepath.Join(".", utils.TrimInvalidFilePathChars(p.Response.Resource[0].Title))
	return prepare(p.Response.Resource[0].SongItems, savePath)
}

func request(url string, opts ...sreq.RequestOption) *sreq.Response {
	return provider.Client(provider.MiguMusic).
		Get(url, opts...).
		EnsureStatusOk()
}
