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
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	Ver = "0.1"
)

const (
	ContentType = "Content-Type"
	TypeForm    = "application/x-www-form-urlencoded"
	TypeJSON    = "application/json"
)

var std = New()

type (
	Values  map[string]string
	Object  map[string]interface{}
	Cookies []*http.Cookie
)

type File struct {
	FiledName string
	FileName  string
	File      io.Reader
}

type SSL struct {
	Cert       string
	ClientCert string
	ClientKey  string
}

type Option func(*Request)

type Request struct {
	client   *http.Client
	method   string
	url      string
	params   Values
	formData Values
	json     Object
	headers  Values
	cookies  Cookies
	files    []File
	mux      *sync.Mutex
	withLock bool
}

type Result struct {
	Resp *http.Response
	Err  error
}

func (v Values) Get(key string) string {
	return v[key]
}

func (v Values) Set(key string, value string) {
	v[key] = value
}

func (v Values) Del(key string) {
	delete(v, key)
}

func New(options ...Option) *Request {
	req := &Request{
		client:   http.DefaultClient,
		params:   make(Values),
		formData: make(Values),
		json:     make(Object),
		headers:  make(Values),
		mux:      new(sync.Mutex),
	}

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

func WithProxy(url string) Option {
	return func(req *Request) {
		transport, _ := req.client.Transport.(*http.Transport)
		proxyURL, err := urlpkg.Parse(url)
		if err != nil {
			transport.Proxy = http.ProxyFromEnvironment
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
}

func EnableSession() Option {
	return func(req *Request) {
		jar, _ := cookiejar.New(&cookiejar.Options{
			PublicSuffixList: publicsuffix.List,
		})
		req.client.Jar = jar
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
		transport, _ := req.client.Transport.(*http.Transport)
		transport.DisableKeepAlives = true
	}
}

func InsecureSkipVerify() Option {
	return func(req *Request) {
		transport, _ := req.client.Transport.(*http.Transport)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

func (req *Request) Acquire() *Request {
	req.mux.Lock()
	req.withLock = true
	return req
}

func (req *Request) Reset() {
	req.method = ""
	req.url = ""
	req.params = make(Values)
	req.formData = make(Values)
	req.json = make(Object)
	req.headers = make(Values)
	req.cookies = make([]*http.Cookie, 0)
	req.files = make([]File, 0)

	if req.withLock {
		req.withLock = false
		req.mux.Unlock()
	}
}

func (req *Request) Params(params Values) *Request {
	req.params = params
	return req
}

func (req *Request) FormData(formData Values) *Request {
	req.headers.Set(ContentType, TypeForm)
	for k, v := range formData {
		req.formData.Set(k, v)
	}
	return req
}

func (req *Request) Json(object Object) *Request {
	req.headers.Set(ContentType, TypeJSON)
	req.json = object
	return req
}

func (req *Request) Headers(headers Values) *Request {
	for k, v := range headers {
		req.headers.Set(k, v)
	}
	return req
}

func (req *Request) Cookies(cookies Cookies) *Request {
	req.cookies = cookies
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

func (req *Request) AcquireCertificate(ssl SSL) error {
	var err error
	var pool *x509.CertPool
	if ssl.Cert != "" {
		caCert, err := ioutil.ReadFile(ssl.Cert)
		if err != nil {
			return err
		}
		pool = x509.NewCertPool()
		pool.AppendCertsFromPEM(caCert)
	}

	var clientCert tls.Certificate
	if ssl.ClientCert != "" && ssl.ClientKey != "" {
		clientCert, err = tls.LoadX509KeyPair(ssl.ClientCert, ssl.ClientKey)
		if err != nil {
			return err
		}
	}

	transport, _ := req.client.Transport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{clientCert},
	}
	return nil
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
	if strings.HasPrefix(contentType, TypeForm) {
		httpReq, err = req.buildFormRequest()
	} else if strings.HasPrefix(contentType, TypeJSON) {
		httpReq, err = req.buildJSONRequest()
	} else {
		httpReq, err = req.buildEmptyRequest()
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

func (req *Request) buildFormRequest() (*http.Request, error) {
	form := urlpkg.Values{}
	for k, v := range req.formData {
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

func (req *Request) buildEmptyRequest() (*http.Request, error) {
	return http.NewRequest(req.method, req.url, nil)
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

func (req *Request) Get(url string) *Request {
	req.Reset()
	req.method = http.MethodGet
	req.url = url
	return req
}

func (req *Request) Head(url string) *Request {
	req.Reset()
	req.method = http.MethodHead
	req.url = url
	return req
}

func (req *Request) Post(url string) *Request {
	req.Reset()
	req.method = http.MethodPost
	req.url = url
	return req
}

func (req *Request) Put(url string) *Request {
	req.Reset()
	req.method = http.MethodPut
	req.url = url
	return req
}

func (req *Request) Patch(url string) *Request {
	req.Reset()
	req.method = http.MethodPatch
	req.url = url
	return req
}

func (req *Request) Delete(url string) *Request {
	req.Reset()
	req.method = http.MethodDelete
	req.url = url
	return req
}

func (req *Request) Connect(url string) *Request {
	req.Reset()
	req.method = http.MethodConnect
	req.url = url
	return req
}

func (req *Request) Options(url string) *Request {
	req.Reset()
	req.method = http.MethodOptions
	req.url = url
	return req
}

func (req *Request) Trace(url string) *Request {
	req.method = http.MethodTrace
	req.url = url
	return req
}

func Get(url string) *Request {
	return std.Get(url)
}

func Head(url string) *Request {
	return std.Head(url)
}

func Post(url string) *Request {
	return std.Post(url)
}

func Put(url string) *Request {
	return std.Put(url)
}

func Patch(url string) *Request {
	return std.Get(url)
}

func Delete(url string) *Request {
	return std.Delete(url)
}

func Connect(url string) *Request {
	return std.Connect(url)
}

func Options(url string) *Request {
	return std.Options(url)
}

func Trace(url string) *Request {
	return std.Trace(url)
}

func (r *Result) EnsureStatusOk() *Result {
	if r.Err != nil {
		return r
	}
	if r.Resp.StatusCode != http.StatusOK {
		r.Err = fmt.Errorf("status code requires 200 but got: %d", r.Resp.StatusCode)
		return r
	}

	return r
}

func (r *Result) EnsureStatus2xx() *Result {
	if r.Err != nil {
		return r
	}
	if r.Resp.StatusCode/100 == 2 {
		r.Err = fmt.Errorf("status code requires 2xx but got: %d", r.Resp.StatusCode)
		return r
	}

	return r
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
