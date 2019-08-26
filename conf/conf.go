package conf

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/winterssy/music-get/pkg/easylog"
)

const (
	DefaultMP3DownloadBr             = 128
	MaxConcurrentDownloadTasksNumber = 16
)

type MusicGet struct {
	Cookies []*http.Cookie `json:"cookies,omitempty"`
	Br      int            `json:"br"`
}

var (
	M                                MusicGet
	MP3DownloadBr                    int
	MP3DownloadDir                   string
	DownloadOverwrite                bool
	MP3ConcurrentDownloadTasksNumber int
)

func init() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		easylog.Fatal(err)
	}

	file := filepath.Join(homedir, "music-get.json")
	if err = Load(file); err != nil {
		M.Br = DefaultMP3DownloadBr
	}

	downloadDir := filepath.Join(homedir, "Music-Get")
	flag.StringVar(&MP3DownloadDir, "o", downloadDir, "MP3 download directory")
	flag.IntVar(&MP3DownloadBr, "br", M.Br, "MP3 prior download bit rate, 128|192|320")
	flag.BoolVar(&DownloadOverwrite, "f", false, "overwrite already downloaded music")
	flag.IntVar(&MP3ConcurrentDownloadTasksNumber, "n", 1, "MP3 concurrent download tasks number, max 16")
	flag.Parse()
	M.Br = MP3DownloadBr
}

func Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, &M); err != nil {
		return err
	}

	return nil
}

func (m *MusicGet) Save() error {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	file := filepath.Join(homedir, "music-get.json")
	return ioutil.WriteFile(file, data, 0644)
}
