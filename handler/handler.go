package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/winterssy/music-get/conf"
)

const (
	LogFileName = "music-get.log"
)

func outputLog(dlErrs []*DownloadError) {
	if len(dlErrs) == 0 {
		return
	}

	var data bytes.Buffer
	for _, i := range dlErrs {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "\t")
		if enc.Encode(i) != nil {
			continue
		}
		data.Write(buf.Bytes())
	}

	if writeToFile(LogFileName, data.Bytes()) == nil {
		fmt.Printf("\nSee more info in %q\n", LogFileName)
	}
}

func writeToFile(fileName string, data []byte) error {
	logFilePath := filepath.Join(conf.Conf.Workspace, fileName)
	return ioutil.WriteFile(logFilePath, data, 0644)
}
