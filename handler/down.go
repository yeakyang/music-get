package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/winterssy/music-get/internal/ecode"
	"github.com/winterssy/music-get/pkg/concurrency"
	"github.com/winterssy/music-get/provider"
)

type (
	DownloadError struct {
		FileName string `json:"filename"`
		URL      string `json:"url,omitempty"`
		Code     int    `json:"code"`
		Reason   string `json:"reason"`
	}
)

func SingleDownload(mp3List []*provider.MP3) {
	total, success, failure, ignore := len(mp3List), 0, 0, 0

	dlErrs := make([]*DownloadError, 0)
	for _, m := range mp3List {
		switch status := m.SingleDownload(); status {
		case ecode.Success:
			success++
		case ecode.AlreadyDownloaded:
			ignore++
		default:
			failure++
			dlErrs = append(dlErrs, &DownloadError{
				FileName: m.FileName,
				URL:      m.DownloadURL,
				Code:     status,
				Reason:   ecode.Message(status),
			})
			if status != ecode.SongUnavailable {
				// ignore error
				os.Remove(filepath.Join(m.SavePath, m.FileName))
			}
		}
	}

	fmt.Printf("\nDownload report --> total: %d, success: %d, failure: %d, ignore: %d\n", total, success, failure, ignore)
	outputLog(dlErrs)
}

func ConcurrentDownload(mp3List []*provider.MP3, n int) {
	total, success, failure, ignore := len(mp3List), 0, 0, 0

	c := concurrency.New(n)
	taskList := make(chan provider.DownloadTask, total)
	for _, i := range mp3List {
		c.Add(1)
		go i.ConcurrentDownload(taskList, c)
	}
	c.Wait()

	dlErrs := make([]*DownloadError, 0)
	for range mp3List {
		task := <-taskList
		switch task.Status {
		case ecode.Success:
			success++
		case ecode.AlreadyDownloaded:
			ignore++
		default:
			failure++
			dlErrs = append(dlErrs, &DownloadError{
				FileName: task.MP3.FileName,
				URL:      task.MP3.DownloadURL,
				Code:     task.Status,
				Reason:   ecode.Message(task.Status),
			})
			if task.Status != ecode.SongUnavailable {
				// ignore error
				os.Remove(filepath.Join(task.MP3.SavePath, task.MP3.FileName))
			}
		}
	}

	fmt.Printf("\nDownload report --> total: %d, success: %d, failure: %d, ignore: %d\n", total, success, failure, ignore)
	outputLog(dlErrs)
}
