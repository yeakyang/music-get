package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/winterssy/music-get/common"
)

func SingleDownload(mp3List []*common.MP3) {
	total, success, failure, ignore := len(mp3List), 0, 0, 0

	wg := &sync.WaitGroup{}
	var failureInfo []DownloadError
	for _, m := range mp3List {
		switch status := m.SingleDownload(); status {
		case common.DownloadSuccess:
			success++
			wg.Add(1)
			go m.UpdateTag(wg)
		case common.DownloadNoCopyrightError, common.DownloadAlready:
			ignore++
		default:
			failure++
			failureInfo = append(failureInfo, newDownloadError(m.FileName, m.DownloadURL, status))
			// ignore error
			os.Remove(filepath.Join(m.SavePath, m.FileName))
		}
	}
	wg.Wait()

	fmt.Printf("\nDownload report --> total: %d, success: %d, failure: %d, ignore: %d\n", total, success, failure, ignore)

	if len(failureInfo) == 0 {
		return
	}
	if err := outputLog(failureInfo); err == nil {
		fmt.Printf("\nSee more info in %q\n", LogFileName)
	}
}

func ConcurrentDownload(mp3List []*common.MP3, n int) {
	total, success, failure, ignore := len(mp3List), 0, 0, 0

	taskList := make(chan common.DownloadTask, total)
	taskQueue := make(chan struct{}, n)
	wg := &sync.WaitGroup{}
	wg.Add(total)
	for _, m := range mp3List {
		taskQueue <- struct{}{}
		go m.ConcurrentDownload(taskList, taskQueue, wg)
	}
	wg.Wait()

	var failureInfo []DownloadError
	for range mp3List {
		task := <-taskList
		switch status := task.Status; status {
		case common.DownloadSuccess:
			success++
			wg.Add(1)
			go task.MP3.UpdateTag(wg)
		case common.DownloadNoCopyrightError, common.DownloadAlready:
			ignore++
		default:
			failure++
			failureInfo = append(failureInfo, newDownloadError(task.MP3.FileName, task.MP3.DownloadURL, status))
			// ignore error
			os.Remove(filepath.Join(task.MP3.SavePath, task.MP3.FileName))
		}
	}
	wg.Wait()

	fmt.Printf("\nDownload report --> total: %d, success: %d, failure: %d, ignore: %d\n", total, success, failure, ignore)

	if len(failureInfo) == 0 {
		return
	}
	if err := outputLog(failureInfo); err == nil {
		fmt.Printf("\nSee more info in %q\n", LogFileName)
	}
}
