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
	"bytes"
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
	"strings"

	"github.com/xgfone/ship/core"
)

var (
	indexPage = "index.html"
	emptyStrS = [256]string{}
)

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

// Context is the alias of core.Context, which stands for a context.
//
// Methods:
//
//    NotFoundHandler() Handler
//
//    Request() *http.Request
//    Response() http.ResponseWriter
//    SetResponse(http.ResponseWriter)
//
//    IsTLS() bool
//    IsDebug() bool
//    IsWebSocket() bool
//
//    Header(name string) (value string)
//    SetHeader(name string, value string)
//
//    URLParams() map[string]string
//    URLParamByName(name string) (value string)
//
//    Scheme() string
//    RealIP() string
//    ContentType() string
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
//    // Get and Set are used to store the key-value information about the context.
//    Get(key string) (value interface{})
//    Set(key string, value interface{})
//
//    Logger() Logger
//    URL(name string, params ...interface{}) string
//
//    Bind(v interface{}) error
//
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

// Context stands for a request and response context.
type context struct {
	req   *http.Request
	resp  http.ResponseWriter
	query url.Values

	pnames  []string
	pvalues []string

	ship     *Ship
	debug    bool
	logger   Logger
	binder   Binder
	renderer Renderer

	store map[string]interface{}
}

// NewContext returns a new context.
func newContext(s *Ship, req *http.Request, resp http.ResponseWriter, maxParam int) *context {
	var pnames, pvalues []string
	if maxParam > 0 {
		pnames = make([]string, maxParam)
		pvalues = make([]string, maxParam)
	}

	return &context{
		req:  req,
		resp: resp,

		// See c.setShip(s)
		ship:     s,
		debug:    s.config.Debug,
		logger:   s.config.Logger,
		binder:   s.config.Binder,
		renderer: s.config.Renderer,

		pnames:  pnames,
		pvalues: pvalues,

		store: make(map[string]interface{}),
	}
}

func (c *context) setShip(s *Ship) {
	c.ship = s
	c.debug = s.config.Debug
	c.logger = s.config.Logger
	c.binder = s.config.Binder
	c.renderer = s.config.Renderer
}

func (c *context) setReqResp(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.resp = w
}

func (c *context) reset() {
	c.req = nil
	c.resp = nil
	c.query = nil

	copy(c.pnames, emptyStrS[:len(c.pnames)])
	copy(c.pvalues, emptyStrS[:len(c.pvalues)])

	for key := range c.store {
		delete(c.store, key)
	}
}

func (c *context) NotFoundHandler() Handler {
	return c.ship.config.NotFoundHandler
}

// URLParamByName returns the parameter in the url path by name.
func (c *context) URLParamByName(name string) string {
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

// URLParams returns all the parameters in the url path.
func (c *context) URLParams() map[string]string {
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

// Get retrieves data from the context.
func (c *context) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Logger returns the logger implementation.
func (c *context) Logger() Logger {
	return c.logger
}

// IsDebug reports whether to enable the debug mode.
func (c *context) IsDebug() bool {
	return c.debug
}

// IsTLS returns true if HTTP connection is TLS otherwise false.
func (c *context) IsTLS() bool {
	return c.req.TLS != nil
}

// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
func (c *context) IsWebSocket() bool {
	return strings.ToLower(c.req.Header.Get(HeaderUpgrade)) == "websocket"
}

// Request returns the inner Request.
func (c *context) Request() *http.Request {
	return c.req
}

// Response returns the inner http.ResponseWriter.
func (c *context) Response() http.ResponseWriter {
	return c.resp
}

// SetResponse resets the response to resp, which will ignore nil.
func (c *context) SetResponse(resp http.ResponseWriter) {
	if resp != nil {
		c.resp = resp
	}
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *context) Scheme() (scheme string) {
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
func (c *context) RealIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
}

func (c *context) ContentType() (ct string) {
	ct = c.req.Header.Get("Content-Type")
	if index := strings.IndexAny(ct, " ;"); index > 0 {
		ct = ct[:index]
	}
	return
}

func (c *context) Header(name string) string {
	return c.req.Header.Get(name)
}

func (c *context) SetHeader(name, value string) {
	c.resp.Header().Set(name, value)
}

// QueryParam returns the query param for the provided name.
func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
}

// QueryParams returns the query parameters as `url.Values`.
func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// QueryRawString returns the URL query string.
func (c *context) QueryRawString() string {
	return c.req.URL.RawQuery
}

// FormValue returns the form field value for the provided name.
func (c *context) FormValue(name string) string {
	return c.req.FormValue(name)
}

// FormParams returns the form parameters as `url.Values`.
func (c *context) FormParams() (url.Values, error) {
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
func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.req.FormFile(name)
	return fh, err
}

// MultipartForm returns the multipart form.
func (c *context) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(MaxMemoryLimit)
	return c.req.MultipartForm, err
}

// Cookie returns the named cookie provided in the request.
func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *context) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.resp, cookie)
}

// Bind binds the request information into provided type v.
//
// The default binder does it based on Content-Type header.
func (c *context) Bind(v interface{}) error {
	return c.binder.Bind(c, v)
}

// Render renders a template with data and sends a text/html response with status
// code. Renderer must be registered using `Echo.Renderer`.
func (c *context) Render(name string, code int, data interface{}) error {
	if c.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err := c.renderer.Render(c, buf, name, code, data); err != nil {
		return err
	}
	return c.HTMLBlob(code, buf.Bytes())
}

func (c *context) writeContentType(value string) {
	header := c.resp.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

// HTML sends an HTTP response with status code.
func (c *context) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

// HTMLBlob sends an HTTP blob response with status code.
func (c *context) HTMLBlob(code int, b []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

// String sends a string response with status code.
func (c *context) String(code int, s string) error {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

// JSON sends a JSON response with status code.
func (c *context) JSON(code int, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

// JSONPretty sends a pretty-print JSON with status code.
func (c *context) JSONPretty(code int, i interface{}, indent string) error {
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

// JSONBlob sends a JSON blob response with status code.
func (c *context) JSONBlob(code int, b []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *context) JSONP(code int, callback string, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, b)
}

// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
// to construct the JSONP payload.
func (c *context) JSONPBlob(code int, callback string, b []byte) (err error) {
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
func (c *context) XML(code int, i interface{}) error {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLPretty sends a pretty-print XML with status code.
func (c *context) XMLPretty(code int, i interface{}, indent string) error {
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLBlob sends an XML blob response with status code.
func (c *context) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

// Blob sends a blob response with status code and content type.
func (c *context) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.Write(b)
	return
}

// Stream sends a streaming response with status code and content type.
func (c *context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = io.Copy(c.resp, r)
	return
}

// File sends a response with the content of the file.
func (c *context) File(file string) (err error) {
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

func (c *context) contentDisposition(file, name, dispositionType string) error {
	c.resp.Header().Set(HeaderContentDisposition,
		fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

// Attachment sends a response as attachment, prompting client to save the
// file.
func (c *context) Attachment(file string, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

// Inline sends a response as inline, opening the file in the browser.
func (c *context) Inline(file string, name string) error {
	return c.contentDisposition(file, name, "inline")
}

// NoContent sends a response with no body and a status code.
func (c *context) NoContent(code int) error {
	c.resp.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *context) Redirect(code int, toURL string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, toURL)
	c.resp.WriteHeader(code)
	return nil
}

// URL generates an URL from route name and provided parameters.
func (c *context) URL(name string, params ...interface{}) string {
	return c.ship.URL(name, params...)
}
