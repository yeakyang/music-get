package main

import (
	"flag"

	"github.com/winterssy/music-get/conf"
	"github.com/winterssy/music-get/handler"
	"github.com/winterssy/music-get/pkg/easylog"
	"github.com/winterssy/music-get/utils"
)

func main() {
	if len(flag.Args()) == 0 {
		easylog.Fatal("Missing music address")
	}

	if err := utils.BuildPathIfNotExist(conf.MP3DownloadDir); err != nil {
		easylog.Fatalf("Failed to build path: %s: %s", conf.MP3DownloadDir, err)
	}

	url := flag.Args()[0]
	req, err := handler.Parse(url)
	if err != nil {
		easylog.Fatal(err)
	}

	if req.RequireLogin() {
		easylog.Info("Local cached cookies expired, please login to refresh...")
		if err = req.Login(); err != nil {
			easylog.Errorf("Login failed: %s", err.Error())
		}
		easylog.Info("Login successful")
	}

	if err := conf.M.Save(); err != nil {
		easylog.Warnf("Save config error: %s", err.Error())
	}

	if err = req.Do(); err != nil {
		easylog.Fatal(err)
	}

	mp3List, err := req.Extract()
	if err != nil {
		easylog.Fatal(err)
	}

	if len(mp3List) == 0 {
		return
	}

	n := conf.MP3ConcurrentDownloadTasksNumber
	if n > conf.MaxConcurrentDownloadTasksNumber {
		n = conf.MaxConcurrentDownloadTasksNumber
	}
	switch {
	case n > 1:
		handler.ConcurrentDownload(mp3List, n)
	default:
		handler.SingleDownload(mp3List)
	}
}
