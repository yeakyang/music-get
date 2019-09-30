package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/winterssy/music-get/conf"
)

const (
	LogFileName = "music-get.log"
)

func outputLog(errs []DownloadError) {
	n := len(errs)
	if n == 0 {
		return
	}

	lines := make([]string, 0, n)
	for _, i := range errs {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "    ")
		if enc.Encode(i) != nil {
			continue
		}
		lines = append(lines, buf.String())
	}

	if writeToFile(LogFileName, lines) == nil {
		fmt.Printf("\nSee more info in %q\n", LogFileName)
	}
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
