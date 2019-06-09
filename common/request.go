package common

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"strings"
	"sync"
	"time"

	"github.com/winterssy/music-get/config"
	"golang.org/x/net/publicsuffix"
)

const (
	NeteaseMusicOrigin  = "https://music.163.com"
	NeteaseMusicReferer = "https://music.163.com"
	TencentMusicOrigin  = "https://c.y.qq.com"
	TencentMusicReferer = "https://c.y.qq.com"
	RequestTimeout      = 120 * time.Second
)

var (
	getHTTPClientOnce     sync.Once
	loadCachedCookiesOnce sync.Once
	DefaultHTTPClient     *http.Client
)

// any parsed request must implement this interface
type MusicRequest interface {
	RequireLogin() bool
	Login() error
	Do() error
	Extract() ([]*MP3, error)
}

func loadCachedCookies(reqURL *urlpkg.URL, client *http.Client) {
	f := func() {
		if len(config.M.Cookies) > 0 {
			client.Jar.SetCookies(reqURL, config.M.Cookies)
		}
	}
	loadCachedCookiesOnce.Do(f)
}

func getHTTPClient() *http.Client {
	f := func() {
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		DefaultHTTPClient = &http.Client{
			Timeout: RequestTimeout,
			Jar:     jar,
		}
	}
	getHTTPClientOnce.Do(f)
	return DefaultHTTPClient
}

func chooseUserAgent() string {
	var userAgentList = []string{
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko) Version/9.0 Mobile/13B143 Safari/601.1",
		"Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 5.1.1; Nexus 6 Build/LYZ28E) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Mobile/14F89;GameHelper",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
		"Mozilla/5.0 (iPad; CPU OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A300 Safari/602.1",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:46.0) Gecko/20100101 Firefox/46.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_5) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.1.1 Safari/603.2.4",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:46.0) Gecko/20100101 Firefox/46.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/13.10586",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return userAgentList[r.Intn(len(userAgentList))]
}

func Request(method, url string, query map[string]string, body io.Reader, origin int) (*http.Response, error) {
	reqURL, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}

	client := getHTTPClient()
	loadCachedCookies(reqURL, client)

	method = strings.ToUpper(method)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	switch origin {
	case NeteaseMusic:
		req.Header.Set("Origin", NeteaseMusicOrigin)
		req.Header.Set("Referer", NeteaseMusicReferer)
	case TencentMusic:
		req.Header.Set("Origin", TencentMusicOrigin)
		req.Header.Set("Referer", TencentMusicReferer)
	}
	req.Header.Set("User-Agent", chooseUserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("%s %s error: %s", method, url, resp.Status)
	}

	client.Jar.SetCookies(reqURL, resp.Cookies())
	return resp, nil
}
