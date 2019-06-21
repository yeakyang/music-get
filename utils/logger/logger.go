package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	Debug   = log.New(os.Stdout, "[Debug] ", log.LstdFlags|log.Lshortfile)
	Info    = log.New(os.Stdout, "[Info] ", log.LstdFlags)
	Warning = log.New(os.Stdout, "[Warning] ", log.LstdFlags)
	Error   = log.New(os.Stderr, "[Error] ", log.LstdFlags)
)

func WriteToFile(filename string, lines []string) error {
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
