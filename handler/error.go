package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	return writeToFile(LogFileName, lines)
}

func writeToFile(filename string, lines []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(currentDir, filename), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return w.Flush()
}
