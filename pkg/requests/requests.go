package requests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	Ver = "0.1"

	DefaultTimeout = 120 * time.Second
	ContentType    = "Content-Type"
	TypeForm       = "application/x-www-form-urlencoded"
	TypeJSON       = "application/json"

	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

var (
	std = New()
)

type (
	Option func(*Request)

	Values map[string]string
	Object map[string]interface{}

	File struct {
		FieldName string
		FileName  string
		FilePath  string
	}

	Request struct {
		client   *http.Client
		method   string
		url      string
		params   Values
		data     Values
		json     Object
		headers  Values
		cookies  []*http.Cookie
		files    []*File
		mux      *sync.Mutex
		withLock bool
	}

	Result struct {
		Resp *http.Response
		Err  error
	}
)

func (v Values) Get(key string) string {
	return v[key]
}

func (v Values) Set(key string, value string) {
	v[key] = value
}

func (v Values) Del(key string) {
	delete(v, key)
}

func (obj Object) Get(key string) interface{} {
	return obj[key]
}

func (obj Object) Set(key string, value interface{}) {
	obj[key] = value
}

func (obj Object) Del(key string) {
	delete(obj, key)
}

func New(options ...Option) *Request {
	req := &Request{
		client:  http.DefaultClient,
		params:  make(Values),
		data:    make(Values),
		json:    make(Object),
		headers: make(Values),
		mux:     new(sync.Mutex),
	}

	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	req.client.Jar = jar
	req.client.Transport = http.DefaultTransport
	req.client.Timeout = DefaultTimeout

	for _, opt := range options {
		opt(req)
	}

	req.headers.Set("User-Agent", "Go-Requests "+Ver)
	return req
}

func WithTransport(transport http.RoundTripper) Option {
	return func(req *Request) {
		req.client.Transport = transport
	}
}

func WithRedirectPolicy(policy func(req *http.Request, via []*http.Request) error) Option {
	return func(req *Request) {
		req.client.CheckRedirect = policy
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(req *Request) {
		req.client.Timeout = timeout
	}
}

func WithClientCertificates(certs ...tls.Certificate) Option {
	return func(req *Request) {
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, certs...)
	}
}

func WithRootCAs(pemFilePath string) Option {
	return func(req *Request) {
		pemCert, err := ioutil.ReadFile(pemFilePath)
		if err != nil {
			return
		}
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		if transport.TLSClientConfig.RootCAs == nil {
			transport.TLSClientConfig.RootCAs = x509.NewCertPool()
		}
		transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCert)
	}
}

func ProxyFromURL(url string) Option {
	return func(req *Request) {
		proxyURL, err := urlpkg.Parse(url)
		if err != nil {
			return
		}
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
}

func DisableProxyFromEnvironment() Option {
	return func(req *Request) {
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		transport.Proxy = nil
	}
}

func DisableSession() Option {
	return func(req *Request) {
		req.client.Jar = nil
	}
}

func DisableRedirect() Option {
	return func(req *Request) {
		req.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

func DisableKeepAlives() Option {
	return func(req *Request) {
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		transport.DisableKeepAlives = true
	}
}

func InsecureSkipVerify() Option {
	return func(req *Request) {
		transport, ok := req.client.Transport.(*http.Transport)
		if !ok {
			return
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}
}

// for concurrent reuse
func (req *Request) AcquireLock() *Request {
	req.mux.Lock()
	req.withLock = true
	return req
}

func (req *Request) Reset() {
	req.method = ""
	req.url = ""
	req.params = make(Values)
	req.data = make(Values)
	req.json = make(Object)
	req.headers = make(Values)
	req.cookies = nil
	req.files = nil

	if req.withLock {
		req.mux.Unlock()
	}
}

func Get(url string) *Request {
	return std.Get(url)
}

func (req *Request) Get(url string) *Request {
	req.method = MethodGet
	req.url = url
	return req
}

func Head(url string) *Request {
	return std.Head(url)
}

func (req *Request) Head(url string) *Request {
	req.method = MethodHead
	req.url = url
	return req
}

func Post(url string) *Request {
	return std.Post(url)
}

func (req *Request) Post(url string) *Request {
	req.method = MethodPost
	req.url = url
	return req
}

func Put(url string) *Request {
	return std.Put(url)
}

func (req *Request) Put(url string) *Request {
	req.method = MethodPut
	req.url = url
	return req
}

func Patch(url string) *Request {
	return std.Get(url)
}

func (req *Request) Patch(url string) *Request {
	req.method = MethodPatch
	req.url = url
	return req
}

func Delete(url string) *Request {
	return std.Delete(url)
}

func (req *Request) Delete(url string) *Request {
	req.method = MethodDelete
	req.url = url
	return req
}

func Connect(url string) *Request {
	return std.Connect(url)
}

func (req *Request) Connect(url string) *Request {
	req.method = MethodConnect
	req.url = url
	return req
}

func Options(url string) *Request {
	return std.Options(url)
}

func (req *Request) Options(url string) *Request {
	req.method = MethodOptions
	req.url = url
	return req
}

func Trace(url string) *Request {
	return std.Trace(url)
}

func (req *Request) Trace(url string) *Request {
	req.method = MethodTrace
	req.url = url
	return req
}

func (req *Request) Params(params Values) *Request {
	for k, v := range params {
		req.params.Set(k, v)
	}
	return req
}

func (req *Request) Data(data Values) *Request {
	req.headers.Set(ContentType, TypeForm)
	for k, v := range data {
		req.data.Set(k, v)
	}
	return req
}

func (req *Request) JSON(obj Object) *Request {
	req.headers.Set(ContentType, TypeJSON)
	for k, v := range obj {
		req.json.Set(k, v)
	}
	return req
}

func (req *Request) Files(files ...*File) *Request {
	req.files = append(req.files, files...)
	return req
}

func (req *Request) Headers(headers Values) *Request {
	for k, v := range headers {
		req.headers.Set(k, v)
	}
	return req
}

func (req *Request) Cookies(cookies ...*http.Cookie) *Request {
	req.cookies = append(req.cookies, cookies...)
	return req
}

func (req *Request) BasicAuth(username, password string) *Request {
	req.headers.Set("Authorization", "Basic "+basicAuth(username, password))
	return req
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (req *Request) BearerToken(token string) *Request {
	req.headers.Set("Authorization", "Bearer "+token)
	return req
}

func (req *Request) Send() *Result {
	result := new(Result)
	if req.url == "" {
		result.Err = errors.New("url not specified")
		req.Reset()
		return result
	}

	var httpReq *http.Request
	var err error
	contentType := req.headers.Get(ContentType)
	if req.files != nil {
		httpReq, err = req.buildMultipartRequest()
	} else if strings.HasPrefix(contentType, TypeForm) {
		httpReq, err = req.buildFormRequest()
	} else if strings.HasPrefix(contentType, TypeJSON) {
		httpReq, err = req.buildJSONRequest()
	} else {
		httpReq, err = req.buildStdRequest()
	}
	if err != nil {
		result.Err = err
		req.Reset()
		return result
	}

	if len(req.params) != 0 {
		req.addParams(httpReq)
	}
	if len(req.headers) != 0 {
		req.addHeaders(httpReq)
	}
	if len(req.cookies) != 0 {
		req.addCookies(httpReq)
	}

	req.Reset()

	result.Resp, err = req.client.Do(httpReq)
	return result
}

func (req *Request) buildStdRequest() (*http.Request, error) {
	return http.NewRequest(req.method, req.url, nil)
}

func (req *Request) buildFormRequest() (*http.Request, error) {
	form := urlpkg.Values{}
	for k, v := range req.data {
		form.Set(k, v)
	}
	return http.NewRequest(req.method, req.url, strings.NewReader(form.Encode()))
}

func (req *Request) buildJSONRequest() (*http.Request, error) {
	b, err := json.Marshal(req.json)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(req.method, req.url, bytes.NewReader(b))
}

func (req *Request) buildMultipartRequest() (*http.Request, error) {
	r, w := io.Pipe()
	mw := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer mw.Close()

		for _, i := range req.files {
			fieldName, fileName, filePath := i.FieldName, i.FileName, i.FilePath
			if fieldName == "" {
				fileName = "data"
			}
			if fileName == "" {
				fileName = filepath.Base(filePath)
			}

			part, err := mw.CreateFormFile(fieldName, fileName)
			if err != nil {
				return
			}
			file, err := os.Open(filePath)
			if err != nil {
				return
			}
			if _, err = io.Copy(part, file); err != nil {
				return
			}

			file.Close()
		}
	}()

	req.headers.Set(ContentType, mw.FormDataContentType())
	return http.NewRequest(req.method, req.url, r)
}

func (req *Request) addParams(httpReq *http.Request) {
	query := httpReq.URL.Query()
	for k, v := range req.params {
		query.Set(k, v)
	}
	httpReq.URL.RawQuery = query.Encode()
}

func (req *Request) addHeaders(httpReq *http.Request) {
	for k, v := range req.headers {
		httpReq.Header.Set(k, v)
	}
}

func (req *Request) addCookies(httpReq *http.Request) {
	for _, c := range req.cookies {
		httpReq.AddCookie(c)
	}
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

func (r *Result) JSON(v interface{}) error {
	b, err := r.Raw()
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (r *Result) EnsureStatusOk() *Result {
	if r.Err != nil {
		return r
	}
	if r.Resp.StatusCode != http.StatusOK {
		r.Err = fmt.Errorf("status code 200 expected but got: %d", r.Resp.StatusCode)
		return r
	}

	return r
}

func (r *Result) EnsureStatus2xx() *Result {
	if r.Err != nil {
		return r
	}
	if r.Resp.StatusCode/100 == 2 {
		r.Err = fmt.Errorf("status code 2xx expected but got: %d", r.Resp.StatusCode)
		return r
	}

	return r
}
