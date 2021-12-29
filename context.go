// Copyright 2020 xgfone
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
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type reqctx uint8

// GetContext returns the http reqeust context from the context.
func GetContext(ctx context.Context) *Context {
	c, _ := ctx.Value(reqctx(255)).(*Context)
	return c
}

// SetContext sets the http request context into the context.
func SetContext(ctx context.Context, c *Context) (newctx context.Context) {
	return context.WithValue(ctx, reqctx(255), c)
}

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

// BufferAllocator is used to acquire and release a buffer.
type BufferAllocator interface {
	AcquireBuffer() *bytes.Buffer
	ReleaseBuffer(*bytes.Buffer)
}

// Context represetns a request and response context.
type Context struct {
	// Route is the route information associated with the route.
	Route Route

	// Any is the any context data associated with the route.
	//
	// Notice: when the new request is coming, they will be reset to nil.
	Any interface{}

	// Data is used to store many key-value pairs about the context.
	//
	// Notice: when the new request is coming, they will be cleaned out.
	Data map[string]interface{}

	// Public Configuration, which are not reset when calling Reset().
	BufferAllocator
	Logger
	Router      Router
	Session     Session
	NotFound    Handler
	Binder      Binder
	Renderer    Renderer
	Defaulter   Defaulter
	Validator   Validator
	Responder   func(*Context, ...interface{}) error
	QueryBinder func(interface{}, url.Values) error

	res *Response
	req *http.Request

	plen    int
	pnames  []string
	pvalues []string
	cookies []*http.Cookie
	query   url.Values
}

// NewContext returns a new Context.
func NewContext(urlParamMaxNum, dataInitCap int) *Context {
	var pnames, pvalues []string
	if urlParamMaxNum > 0 {
		pnames = make([]string, urlParamMaxNum)
		pvalues = make([]string, urlParamMaxNum)
	}

	return &Context{
		res:     NewResponse(nil),
		Data:    make(map[string]interface{}, dataInitCap),
		pnames:  pnames,
		pvalues: pvalues,
	}
}

// Reset resets the context to the initalizing state.
func (c *Context) Reset() {
	c.Any = nil
	c.Route = Route{}
	if len(c.Data) != 0 {
		for key := range c.Data {
			delete(c.Data, key)
		}
	}

	c.req = nil
	c.res.Reset(nil)
	c.cookies = nil
	c.query = nil
	c.plen = 0
}

// URL generates a url path by the route path name and provided parameters.
//
// Return "" if there is not the route named name.
func (c *Context) URL(name string, params ...interface{}) string {
	return c.Router.Path(name, params...)
}

// FindRoute finds the route by the request method and path and put it
// into the field Route of Context.
//
// For the handler registered into the underlying Router, it supports
// three kinds of types as follow:
//
//   - Route
//   - Handler
//   - http.Handler
//   - http.HandlerFunc
//
func (c *Context) FindRoute() (ok bool) {
	h, n := c.Router.Match(c.req.URL.Path, c.req.Method, c.pnames, c.pvalues)
	if h == nil {
		return false
	}

	c.plen = n
	switch r := h.(type) {
	case Route:
		c.Route = r
	case Handler:
		c.Route.Handler = r
	case http.Handler:
		c.Route.Handler = FromHTTPHandler(r)
	case http.HandlerFunc:
		c.Route.Handler = FromHTTPHandlerFunc(r)
	default:
		panic(fmt.Errorf("unknown handler type '%T'", h))
	}

	return true
}

// ExecuteRoute executes the handler of the found route.
//
// Notice: You should call FindRoute before calling this method.
func (c *Context) ExecuteRoute() error {
	if c.Route.Handler != nil {
		return c.Route.Handler(c)
	}
	return c.NotFound(c)
}

// Execute finds the route by the request method and path, then executes
// the handler of the found route, which is equal to the union of FindRoute
// and ExecuteRoute.
func (c *Context) Execute() error {
	h, n := c.Router.Match(c.req.URL.Path, c.req.Method, c.pnames, c.pvalues)
	if h == nil {
		return c.NotFound(c)
	}

	c.plen = n
	switch r := h.(type) {
	case Route:
		c.Route = r
	case Handler:
		c.Route.Handler = r
	case http.Handler:
		c.Route.Handler = FromHTTPHandler(r)
	case http.HandlerFunc:
		c.Route.Handler = FromHTTPHandlerFunc(r)
	default:
		panic(fmt.Errorf("unknown handler type '%T'", h))
	}

	return c.Route.Handler(c)
}

//----------------------------------------------------------------------------
// Request & Response
//----------------------------------------------------------------------------

// SetReqResp is the same as Reset, but only reset the request and response,
// not all things.
func (c *Context) SetReqResp(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.res.SetWriter(w)
}

// SetRequest resets the request to req.
func (c *Context) SetRequest(req *http.Request) { c.req = req }

// SetResponse resets the response to resp, which will ignore nil.
func (c *Context) SetResponse(resp http.ResponseWriter) { c.res.SetWriter(resp) }

// Request returns the inner Request.
func (c *Context) Request() *http.Request { return c.req }

// Response returns the inner Response.
func (c *Context) Response() *Response { return c.res }

// ResponseWriter returns the underlying http.ResponseWriter.
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.res.ResponseWriter
}

// StatusCode returns the status code of the response.
func (c *Context) StatusCode() int { return c.res.Status }

// IsResponded reports whether the response is sent or not.
func (c *Context) IsResponded() bool { return c.res.Wrote }

// Respond responds the result to the peer by using Ship.Responder.
func (c *Context) Respond(args ...interface{}) error {
	return c.Responder(c, args...)
}

//----------------------------------------------------------------------------
// Request Information
//----------------------------------------------------------------------------

// Body returns the reader of the request body.
func (c *Context) Body() io.ReadCloser { return c.req.Body }

// IsTLS reports whether HTTP connection is TLS or not.
func (c *Context) IsTLS() bool { return c.req.TLS != nil }

// IsAjax reports whether the request is ajax or not.
func (c *Context) IsAjax() bool {
	return c.req.Header.Get(HeaderXRequestedWith) == "XMLHttpRequest"
}

// IsWebSocket reports whether HTTP connection is WebSocket or not.
func (c *Context) IsWebSocket() bool {
	if c.req.Method == http.MethodGet &&
		c.req.Header.Get(HeaderConnection) == "Upgrade" &&
		c.req.Header.Get(HeaderUpgrade) == "websocket" {
		return true
	}
	return false
}

// Host returns the host of the request.
func (c *Context) Host() string { return c.req.Host }

// Hostname returns the hostname of the request.
func (c *Context) Hostname() string { return c.req.URL.Hostname() }

// Method returns the url method of the request.
func (c *Context) Method() string { return c.req.Method }

// Path returns the url path of the request.
func (c *Context) Path() string { return c.req.URL.Path }

// Referer returns the header "Referer" of the request.
func (c *Context) Referer() string { return c.req.Referer() }

// UserAgent returns the header "User-Agent" of the request.
func (c *Context) UserAgent() string { return c.req.UserAgent() }

// RemoteAddr returns the remote address of the http connection.
func (c *Context) RemoteAddr() string { return c.req.RemoteAddr }

// RequestURI returns the URI of the request.
func (c *Context) RequestURI() string { return c.req.RequestURI }

// ContentLength return the length of the request body.
func (c *Context) ContentLength() int64 { return c.req.ContentLength }

// Accept returns the accepted Content-Type list from the request header "Accept",
// which are sorted by the q-factor weight from high to low.
//
// If there is no the request header "Accept", return nil.
//
// Notice:
//   1. If the value is "*/*", it will be amended as "".
//   2. If the value is "<MIME_type>/*", it will be amended as "<MIME_type>/".
//      So it can be used to match the prefix.
//
func (c *Context) Accept() []string {
	type acceptT struct {
		ct string
		q  float64
	}

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

	sort.SliceStable(accepts, func(i, j int) bool {
		return accepts[i].q < accepts[j].q
	})

	results := make([]string, len(accepts))
	for i := range accepts {
		results[i] = accepts[i].ct
	}
	return results
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *Context) Scheme() (scheme string) {
	header := c.req.Header

	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() || header.Get(HeaderXForwardedSSL) == "on" {
		return "https"
	} else if scheme = header.Get(HeaderXForwardedProto); scheme != "" {
		return
	} else if scheme = header.Get(HeaderXForwardedProtocol); scheme != "" {
		return
	} else if scheme = header.Get(HeaderXUrlScheme); scheme != "" {
		return
	}

	return "http"
}

// ClientIP returns the real client's network address based on `X-Forwarded-For`
// or `X-Real-Ip` request header. Or returns the remote address.
func (c *Context) ClientIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	} else if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	} else if ra, _, _ := net.SplitHostPort(c.req.RemoteAddr); ra != "" {
		return ra
	}
	return c.req.RemoteAddr
}

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
func (c *Context) Charset() string {
	ct := c.req.Header.Get(HeaderContentType)
	index := strings.IndexByte(ct, ';')
	for ; index > 0; index = strings.IndexByte(ct, ';') {
		ct = ct[index+1:]
		if index = strings.IndexByte(ct, '='); index > 0 {
			if strings.TrimSpace(ct[:index]) == "charset" {
				return strings.TrimSpace(ct[index+1:])
			}
			ct = ct[index+1:]
		}
	}
	return ""
}

// ContentType returns the Content-Type of the request without the charset.
func (c *Context) ContentType() (ct string) {
	ct = c.req.Header.Get(HeaderContentType)
	if index := strings.IndexAny(ct, ";"); index > 0 {
		ct = strings.TrimSpace(ct[:index])
	}
	return
}

//----------------------------------------------------------------------------
// URL Params
//----------------------------------------------------------------------------

// Param returns the parameter value in the url path by name.
func (c *Context) Param(name string) string {
	for i := 0; i < c.plen; i++ {
		if c.pnames[i] == name {
			return c.pvalues[i]
		}
	}
	return ""
}

// Params returns all the parameters as the key-value map in the url path.
func (c *Context) Params() map[string]string {
	ms := make(map[string]string, c.plen)
	for i := 0; i < c.plen; i++ {
		ms[c.pnames[i]] = c.pvalues[i]
	}
	return ms
}

// ParamNames returns the names of all the URL parameters.
func (c *Context) ParamNames() []string { return c.pnames[:c.plen] }

// ParamValues returns the values of all the URL parameters.
func (c *Context) ParamValues() []string { return c.pvalues[:c.plen] }

//----------------------------------------------------------------------------
// Header
//----------------------------------------------------------------------------

// Header implements the interface http.ResponseWriter,
// which is the alias of RespHeader.
func (c *Context) Header() http.Header { return c.res.Header() }

// ReqHeader returns the header of the request.
func (c *Context) ReqHeader() http.Header { return c.req.Header }

// RespHeader returns the header of the response.
func (c *Context) RespHeader() http.Header { return c.res.Header() }

// GetReqHeader returns the first value of the request header named name.
//
// Return defaultValue instead if the header does not exist.
func (c *Context) GetReqHeader(name string, defaultValue ...string) string {
	if vs, ok := c.req.Header[textproto.CanonicalMIMEHeaderKey(name)]; ok {
		return vs[0]
	} else if len(defaultValue) != 0 {
		return defaultValue[0]
	}

	return ""
}

// SetRespHeader sets the response header name to value.
func (c *Context) SetRespHeader(name, value string) {
	c.res.Header().Set(name, value)
}

// AddRespHeader appends the value into the response header named name.
func (c *Context) AddRespHeader(name, value string) {
	c.res.Header().Add(name, value)
}

// DelRespHeader deletes the response header named name.
func (c *Context) DelRespHeader(name string) {
	c.res.Header().Del(name)
}

//----------------------------------------------------------------------------
// Cookie
//----------------------------------------------------------------------------

// Cookies returns the HTTP cookies sent with the request.
func (c *Context) Cookies() []*http.Cookie {
	if c.cookies == nil {
		c.cookies = c.req.Cookies()
	}
	return c.cookies
}

// Cookie returns the named cookie provided in the request.
//
// Return nil if no the cookie named name.
func (c *Context) Cookie(name string) *http.Cookie {
	for _, cookie := range c.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// SetCookie appends a http cookie to the response header `Set-Cookie`.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.res, cookie)
}

//----------------------------------------------------------------------------
// Request Query
//----------------------------------------------------------------------------

// Query returns the query value by the query name.
//
// Return defaultValue instead if the query name does not exist.
func (c *Context) Query(name string, defaultValue ...string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}

	if values := c.query[name]; len(values) != 0 {
		return values[0]
	} else if len(defaultValue) != 0 {
		return defaultValue[0]
	}

	return ""
}

// Queries returns all the query values.
func (c *Context) Queries() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// QueryRawString returns the URL query string.
func (c *Context) QueryRawString() string {
	return c.req.URL.RawQuery
}

//----------------------------------------------------------------------------
// Request Form
//----------------------------------------------------------------------------

// Form returns the form value by the field name.
//
// Return defaultValue instead if the form field name does not exist.
func (c *Context) Form(name string, defaultValue ...string) string {
	if c.req.Form == nil {
		c.req.ParseMultipartForm(MaxMemoryLimit)
	}

	if values, ok := c.req.Form[name]; ok {
		return values[0]
	} else if len(defaultValue) != 0 {
		return defaultValue[0]
	}

	return ""
}

// Forms returns all the form values.
func (c *Context) Forms() (url.Values, error) {
	if strings.HasPrefix(c.req.Header.Get("Content-Type"), MIMEMultipartForm) {
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

// FormFile returns the multipart form file by the field name.
func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.req.FormFile(name)
}

// MultipartForm returns the multipart form.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(MaxMemoryLimit)
	return c.req.MultipartForm, err
}

// MultipartReader returns the multipart reader from the request.
func (c *Context) MultipartReader() (*multipart.Reader, error) {
	return c.req.MultipartReader()
}

//----------------------------------------------------------------------------
// Session Management
//----------------------------------------------------------------------------

// GetSession returns the session content by id from the backend store.
//
// If the session id does not exist, it returns ErrSessionNotExist.
func (c *Context) GetSession(id string) (v interface{}, err error) {
	if id == "" {
		err = ErrInvalidSession
	} else if v, err = c.Session.GetSession(id); err == nil && v == nil {
		err = ErrSessionNotExist
	}

	return
}

// SetSession sets the session to the backend store.
func (c *Context) SetSession(id string, value interface{}) (err error) {
	if id == "" || value == nil {
		return ErrInvalidSession
	}
	return c.Session.SetSession(id, value)
}

// DelSession deletes the session from the backend store.
func (c *Context) DelSession(id string) (err error) {
	if id == "" {
		return ErrInvalidSession
	}
	return c.Session.DelSession(id)
}

//----------------------------------------------------------------------------
// Bind & SetDefault & Validator
//----------------------------------------------------------------------------

// Bind extracts the data information from the request and assigns it to v,
// then validates whether it is valid or not.
func (c *Context) Bind(v interface{}) (err error) {
	if err = c.Binder.Bind(v, c.req); err == nil {
		if err = c.Defaulter.SetDefault(v); err == nil {
			err = c.Validator.Validate(v)
		}
	}
	return
}

// BindQuery extracts the data from the request url query and assigns it to v,
// then validates whether it is valid or not.
func (c *Context) BindQuery(v interface{}) (err error) {
	if err = c.QueryBinder(v, c.Queries()); err == nil {
		if err = c.Defaulter.SetDefault(v); err == nil {
			err = c.Validator.Validate(v)
		}
	}
	return
}

//----------------------------------------------------------------------------
// Renderer
//----------------------------------------------------------------------------

// Render renders a template named name with data and sends it as the response
// with status code.
func (c *Context) Render(name string, code int, data interface{}) error {
	return c.Renderer.Render(c, name, code, data)
}

// RenderOk is short for c.Render(name, http.StatusOK, data).
func (c *Context) RenderOk(name string, data interface{}) error {
	return c.Render(name, http.StatusOK, data)
}

//----------------------------------------------------------------------------
// Set Repsonse
//----------------------------------------------------------------------------

// SetConnectionClose sets the response header "Connection: close"
// to tell the server to close the connection.
func (c *Context) SetConnectionClose() {
	c.res.Header().Set(HeaderConnection, "close")
}

// SetContentType sets the response header "Content-Type" to ct,
//
// If ct is "", do nothing.
func (c *Context) SetContentType(ct string) {
	SetContentType(c.res.Header(), ct)
}

//----------------------------------------------------------------------------
// Send Repsonse
//----------------------------------------------------------------------------

// Error returns a http error with the status code.
func (c *Context) Error(code int, err error) HTTPServerError {
	return HTTPServerError{Code: code, Err: err}
}

// WriteHeader implements the interface http.ResponseWriter.
func (c *Context) WriteHeader(statusCode int) {
	c.res.WriteHeader(statusCode)
}

// Write implements the interface http.ResponseWriter.
//
// It will write the response header with the status code 200 firstly
// if the header is not sent.
func (c *Context) Write(b []byte) (int, error) {
	return c.res.Write(b)
}

// NoContent sends a response with the status code and without the body.
func (c *Context) NoContent(code int) error {
	c.res.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, toURL string) error {
	if code < 300 || code >= 400 {
		return ErrInvalidRedirectCode
	}
	c.res.Header().Set(HeaderLocation, toURL)
	return c.NoContent(code)
}

func (c *Context) setContentTypeAndCode(code int, ct string) {
	c.SetContentType(ct)
	c.res.WriteHeader(code)
}

// Stream sends a streaming response with the status code and the content type.
func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.setContentTypeAndCode(code, contentType)
	_, err = io.CopyBuffer(c.res, r, make([]byte, 2048))
	return
}

// Blob sends a blob response with the status code and the content type.
func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.setContentTypeAndCode(code, contentType)
	_, err = c.res.Write(b)
	return
}

// BlobText sends a string blob response with the status code and the content type.
func (c *Context) BlobText(code int, contentType string,
	format string, args ...interface{}) (err error) {
	c.setContentTypeAndCode(code, contentType)
	if len(args) > 0 {
		_, err = fmt.Fprintf(c.res, format, args...)
	} else {
		_, err = c.res.WriteString(format)
	}
	return err
}

// BlobXML sends an XML blob response with the status code.
func (c *Context) BlobXML(code int, b []byte) (err error) {
	c.setContentTypeAndCode(code, MIMEApplicationXMLCharsetUTF8)
	if _, err = c.res.WriteString(xml.Header); err != nil {
		return
	}
	_, err = c.res.Write(b)
	return
}

// XML sends an XML response with the status code.
func (c *Context) XML(code int, v interface{}) (err error) {
	buf := c.AcquireBuffer()
	buf.WriteString(xml.Header)
	if err = xml.NewEncoder(buf).Encode(v); err == nil {
		c.setContentTypeAndCode(code, MIMEApplicationXMLCharsetUTF8)
		_, err = c.res.Write(buf.Bytes())
	}
	c.ReleaseBuffer(buf)
	return
}

// JSON sends a JSON response with the status code.
func (c *Context) JSON(code int, v interface{}) (err error) {
	buf := c.AcquireBuffer()
	if err = json.NewEncoder(buf).Encode(v); err == nil {
		c.setContentTypeAndCode(code, MIMEApplicationJSONCharsetUTF8)
		_, err = c.res.Write(buf.Bytes())
	}
	c.ReleaseBuffer(buf)
	return
}

// HTML sends an HTML response with the status code.
func (c *Context) HTML(code int, htmlfmt string, htmlargs ...interface{}) error {
	return c.BlobText(code, MIMETextHTMLCharsetUTF8, htmlfmt, htmlargs...)
}

// Text sends a string response with the status code.
func (c *Context) Text(code int, format string, args ...interface{}) error {
	return c.BlobText(code, MIMETextPlainCharsetUTF8, format, args...)
}

// File sends a file response, and the body is the content of the file.
//
// If not set the Content-Type, it will deduce it from the extension
// of the file name. If the file does not exist, it returns ErrNotFound.
func (c *Context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return ErrInternalServerError.New(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return ErrInternalServerError.New(err)
	} else if fi.IsDir() {
		f, err := os.Open(filepath.Join(file, "index.html"))
		if err != nil {
			if os.IsNotExist(err) {
				return ErrNotFound
			}
			return ErrInternalServerError.New(err)
		}
		defer f.Close()

		if fi, err = f.Stat(); err != nil {
			return ErrInternalServerError.New(err)
		}

		http.ServeContent(c.res.ResponseWriter, c.req, fi.Name(), fi.ModTime(), f)
	} else {
		http.ServeContent(c.res.ResponseWriter, c.req, fi.Name(), fi.ModTime(), f)
	}

	return
}

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	if name == "" {
		name = filepath.Base(file)
	}

	disposition := fmt.Sprintf("%s; filename=%q", dispositionType, name)
	c.res.Header().Set(HeaderContentDisposition, disposition)
	return c.File(file)
}

// Attachment is the same as File, but sets the header "Content-Disposition"
// with the type "attachment" to prompt the client to save the file with the name.
//
// If the file does not exist, it returns ErrNotFound.
func (c *Context) Attachment(file string, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

// Inline sends a file response as the inline to open the file in the browser.
//
// If the file does not exist, it returns ErrNotFound.
func (c *Context) Inline(file string, name string) error {
	return c.contentDisposition(file, name, "inline")
}
