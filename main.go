package main

import (
	"flag"

	"github.com/winterssy/music-get/config"
	"github.com/winterssy/music-get/handler"
	"github.com/winterssy/music-get/utils"
	"github.com/winterssy/music-get/utils/logger"
)

func main() {
	if len(flag.Args()) == 0 {
		logger.Error.Fatal("Missing music address")
	}

	if err := utils.BuildPathIfNotExist(config.MP3DownloadDir); err != nil {
		logger.Error.Fatalf("Failed to build path: %s: %s", config.MP3DownloadDir, err)
	}

	url := flag.Args()[0]
	req, err := handler.Parse(url)
	if err != nil {
		logger.Error.Fatal(err)
	}

	if req.RequireLogin() {
		logger.Info.Print("Local cached cookies expired, please login to refresh...")
		if err = req.Login(); err != nil {
			logger.Error.Fatalf("Login failed: %s", err.Error())
		}
		logger.Info.Print("Login successful")
	}

	if err := config.M.Save(); err != nil {
		logger.Warning.Printf("Save config error: %s", err.Error())
	}

	if err = req.Do(); err != nil {
		logger.Error.Fatal(err)
	}

	mp3List, err := req.Extract()
	if err != nil {
		logger.Error.Fatal(err)
	}

	n := config.MP3ConcurrentDownloadTasksNumber
	if n > config.MaxConcurrentDownloadTasksNumber {
		n = config.MaxConcurrentDownloadTasksNumber
	}
	switch {
	case n > 1:
		handler.ConcurrentDownload(mp3List, n)
	default:
		handler.SingleDownload(mp3List)
	}
}
