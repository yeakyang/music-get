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
	MaxConcurrentDownloadTasksCount = 16
	DefaultDownloadBr               = 128
)

var (
	confPath                     string
	Conf                         = &Config{}
	downloadOverwrite            bool
	concurrentDownloadTasksCount int
	Debug                        bool
)

type Config struct {
	Cookies                      []*http.Cookie `json:"cookies,omitempty"`
	Workspace                    string         `json:"-"`
	DownloadDir                  string         `json:"-"`
	DownloadOverwrite            bool           `json:"-"`
	ConcurrentDownloadTasksCount int            `json:"-"`
}

func init() {
	flag.BoolVar(&Debug, "v", false, "debug mode")
	flag.BoolVar(&downloadOverwrite, "f", false, "overwrite already downloaded music")
	flag.IntVar(&concurrentDownloadTasksCount, "n", 1, "concurrent download tasks count, max 16")
}

func Init() error {
	if Debug {
		easylog.SetLevel(easylog.Ldebug)
	}
	if concurrentDownloadTasksCount < 1 || concurrentDownloadTasksCount > MaxConcurrentDownloadTasksCount {
		easylog.Warn("Invalid n parameter, use default value")
		concurrentDownloadTasksCount = 1
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	confPath = filepath.Join(pwd, "music-get.json")
	if err = load(confPath); err != nil {
		easylog.Warn("Load config file failed, you may run for the first time")
	}

	downloadDir := filepath.Join(pwd, "downloads")
	if err = utils.BuildPathIfNotExist(downloadDir); err != nil {
		return err
	}

	Conf.Workspace = pwd
	Conf.DownloadDir = downloadDir
	Conf.DownloadOverwrite = downloadOverwrite
	Conf.ConcurrentDownloadTasksCount = concurrentDownloadTasksCount
	return nil
}

func load(confPath string) error {
	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &Conf)
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(confPath, data, 0644)
}
