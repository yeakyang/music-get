package handler

import (
	"encoding/json"
	"github.com/winterssy/music-get/utils/logger"
)

const (
	LogFileName = "music-get.log"
)

type DownloadError struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Code int    `json:"code"`
}

func newDownloadError(name, url string, code int) DownloadError {
	return DownloadError{name, url, code}
}

func outputLog(errs []DownloadError) error {
	lines := make([]string, 0, len(errs))
	for _, i := range errs {
		line, err := json.MarshalIndent(i, "", "    ")
		if err != nil {
			continue
		}
		lines = append(lines, string(line))
	}
	return logger.WriteToFile(LogFileName, lines)
}
