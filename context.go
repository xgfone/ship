// Copyright 2018 xgfone <xgfone@126.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ship

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/xgfone/ship/core"
	"github.com/xgfone/ship/utils"
)

var (
	indexPage = "index.html"
	emptyStrS = [256]string{}

	emptyValue = emptyType(0)
)

type emptyType uint8

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

// Context is the alias of core.Context, which stands for a context.
//
// Methods:
//
//    // Report whether the response has been sent.
//    IsResponse() bool
//
//    // Find the registered router handler by the method and path of the request.
//    //
//    // Return nil if not found.
//    FindHandler(method string, path string) Handler
//
//    NotFoundHandler() Handler
//
//    AcquireBuffer() *bytes.Buffer
//    ReleaseBuffer(*bytes.Buffer)
//
//    Request() *http.Request
//    Response() http.ResponseWriter
//    SetResponse(http.ResponseWriter)
//    SetResponded(bool)
//    SetConnectionClose() // Tell the server to close the connection.
//
//    // These may be passed the error between the handlers.
//    Error() error
//    HasError() bool
//    SetError(err error)
//
//    IsTLS() bool
//    IsDebug() bool
//    IsAjax() bool
//    IsWebSocket() bool
//
//    Header(name string) (value string)
//    SetHeader(name string, value string)
//
//    // URL Parameter. We remove the URL prefix for the convenience.
//    Param(name string) (value string)
//    Params() map[string]string // Return the key-value map of the url parameters
//    ParamNames() []string      // Return the list of the url parameter names
//    ParamValues() []string     // Return the list of the url parameter values
//
//    // Accept returns the content of the header Accept.
//    //
//    // If there is no the header Accept , it return nil.
//    //
//    // Notice:
//    //
//    //   1. It will sort the content by the q-factor weighting.
//    //   2. If the value is "<MIME_type>/*", it will be amended as "<MIME_type>/".
//    //      So you can use it to match the prefix.
//    //   3. If the value is "*/*", it will be amended as "".
//    //
//    Accept() []string
//    Host() string
//    Method() string
//    Scheme() string
//    RealIP() string
//    Charset() string
//    RemoteAddr() string
//    RequestURI() string
//    ContentType() string // It should not contain the charset.
//    ContentLength() int64
//    GetBody() (string, error)
//    // You should call Context.ReleaseBuffer(buf) to release the buffer at last.
//    GetBodyReader() (*bytes.Buffer, error)
//    SetContentType(string)
//
//    QueryParam(name string) (value string)
//    QueryParams() url.Values
//    QueryRawString() string
//
//    FormParams() (url.Values, error)
//    FormValue(name string) (value string)
//    FormFile(name string) (*multipart.FileHeader, error)
//    MultipartForm() (*multipart.Form, error)
//
//    Cookies() []*http.Cookie
//    Cookie(name string) (*http.Cookie, error)
//    SetCookie(cookie *http.Cookie)
//
//    // If the session id does not exist, it maybe return (nil, nil).
//    //
//    // Notice: for the same session id, the context maybe optimize GetSession
//    // by the cache, which will call the backend store only once.
//    GetSession(id string) (interface{}, error)
//    // id must not be "".
//    //
//    // value should not be nil. If nil, however, it will tell the context
//    // that the session id is missing, and the context should not forward
//    // the request to the underlying session store when calling GetSession.
//    SetSession(id string, value interface{}) error
//    // id must not be "".
//    DelSession(id string) error
//
//    // Get and Set are used to store the key-value information about the context.
//    Store() map[string]interface{}
//    Get(key string) (value interface{})
//    Set(key string, value interface{})
//    Del(key string)
//
//    // You can set a handler then call it across the functions, which is used to
//    // handle the various arguments. For example,
//    //
//    //    responder := func(ctx Context, args ...interface{}) error {
//    //        switch len(args) {
//    //        case 0:
//    //            return ctx.NoContent(http.StatusOK)
//    //        case 1:
//    //            switch v := args[0].(type) {
//    //            case int:
//    //                return ctx.NoContent(v)
//    //            case string:
//    //                return ctx.String(http.StatusOK, v)
//    //            }
//    //        case 2:
//    //            switch v0 := args[0].(type) {
//    //            case int:
//    //                return ctx.String(v0, fmt.Sprintf("%v", args[1]))
//    //            }
//    //        }
//    //        return ctx.NoContent(http.StatusInternalServerError)
//    //    }
//    //
//    //    sethandler := func(next Handler) Handler {
//    //        return func(ctx Context) error {
//    //            ctx.SetHandler(responder)
//    //            return next(ctx)
//    //        }
//    //    }
//    //
//    //    router := New()
//    //    router.Use(sethandler)
//    //    router.Route("/path1").GET(func(c Context) error { return c.Handle() })
//    //    router.Route("/path2").GET(func(c Context) error { return c.Handle(200) })
//    //    router.Route("/path3").GET(func(c Context) error { return c.Handle("Hello, World") })
//    //    router.Route("/path4").GET(func(c Context) error { return c.Handle(200, "Hello, World") })
//    //
//    SetHandler(func(ctx Context, args ...interface{}) error)
//    Handle(args ...interface{}) error
//
//    Logger() Logger
//    URL(name string, params ...interface{}) string
//
//    Bind(v interface{}) error
//    BindQuery(v interface{}) error
//
//    Write([]byte) (int, error)
//    Render(name string, code int, data interface{}) error
//
//    NoContent(code int) error
//    Redirect(code int, toURL string) error
//    String(code int, data string) error
//    Blob(code int, contentType string, b []byte) error
//
//    HTML(code int, html string) error
//    HTMLBlob(code int, b []byte) error
//
//    JSON(code int, i interface{}) error
//    JSONBlob(code int, b []byte) error
//    JSONPretty(code int, i interface{}, indent string) error
//
//    JSONP(code int, callback string, i interface{}) error
//    JSONPBlob(code int, callback string, b []byte) error
//
//    XML(code int, i interface{}) error
//    XMLBlob(code int, b []byte) error
//    XMLPretty(code int, i interface{}, indent string) error
//
//    File(file string) error
//    Inline(file string, name string) error
//    Attachment(file string, name string) error
//    Stream(code int, contentType string, r io.Reader) error
//
type Context = core.Context

type contextKeyT int

var contextKey contextKeyT

func setContext(ctx *contextT) {
	if ctx.req != nil {
		ctx.req = ctx.req.WithContext(context.WithValue(context.TODO(), contextKey, ctx))
	}
}

// GetContext gets the Context from the http Request.
func GetContext(req *http.Request) Context {
	if v := req.Context().Value(contextKey); v != nil {
		return v.(Context)
	}
	return nil
}

type responder struct {
	http.ResponseWriter
	ctx *contextT
}

// WriteHeader implements http.ResponseWriter#WriteHeader().
func (r responder) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.ctx.wrote = true
}

// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r responder) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}

// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r responder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r responder) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Context stands for a request and response context.
type contextT struct {
	wrote bool

	err   error
	req   *http.Request
	resp  responder
	query url.Values

	handler func(Context, ...interface{}) error
	pnames  []string
	pvalues []string

	ship     *Ship
	debug    bool
	logger   Logger
	binder   Binder
	renderer Renderer
	binderQ  func(url.Values, interface{}) error
	router   Router
	session  Session

	sessionK string
	sessionV interface{}

	store map[string]interface{}
}

// NewContext returns a new context.
func newContext(s *Ship, req *http.Request, resp http.ResponseWriter, maxParam int) *contextT {
	var pnames, pvalues []string
	if maxParam > 0 {
		pnames = make([]string, maxParam)
		pvalues = make([]string, maxParam)
	}

	ctx := &contextT{
		pnames:  pnames,
		pvalues: pvalues,

		store: make(map[string]interface{}, s.config.ContextStoreSize),
	}
	ctx.setShip(s)
	ctx.setReqResp(req, resp)
	return ctx
}

func (c *contextT) reset() {
	c.err = nil
	c.req = nil
	c.resp.ResponseWriter = nil
	c.resp.ctx = nil
	c.query = nil
	c.wrote = false
	c.router = nil
	c.handler = nil
	c.sessionK = ""
	c.sessionV = nil

	c.resetURLParam()
	for key := range c.store {
		delete(c.store, key)
	}
}

func (c *contextT) resetURLParam() {
	copy(c.pnames, emptyStrS[:len(c.pnames)])
	copy(c.pvalues, emptyStrS[:len(c.pvalues)])
}

func (c *contextT) setShip(s *Ship) {
	c.ship = s
	c.debug = s.config.Debug
	c.logger = s.config.Logger
	c.binder = s.config.Binder
	c.renderer = s.config.Renderer
	c.binderQ = s.config.BindQuery
	c.session = s.config.Session
}

func (c *contextT) setReqResp(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.resp.ResponseWriter = w
	c.resp.ctx = c
	setContext(c)
}

func (c *contextT) AcquireBuffer() *bytes.Buffer {
	return c.ship.AcquireBuffer()
}

func (c *contextT) ReleaseBuffer(buf *bytes.Buffer) {
	c.ship.ReleaseBuffer(buf)
}

// URL generates an URL from route name and provided parameters.
func (c *contextT) URL(name string, params ...interface{}) string {
	return c.ship.URL(name, params...)
}

func (c *contextT) FindHandler(method, path string) Handler {
	c.resetURLParam()
	return c.ship.router.Find(method, path, c.pnames, c.pvalues)
}

func (c *contextT) NotFoundHandler() Handler {
	return c.ship.config.NotFoundHandler
}

func (c *contextT) IsResponse() bool {
	return c.wrote
}

func (c *contextT) SetResponded(yes bool) {
	c.wrote = yes
}

func (c *contextT) Error() error {
	return c.err
}

func (c *contextT) SetError(err error) {
	c.err = err
}

func (c *contextT) HasError() bool {
	return c.err != nil
}

// Param returns the parameter value in the url path by name.
func (c *contextT) Param(name string) string {
	_len := len(c.pnames)
	for i := 0; i < _len; i++ {
		if len(c.pnames[i]) == 0 {
			return ""
		} else if c.pnames[i] == name {
			return c.pvalues[i]
		}
	}
	return ""
}

// Params returns all the parameters as the key-value map in the url path.
func (c *contextT) Params() map[string]string {
	_len := len(c.pnames)
	ms := make(map[string]string, _len)
	for i := 0; i < _len; i++ {
		if len(c.pnames[i]) == 0 {
			break
		}
		ms[c.pnames[i]] = c.pvalues[i]
	}
	return ms
}

func (c *contextT) ParamNames() []string {
	_len := len(c.pnames)
	for i := 0; i < _len; i++ {
		if c.pnames[i] == "" {
			return c.pnames[:i]
		}
	}
	return nil
}

func (c *contextT) ParamValues() []string {
	_len := len(c.pnames)
	for i := 0; i < _len; i++ {
		if c.pnames[i] == "" {
			return c.pvalues[:i]
		}
	}
	return nil
}

func (c *contextT) Store() map[string]interface{} {
	return c.store
}

// Get retrieves data from the context.
func (c *contextT) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *contextT) Set(key string, value interface{}) {
	c.store[key] = value
}

func (c *contextT) Del(key string) {
	delete(c.store, key)
}

// Logger returns the logger implementation.
func (c *contextT) Logger() Logger {
	return c.logger
}

// IsDebug reports whether to enable the debug mode.
func (c *contextT) IsDebug() bool {
	return c.debug
}

// IsTLS returns true if HTTP connection is TLS otherwise false.
func (c *contextT) IsTLS() bool {
	return c.req.TLS != nil
}

// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
func (c *contextT) IsWebSocket() bool {
	if c.req.Method == http.MethodGet &&
		c.req.Header.Get(HeaderConnection) == "Upgrade" &&
		c.req.Header.Get(HeaderUpgrade) == "websocket" {
		return true
	}
	return false
}

func (c *contextT) IsAjax() bool {
	return c.req.Header.Get(HeaderXRequestedWith) == "XMLHttpRequest"
}

// Request returns the inner Request.
func (c *contextT) Request() *http.Request {
	return c.req
}

// Response returns the inner http.ResponseWriter.
func (c *contextT) Response() http.ResponseWriter {
	return responder{ResponseWriter: c.resp.ResponseWriter, ctx: c}
}

// SetResponse resets the response to resp, which will ignore nil.
func (c *contextT) SetResponse(resp http.ResponseWriter) {
	if resp != nil {
		c.resp.ResponseWriter = resp
	}
}

func (c *contextT) SetConnectionClose() {
	c.resp.Header().Set(HeaderConnection, "close")
}

func (c *contextT) GetBody() (string, error) {
	buf := c.AcquireBuffer()
	err := utils.ReadNWriter(buf, c.req.Body, c.req.ContentLength)
	body := buf.String()
	c.ReleaseBuffer(buf)
	return body, err
}

func (c *contextT) GetBodyReader() (*bytes.Buffer, error) {
	buf := c.AcquireBuffer()
	err := utils.ReadNWriter(buf, c.req.Body, c.req.ContentLength)
	if err != nil {
		c.ReleaseBuffer(buf)
		return nil, err
	}
	return buf, err
}

type acceptT struct {
	ct string
	q  float64
}

func (c *contextT) Accept() []string {
	accept := c.req.Header.Get(HeaderAccept)
	if accept == "" {
		return nil
	}

	ss := strings.Split(accept, ",")
	accepts := make([]acceptT, 0, len(ss))
	for _, s := range ss {
		q := 1.0
		if k := strings.IndexByte(s, ';'); k > 0 {
			qs := s[k+1:]
			s = s[:k]

			if j := strings.IndexByte(qs, '='); j > 0 {
				if qs = qs[j+1:]; qs == "" {
					continue
				}
				if v, _ := strconv.ParseFloat(qs, 32); v > 1.0 || v <= 0.0 {
					continue
				} else {
					q = v
				}
			} else {
				continue
			}
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		} else if s == "*/*" {
			s = ""
		} else if strings.HasSuffix(s, "/*") {
			s = s[:len(s)-1]
		}
		accepts = append(accepts, acceptT{ct: s, q: -q})
	}

	sort.SliceStable(accepts, func(i, j int) bool { return accepts[i].q < accepts[j].q })

	results := make([]string, len(accepts))
	for i := range accepts {
		results[i] = accepts[i].ct
	}
	return results
}

func (c *contextT) Host() string {
	return c.req.Host
}

func (c *contextT) Method() string {
	return c.req.Method
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *contextT) Scheme() (scheme string) {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}

	header := c.req.Header
	if scheme = header.Get(HeaderXForwardedProto); scheme != "" {
		return
	}
	if scheme = header.Get(HeaderXForwardedProtocol); scheme != "" {
		return
	}
	if scheme = header.Get(HeaderXUrlScheme); scheme != "" {
		return
	}
	if ssl := header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	return "http"
}

// RealIP returns the client's network address based on `X-Forwarded-For`
// or `X-Real-IP` request header.
func (c *contextT) RealIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
}

func (c *contextT) RemoteAddr() string {
	return c.req.RemoteAddr
}

func (c *contextT) RequestURI() string {
	return c.req.RequestURI
}

func (c *contextT) Charset() string {
	ct := c.req.Header.Get(HeaderContentType)
	if index := strings.IndexByte(ct, ';'); index > 0 {
		ct = ct[index:]
		if index = strings.IndexByte(ct, '='); index > 0 {
			return ct[index+1:]
		}
	}
	return ""
}

func (c *contextT) ContentType() (ct string) {
	ct = c.req.Header.Get(HeaderContentType)
	if index := strings.IndexAny(ct, " ;"); index > 0 {
		ct = ct[:index]
	}
	return
}

func (c *contextT) SetContentType(contentType string) {
	if contentType != "" {
		c.SetHeader(HeaderContentType, contentType)
	}
}

func (c *contextT) ContentLength() int64 {
	return c.req.ContentLength
}

func (c *contextT) Header(name string) string {
	return c.req.Header.Get(name)
}

func (c *contextT) SetHeader(name, value string) {
	c.resp.Header().Set(name, value)
}

// QueryParam returns the query param for the provided name.
func (c *contextT) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
}

// QueryParams returns the query parameters as `url.Values`.
func (c *contextT) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// QueryRawString returns the URL query string.
func (c *contextT) QueryRawString() string {
	return c.req.URL.RawQuery
}

// FormValue returns the form field value for the provided name.
func (c *contextT) FormValue(name string) string {
	return c.req.FormValue(name)
}

// FormParams returns the form parameters as `url.Values`.
func (c *contextT) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.req.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.req.ParseMultipartForm(MaxMemoryLimit); err != nil {
			return nil, err
		}
	} else {
		if err := c.req.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.req.Form, nil
}

// FormFile returns the multipart form file for the provided name.
func (c *contextT) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.req.FormFile(name)
	return fh, err
}

// MultipartForm returns the multipart form.
func (c *contextT) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(MaxMemoryLimit)
	return c.req.MultipartForm, err
}

// Cookie returns the named cookie provided in the request.
func (c *contextT) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *contextT) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *contextT) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.resp, cookie)
}

func (c *contextT) GetSession(id string) (v interface{}, err error) {
	if id == "" {
		return nil, ErrInvalidSession
	}
	if c.sessionK == id {
		switch c.sessionV {
		case nil, emptyValue:
			return nil, ErrSessionNotExist
		}
		return c.sessionV, nil
	}

	if c.session == nil {
		return nil, ErrNoSessionSupport
	}
	if v, err = c.session.GetSession(id); err == nil {
		c.sessionV = v
	} else {
		c.sessionV = emptyValue
	}
	c.sessionK = id
	return
}

func (c *contextT) SetSession(id string, value interface{}) (err error) {
	if c.session == nil {
		return ErrNoSessionSupport
	}
	if id == "" {
		return ErrInvalidSession
	}
	if value == nil {
		c.sessionK = id
		c.sessionV = emptyValue
		return nil
	}
	if err = c.session.SetSession(id, value); err != nil {
		return err
	}
	c.sessionK = id
	c.sessionV = value
	return nil
}

func (c *contextT) DelSession(id string) (err error) {
	if id == "" {
		return ErrInvalidSession
	}
	if c.session == nil {
		return ErrNoSessionSupport
	}
	if err = c.session.DelSession(id); err != nil {
		return
	}
	if c.sessionK == id {
		c.sessionV = nil
	}
	return nil
}

func (c *contextT) SetHandler(h func(Context, ...interface{}) error) {
	c.handler = h
}

func (c *contextT) Handle(args ...interface{}) error {
	if c.handler == nil {
		return ErrNoHandler
	}
	return c.handler(c, args...)
}

// Bind binds the request information into provided type v.
//
// The default binder does it based on Content-Type header.
func (c *contextT) Bind(v interface{}) error {
	return c.binder.Bind(c, v)
}

func (c *contextT) BindQuery(v interface{}) error {
	return c.binderQ(c.QueryParams(), v)
}

func (c *contextT) Write(b []byte) (int, error) {
	if !c.wrote {
		c.resp.WriteHeader(http.StatusOK)
	}
	return c.resp.Write(b)
}

// Render renders a template with data and sends a text/html response with status
// code. Renderer must be registered using `Echo.Renderer`.
func (c *contextT) Render(name string, code int, data interface{}) error {
	if c.renderer == nil {
		return ErrRendererNotRegistered
	}
	return c.renderer.Render(c, name, code, data)
}

// JSON sends a JSON response with status code.
func (c *contextT) JSON(code int, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

// JSONPretty sends a pretty-print JSON with status code.
func (c *contextT) JSONPretty(code int, i interface{}, indent string) error {
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

// JSONBlob sends a JSON blob response with status code.
func (c *contextT) JSONBlob(code int, b []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *contextT) JSONP(code int, callback string, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, b)
}

// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
// to construct the JSONP payload.
func (c *contextT) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.resp.Write(b); err != nil {
		return
	}
	_, err = c.resp.Write([]byte("):"))
	return
}

// XML sends an XML response with status code.
func (c *contextT) XML(code int, i interface{}) error {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLPretty sends a pretty-print XML with status code.
func (c *contextT) XMLPretty(code int, i interface{}, indent string) error {
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLBlob sends an XML blob response with status code.
func (c *contextT) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

// HTML sends an HTTP response with status code.
func (c *contextT) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

// HTMLBlob sends an HTTP blob response with status code.
func (c *contextT) HTMLBlob(code int, b []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

// String sends a string response with status code.
func (c *contextT) String(code int, s string) error {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *contextT) writeContentType(value string) {
	if value == "" {
		return
	}
	header := c.resp.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

// Blob sends a blob response with status code and content type.
func (c *contextT) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.Write(b)
	return
}

// Stream sends a streaming response with status code and content type.
func (c *contextT) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = io.Copy(c.resp, r)
	return
}

// File sends a response with the content of the file.
func (c *contextT) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	if fi, _ := f.Stat(); fi != nil {
		if fi.IsDir() {
			file = filepath.Join(file, indexPage)
			if f, err = os.Open(file); err != nil {
				return ErrNotFound
			}
			defer f.Close()
			if fi, err = f.Stat(); err != nil {
				return
			}
		}
		http.ServeContent(c.resp, c.req, fi.Name(), fi.ModTime(), f)
	}

	return
}

func (c *contextT) contentDisposition(file, name, dispositionType string) error {
	c.resp.Header().Set(HeaderContentDisposition,
		fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

// Attachment sends a response as attachment, prompting client to save the
// file.
func (c *contextT) Attachment(file string, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

// Inline sends a response as inline, opening the file in the browser.
func (c *contextT) Inline(file string, name string) error {
	return c.contentDisposition(file, name, "inline")
}

// NoContent sends a response with no body and a status code.
func (c *contextT) NoContent(code int) error {
	c.resp.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *contextT) Redirect(code int, toURL string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, toURL)
	c.resp.WriteHeader(code)
	return nil
}
