package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/winterssy/music-get/conf"
)

const (
	LogFileName = "music-get.log"
)

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
	file, err := os.OpenFile(filepath.Join(conf.Conf.Workspace, filename), os.O_CREATE|os.O_WRONLY, 0644)
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
