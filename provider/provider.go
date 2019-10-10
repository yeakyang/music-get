package provider

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/conf"
	"github.com/winterssy/music-get/internal/ecode"
	"github.com/winterssy/music-get/pkg/concurrency"
	"github.com/winterssy/music-get/utils"
)

const (
	NetEaseMusic = iota
	QQMusic
	MiguMusic
	KugouMusic
	KuwoMusic
)

type (
	MusicRequest interface {
		// 是否需要登录
		RequireLogin() bool
		// 发起登录请求
		Login() error
		// 发起API请求
		Do() error
		// 解析API响应获取音源
		Prepare() ([]*MP3, error)
	}

	MP3 struct {
		FileName    string
		SavePath    string
		Playable    bool
		DownloadURL string
		Provider    int
	}

	DownloadTask struct {
		MP3    *MP3
		Status int
	}
)

func (m *MP3) SingleDownload() (status int) {
	defer func() {
		switch status {
		case ecode.Success:
			easylog.Infof("Download complete")
		case ecode.SongUnavailable, ecode.AlreadyDownloaded:
			easylog.Warnf("Download interrupt: %s", ecode.Message(status))
		default:
			easylog.Errorf("Download error: %s", ecode.Message(status))
		}
	}()

	easylog.Infof("Downloading: %s", m.FileName)
	if !m.Playable || m.DownloadURL == "" {
		status = ecode.SongUnavailable
		return
	}

	m.SavePath = filepath.Join(conf.Conf.DownloadDir, m.SavePath)
	if err := utils.BuildPathIfNotExist(m.SavePath); err != nil {
		status = ecode.BuildPathException
		return
	}

	fPath := filepath.Join(m.SavePath, m.FileName)
	if !conf.Conf.DownloadOverwrite {
		if downloaded, _ := utils.ExistsPath(fPath); downloaded {
			status = ecode.AlreadyDownloaded
			return
		}
	}

	easylog.Debugf("URL: %s", m.DownloadURL)
	resp, err := Client(m.Provider).
		Get(m.DownloadURL).
		Resolve()
	if err != nil {
		status = ecode.HTTPRequestException
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		status = ecode.BuildFileException
		return
	}
	defer f.Close()

	bar := pb.Full.Start64(resp.ContentLength)
	bar.Start()
	barReader := bar.NewProxyReader(resp.Body)
	n, err := io.Copy(f, barReader)
	if err != nil || n != resp.ContentLength {
		status = ecode.FileTransferException
		return
	}

	bar.Finish()
	status = ecode.Success
	return
}

func (m *MP3) ConcurrentDownload(taskList chan DownloadTask, c *concurrency.C) {
	var status int

	defer func() {
		switch status {
		case ecode.Success:
			easylog.Infof("Download complete: %s", m.FileName)
		case ecode.SongUnavailable, ecode.AlreadyDownloaded:
			easylog.Warnf("Download interrupt: %s: %s", m.FileName, ecode.Message(status))
		default:
			easylog.Errorf("Download error: %s: %s", m.FileName, ecode.Message(status))
		}
		c.Done()
		taskList <- DownloadTask{m, status}
	}()

	easylog.Infof("Downloading: %s", m.FileName)
	if !m.Playable || m.DownloadURL == "" {
		status = ecode.SongUnavailable
		return
	}

	m.SavePath = filepath.Join(conf.Conf.DownloadDir, m.SavePath)
	if err := utils.BuildPathIfNotExist(m.SavePath); err != nil {
		status = ecode.BuildPathException
		return
	}

	fPath := filepath.Join(m.SavePath, m.FileName)
	if !conf.Conf.DownloadOverwrite {
		if downloaded, _ := utils.ExistsPath(fPath); downloaded {
			status = ecode.AlreadyDownloaded
			return
		}
	}

	easylog.Debugf("URL: %s", m.DownloadURL)
	resp, err := Client(m.Provider).
		Get(m.DownloadURL).
		Resolve()
	if err != nil {
		status = ecode.HTTPRequestException
		return
	}
	defer resp.Body.Close()

	f, err := os.Create(fPath)
	if err != nil {
		status = ecode.BuildFileException
		return
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil || n != resp.ContentLength {
		status = ecode.FileTransferException
		return
	}

	status = ecode.Success
	return
}
