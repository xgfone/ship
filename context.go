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
)

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

var indexPage = "index.html"

type contextT struct {
	req  *http.Request
	resp http.ResponseWriter

	router Router
	params URLParam
	query  url.Values

	debug  bool
	logger Logger

	binder   Binder
	renderer Renderer

	store map[string]interface{}
}

// NewContext returns a new context.
//
// Notice: For the method,
//
//     SetRouter(router Router)
//
// if router is the default implementation returned by NewRouter(),
// it will also call the following method：
//
//     c.SetDebug()
//     c.SetLogger()
//     c.SetBinder()
//     c.SetRenderer()
//
// the arguments of which comes from router.
func NewContext() Context {
	return &contextT{store: make(map[string]interface{})}
}

func (c *contextT) writeContentType(value string) {
	header := c.resp.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *contextT) Router() Router {
	return c.router
}

func (c *contextT) SetRouter(router Router) {
	if router == nil {
		panic(fmt.Errorf("router must not be nil"))
	}

	if r, ok := router.(*routerT); ok {
		c.SetDebug(r.config.Debug)
		c.SetLogger(r.config.Logger)
		c.SetBinder(r.config.Binder)
		c.SetRenderer(r.config.Renderer)
	}
}

func (c *contextT) Request() *http.Request {
	return c.req
}

func (c *contextT) Response() http.ResponseWriter {
	return c.resp
}

func (c *contextT) SetRequest(req *http.Request) {
	c.req = req
}

func (c *contextT) SetReqResp(req *http.Request, resp http.ResponseWriter) {
	c.req = req
	c.resp = resp
}

func (c *contextT) IsDebug() bool {
	return c.debug
}

func (c *contextT) SetDebug(debug bool) {
	c.debug = debug
}

func (c *contextT) Logger() Logger {
	return c.logger
}

func (c *contextT) SetLogger(logger Logger) {
	c.logger = logger
}

// IsTLS returns true if HTTP connection is TLS otherwise false.
func (c *contextT) IsTLS() bool {
	return c.req.TLS != nil

}

// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
func (c *contextT) IsWebSocket() bool {
	return strings.ToLower(c.req.Header.Get(HeaderUpgrade)) == "websocket"
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *contextT) Scheme() (scheme string) {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	if scheme = c.req.Header.Get(HeaderXForwardedProto); scheme != "" {
		return
	}
	if scheme = c.req.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return
	}
	if scheme = c.req.Header.Get(HeaderXUrlScheme); scheme != "" {
		return
	}
	if ssl := c.req.Header.Get(HeaderXForwardedSsl); ssl == "on" {
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

// Param returns path parameter by name.
func (c *contextT) URLParamByName(name string) string {
	if c.params == nil {
		return ""
	}
	return c.params.Get(name)
}

// ParamNames returns path parameter names.
func (c *contextT) URLParam() URLParam {
	return c.params
}

// SetParamNames sets path parameter names.
func (c *contextT) SetURLParam(params URLParam) {
	c.params = params
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

// QueryString returns the URL query string.
func (c *contextT) QueryString() string {
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

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *contextT) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.resp, cookie)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *contextT) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

// Get retrieves data from the context.
func (c *contextT) Get(key string) interface{} {
	return c.store[key]
}

// Set saves data in the context.
func (c *contextT) Set(key string, value interface{}) {
	c.store[key] = value
}

// Bind binds the request body into provided type `i`. The default binder
// does it based on Content-Type header.
func (c *contextT) Bind(v interface{}) error {
	return c.binder.Bind(c, v)
}

func (c *contextT) SetBinder(b Binder) {
	c.binder = b
}

// Render renders a template with data and sends a text/html response with status
// code. Renderer must be registered using `Echo.Renderer`.
func (c *contextT) Render(code int, name string, data interface{}) error {
	if c.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err := c.renderer.Render(c, buf, code, name, data); err != nil {
		return err
	}
	return c.HTMLBlob(code, buf.Bytes())
}

// SetRenderer sets the rendered to r.
func (c *contextT) SetRenderer(r Renderer) {
	c.renderer = r
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
	c.resp.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
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

func (c *contextT) URL(name string, params URLParam) string {
	return c.router.URL(name, params)
}

func (c *contextT) Reset() {
	c.req = nil
	c.resp = nil
	c.router = nil
	c.params = nil
	c.query = nil
	c.debug = false
	c.logger = nil
	c.binder = nil
	c.renderer = nil
}
