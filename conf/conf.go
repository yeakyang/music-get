package conf

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/utils"
)

const (
	MaxConcurrentDownloadTasksNumber = 16
)

type MusicGet struct {
	Cookies []*http.Cookie `json:"cookies,omitempty"`
}

var (
	M                                MusicGet
	Workspace                        string
	MP3DownloadDir                   string
	MP3DownloadBr                    int
	DownloadOverwrite                bool
	MP3ConcurrentDownloadTasksNumber int
)

func Init() {
	Workspace, err := os.Getwd()
	if err != nil {
		easylog.Fatal(err)
	}

	configFile := filepath.Join(Workspace, "music-get.json")
	if err = Load(configFile); err != nil {
		easylog.Warnf("config file not found, will be created later...")
	}

	MP3DownloadDir = filepath.Join(Workspace, "downloads")
	if err = utils.BuildPathIfNotExist(MP3DownloadDir); err != nil {
		easylog.Fatalf("unable to create download dir: %s", err.Error())
	}
	flag.IntVar(&MP3DownloadBr, "br", 128, "MP3 prior download bit rate, 128|192|320, removed")
	flag.BoolVar(&DownloadOverwrite, "f", false, "overwrite already downloaded music")
	flag.IntVar(&MP3ConcurrentDownloadTasksNumber, "n", 1, "MP3 concurrent download tasks number, max 16")
	flag.Parse()
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

	file := filepath.Join(Workspace, "music-get.json")
	return ioutil.WriteFile(file, data, 0644)
}
