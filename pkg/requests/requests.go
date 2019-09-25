package requests

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	Ver = "0.1"
)

const (
	ContentType               = "Content-Type"
	ApplicationFormURLEncoded = "application/x-www-form-urlencoded"
	ApplicationJSON           = "application/json"
)

type (
	Header map[string]string
	Params map[string]string
	Data   map[string]string
)

type FileForm struct {
	FiledName string
	FileName  string
	File      io.Reader
}

type Request struct {
	client  *http.Client
	method  string
	url     string
	params  Params
	data    urlpkg.Values
	json    interface{}
	headers http.Header
	cookies []*http.Cookie
	files   []FileForm
}

type Result struct {
	Resp *http.Response
	Err  error
}

var DefaultClient *http.Client

func init() {
	DefaultClient = newClient()
}

func newClient() *http.Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Jar:     jar,
		Timeout: 120 * time.Second,
	}
}

func New(url string, method string, client *http.Client) *Request {
	if client == nil {
		client = DefaultClient
	}

	req := &Request{
		client:  client,
		method:  method,
		url:     url,
		headers: make(http.Header),
		data:    make(urlpkg.Values),
		params:  make(Params),
	}
	req.headers.Set("User-Agent", "Go-Requests "+Ver)
	return req
}

func (r *Request) Timeout(d time.Duration) *Request {
	r.client.Timeout = d
	return r
}

func (r *Request) Params(params Params) *Request {
	r.params = params
	return r
}

func (r *Request) Data(data Data) *Request {
	r.headers.Set(ContentType, ApplicationFormURLEncoded)
	for k, v := range data {
		r.data.Set(k, v)
	}
	return r
}

func (r *Request) Headers(header Header) *Request {
	for k, v := range header {
		r.headers.Set(k, v)
	}
	return r
}

func (r *Request) Cookies(cookies []*http.Cookie) *Request {
	r.cookies = cookies
	return r
}

func (r *Request) Json(json interface{}) *Request {
	r.headers.Set(ContentType, ApplicationJSON)
	r.json = json
	return r
}

func (r *Request) Send() *Result {
	result := new(Result)
	if r.url == "" {
		result.Err = errors.New("url not specified")
		return result
	}

	var req *http.Request
	var err error
	contentType := r.headers.Get(ContentType)
	if strings.HasPrefix(contentType, ApplicationFormURLEncoded) {
		req, err = r.buildFormRequest()
	} else if strings.HasPrefix(contentType, ApplicationJSON) {
		req, err = r.buildJsonRequest()
	} else {
		req, err = r.buildEmptyRequest()
	}
	if err != nil {
		result.Err = err
		return result
	}

	if len(r.params) != 0 {
		r.addQueryParams(req)
	}
	if len(r.cookies) != 0 {
		r.addCookies(req)
	}
	req.Header = r.headers

	result.Resp, err = r.client.Do(req)
	return result
}

func (r *Request) buildFormRequest() (*http.Request, error) {
	return http.NewRequest(r.method, r.url, strings.NewReader(r.data.Encode()))
}

func (r *Request) buildJsonRequest() (*http.Request, error) {
	b, err := json.Marshal(r.json)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(r.method, r.url, bytes.NewReader(b))
}

func (r *Request) buildEmptyRequest() (*http.Request, error) {
	return http.NewRequest(r.method, r.url, nil)
}

func (r *Request) addQueryParams(req *http.Request) {
	query := req.URL.Query()
	for k, v := range r.params {
		query.Set(k, v)
	}
	req.URL.RawQuery = query.Encode()
}

func (r *Request) addCookies(req *http.Request) {
	for _, c := range r.cookies {
		req.AddCookie(c)
	}
}

func Get(url string) *Request {
	return New(url, http.MethodGet, nil)
}

func Head(url string) *Request {
	return New(url, http.MethodHead, nil)
}

func Post(url string) *Request {
	return New(url, http.MethodPost, nil)
}

func Put(url string) *Request {
	return New(url, http.MethodPut, nil)
}

func Patch(url string) *Request {
	return New(url, http.MethodPatch, nil)
}

func Delete(url string) *Request {
	return New(url, http.MethodDelete, nil)
}

func Connect(url string) *Request {
	return New(url, http.MethodConnect, nil)
}

func Options(url string) *Request {
	return New(url, http.MethodOptions, nil)
}

func Trace(url string) *Request {
	return New(url, http.MethodTrace, nil)
}

func (r *Result) StatusOk() bool {
	return r.Resp.StatusCode == http.StatusOK
}

func (r *Result) Status2xx() bool {
	return r.Resp.StatusCode/100 == 2
}

func (r *Result) Resolve() (*http.Response, error) {
	return r.Resp, r.Err
}

func (r *Result) Raw() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	defer r.Resp.Body.Close()

	b, err := ioutil.ReadAll(r.Resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *Result) Text() (string, error) {
	b, err := r.Raw()
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (r *Result) Json(v interface{}) error {
	b, err := r.Raw()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}
