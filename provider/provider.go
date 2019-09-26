package provider

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/winterssy/easylog"
	"github.com/winterssy/music-get/conf"
	"github.com/winterssy/music-get/pkg/ecode"
	"github.com/winterssy/music-get/pkg/requests"
	"github.com/winterssy/music-get/utils"
)

const (
	NetEaseMusic = iota
	QQMusic
)

var (
	userAgent     = chooseUserAgent()
	RequestHeader = map[int]requests.Values{
		NetEaseMusic: requests.Values{
			"Origin":     "https://music.163.com",
			"Referer":    "https://music.163.com",
			"User-Agent": userAgent,
		},
		QQMusic: requests.Values{
			"Origin":     "https://c.y.qq.com",
			"Referer":    "https://c.y.qq.com",
			"User-Agent": userAgent,
		},
	}
)

type MusicRequest interface {
	// 是否需要登录
	RequireLogin() bool
	// 发起登录请求
	Login() error
	// 发起API请求
	Do() error
	// 解析API响应获取音源
	Prepare() ([]*MP3, error)
}

type MP3 struct {
	FileName    string
	SavePath    string
	Playable    bool
	DownloadURL string
	Provider    int
}

type DownloadTask struct {
	MP3    *MP3
	Status int
}

func (m *MP3) SingleDownload() (status int) {
	defer func() {
		switch status {
		case ecode.Success:
			easylog.Infof("Download complete")
		case ecode.NoCopyright, ecode.AlreadyDownloaded:
			easylog.Warnf("Download interrupt: %s", ecode.Message(status))
		default:
			easylog.Errorf("Download error: %s", ecode.Message(status))
		}
	}()

	if !m.Playable {
		status = ecode.NoCopyright
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

	easylog.Infof("Downloading: %s", m.FileName)
	resp, err := Request.Get(m.DownloadURL).
		Headers(RequestHeader[m.Provider]).
		Cookies(conf.Conf.Cookies).
		Send().
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

func (m *MP3) ConcurrentDownload(taskList chan DownloadTask, taskQueue chan struct{}, wg *sync.WaitGroup) {
	var status int

	defer func() {
		switch status {
		case ecode.Success:
			easylog.Infof("Download complete: %s", m.FileName)
		case ecode.NoCopyright, ecode.AlreadyDownloaded:
			easylog.Warnf("Download interrupt: %s: %s", m.FileName, ecode.Message(status))
		default:
			easylog.Errorf("Download error: %s: %s", m.FileName, ecode.Message(status))
		}
		wg.Done()
		taskList <- DownloadTask{m, status}
		<-taskQueue
	}()

	if !m.Playable {
		status = ecode.NoCopyright
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

	easylog.Infof("Downloading: %s", m.FileName)
	resp, err := Request.
		Acquire().
		Get(m.DownloadURL).
		Headers(RequestHeader[m.Provider]).
		Cookies(conf.Conf.Cookies).
		Send().
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
		status = ecode.BuildFileException
		return
	}

	status = ecode.Success
	return
}
