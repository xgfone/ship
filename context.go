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

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

var contenttypes = map[string][]string{}

// AddContentTypeMapping add a content type mapping to convert contentType
// to contentTypeSlice, then call SetContentType to set the Content-Type
// to contentTypeSlice by contentType to avoid to allocate the memory.
func AddContentTypeMapping(contentType string, contentTypeSlice []string) {
	if contentType == "" {
		panic(fmt.Errorf("the Content-Type is empty"))
	}
	if len(contentTypeSlice) == 0 {
		panic(fmt.Errorf("the Content-Type slice is empty"))
	}
	contenttypes[contentType] = contentTypeSlice
}

// SetHeaderContentType sets the Content-Type header to ct.
func SetHeaderContentType(header http.Header, ct string) {
	var cts []string
	switch ct {
	case MIMEApplicationJSON:
		cts = MIMEApplicationJSONs
	case MIMEApplicationJSONCharsetUTF8:
		cts = MIMEApplicationJSONCharsetUTF8s
	case MIMEApplicationJavaScript:
		cts = MIMEApplicationJavaScripts
	case MIMEApplicationJavaScriptCharsetUTF8:
		cts = MIMEApplicationJavaScriptCharsetUTF8s
	case MIMEApplicationXML:
		cts = MIMEApplicationXMLs
	case MIMEApplicationXMLCharsetUTF8:
		cts = MIMEApplicationXMLCharsetUTF8s
	case MIMETextXML:
		cts = MIMETextXMLs
	case MIMETextXMLCharsetUTF8:
		cts = MIMETextXMLCharsetUTF8s
	case MIMEApplicationForm:
		cts = MIMEApplicationForms
	case MIMEApplicationProtobuf:
		cts = MIMEApplicationProtobufs
	case MIMEApplicationMsgpack:
		cts = MIMEApplicationMsgpacks
	case MIMETextHTML:
		cts = MIMETextHTMLs
	case MIMETextHTMLCharsetUTF8:
		cts = MIMETextHTMLCharsetUTF8s
	case MIMETextPlain:
		cts = MIMETextPlains
	case MIMETextPlainCharsetUTF8:
		cts = MIMETextPlainCharsetUTF8s
	case MIMEMultipartForm:
		cts = MIMEMultipartForms
	case MIMEOctetStream:
		cts = MIMEOctetStreams
	default:
		if ss := contenttypes[ct]; ss != nil {
			cts = ss
		} else {
			header.Set(HeaderContentType, ct)
			return
		}
	}
	header[HeaderContentType] = cts
}

// SetContentType is equal to SetHeaderContentType(res.Header(), ct).
func SetContentType(res http.ResponseWriter, ct string) {
	SetHeaderContentType(res.Header(), ct)
}

// Context represetns a request and response context.
type Context struct {
	// RouteInfo is the route information associated with the route.
	RouteInfo RouteInfo

	// Data is used to store many key-value pairs about the context.
	//
	// Data maybe asks the system to allocate many memories.
	// If the interim context value is too few and you don't want the system
	// to allocate many memories, the three context variables is for you
	// and you can consider them as the context register to use.
	//
	// Notice: when the new request is coming, they will be reset to nil.
	Key1 interface{}
	Key2 interface{}
	Key3 interface{}
	Data map[string]interface{}

	res *Response
	req *http.Request

	plen    int
	pnames  []string
	pvalues []string
	cookies []*http.Cookie
	query   url.Values

	logger    Logger
	buffer    BufferAllocator
	router    Router
	binder    Binder
	session   Session
	renderer  Renderer
	defaulter func(interface{}) error
	validate  func(interface{}) error
	qbinder   func(interface{}, url.Values) error
	responder func(*Context, ...interface{}) error
	notFound  Handler
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

// ClearData clears the data.
func (c *Context) ClearData() {
	if len(c.Data) != 0 {
		for key := range c.Data {
			delete(c.Data, key)
		}
	}
}

// Reset resets the context to the initalizing state.
func (c *Context) Reset() {
	c.Key1 = nil
	c.Key2 = nil
	c.Key3 = nil

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

	// (xgfone) Maybe do it??
	// c.logger = nil
	// c.buffer = nil
	// c.router = nil
	// c.binder = nil
	// c.session = nil
	// c.renderer = nil
	// c.validate = nil
	// c.defaulter = nil
	// c.qbinder = nil
	// c.responder = nil
	// c.notFound = nil
}

// SetRouter sets the router to r.
func (c *Context) SetRouter(r Router) { c.router = r }

// Router returns the router.
func (c *Context) Router() Router { return c.router }

// FindRoute finds the route from the router.
func (c *Context) FindRoute() (ok bool) {
	h, n := c.router.Match(c.req.URL.Path, c.req.Method, c.pnames, c.pvalues)
	if ok = h != nil; ok {
		c.plen = n
		switch ri := h.(type) {
		case RouteInfo:
			c.RouteInfo = ri
		case Handler:
			c.RouteInfo.Handler = ri
		default:
			panic(fmt.Errorf("unknown handler type '%T'", h))
		}
	}
	return
}

// Execute finds the route and calls the handler.
//
// SetRouter must be called before calling Execute, which be done
// by the framework.
func (c *Context) Execute() error {
	h, n := c.router.Match(c.req.URL.Path, c.req.Method, c.pnames, c.pvalues)
	if h == nil {
		return c.notFound(c)
	}

	c.plen = n
	switch ri := h.(type) {
	case RouteInfo:
		c.RouteInfo = ri
	case Handler:
		c.RouteInfo.Handler = ri
	default:
		panic(fmt.Errorf("unknown handler type '%T'", h))
	}

	return c.RouteInfo.Handler(c)
}

// SetNotFoundHandler sets the NotFound handler.
func (c *Context) SetNotFoundHandler(notFound Handler) { c.notFound = notFound }

// NotFoundHandler returns the NotFound Handler, but returns nil instead
// if not set.
func (c *Context) NotFoundHandler() Handler { return c.notFound }

//----------------------------------------------------------------------------
// URL
//----------------------------------------------------------------------------

// URLPath generates a url path by the route path name and provided parameters.
//
// Return "" if there is not the route named name.
func (c *Context) URLPath(name string, params ...interface{}) string {
	return c.router.Path(name, params...)
}

//----------------------------------------------------------------------------
// Logger
//----------------------------------------------------------------------------

// SetLogger sets the logger to logger.
func (c *Context) SetLogger(logger Logger) { c.logger = logger }

// Logger returns the logger.
func (c *Context) Logger() Logger { return c.logger }

//----------------------------------------------------------------------------
// Request & Response
//----------------------------------------------------------------------------

// SetReqRes is the same as Reset, but only reset the request and response,
// not all things.
func (c *Context) SetReqRes(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.res.SetWriter(w)
}

// SetRequest resets the request to req.
func (c *Context) SetRequest(req *http.Request) { c.req = req }

// SetResponse resets the response to resp, which will ignore nil.
func (c *Context) SetResponse(res http.ResponseWriter) { c.res.SetWriter(res) }

// Request returns the inner Request.
func (c *Context) Request() *http.Request { return c.req }

// Response returns the inner Response.
func (c *Context) Response() *Response { return c.res }

// ResponseWriter returns the underlying http.ResponseWriter.
func (c *Context) ResponseWriter() http.ResponseWriter { return c.res.ResponseWriter }

// StatusCode returns the status code of the response.
func (c *Context) StatusCode() int { return c.res.Status }

// IsResponded reports whether the response is sent.
func (c *Context) IsResponded() bool { return c.res.Wrote }

//----------------------------------------------------------------------------
// Responder
//----------------------------------------------------------------------------

// SetResponder sets the responder to handle the complicated response.
//
// For example,
//
//    responder := func(ctx *Context, args ...interface{}) error {
//        switch len(args) {
//        case 0:
//            return ctx.NoContent(http.StatusOK)
//        case 1:
//            switch v := args[0].(type) {
//            case int:
//                return ctx.NoContent(v)
//            case string:
//                return ctx.Text(http.StatusOK, v)
//            }
//        case 2:
//            switch v0 := args[0].(type) {
//            case int:
//                return ctx.Text(v0, "%v", args[1])
//            }
//        }
//        return ctx.NoContent(http.StatusInternalServerError)
//    }
//
//    router := New()
//    router.Responder =responder
//    router.Route("/path1").GET(func(c *Context) error { return c.Handle() })
//    router.Route("/path2").GET(func(c *Context) error { return c.Handle(200) })
//    router.Route("/path3").GET(func(c *Context) error { return c.Handle("Hello, World") })
//    router.Route("/path4").GET(func(c *Context) error { return c.Handle(200, "Hello, World") })
//
func (c *Context) SetResponder(h func(*Context, ...interface{}) error) { c.responder = h }

// Respond calls the context handler set by SetHandler.
//
// Return ErrNoResponder if the context handler or the global handler is not set.
func (c *Context) Respond(args ...interface{}) error {
	if c.responder == nil {
		return ErrNoResponder
	}
	return c.responder(c, args...)
}

//----------------------------------------------------------------------------
// Buffer
//----------------------------------------------------------------------------

// BufferAllocator is used to acquire and release a buffer.
type BufferAllocator interface {
	AcquireBuffer() *bytes.Buffer
	ReleaseBuffer(*bytes.Buffer)
}

// SetBufferAllocator sets the buffer allocator to alloc.
func (c *Context) SetBufferAllocator(alloc BufferAllocator) { c.buffer = alloc }

// AcquireBuffer acquires a buffer.
//
// Notice: you should call ReleaseBuffer() to release it.
func (c *Context) AcquireBuffer() *bytes.Buffer { return c.buffer.AcquireBuffer() }

// ReleaseBuffer releases a buffer into the pool.
func (c *Context) ReleaseBuffer(buf *bytes.Buffer) { c.buffer.ReleaseBuffer(buf) }

//----------------------------------------------------------------------------
// URL Params
//----------------------------------------------------------------------------

// URLParam returns the parameter value in the url path by name.
func (c *Context) URLParam(name string) string {
	for i := 0; i < c.plen; i++ {
		if c.pnames[i] == name {
			return c.pvalues[i]
		}
	}
	return ""
}

// URLParams returns all the parameters as the key-value map in the url path.
func (c *Context) URLParams() map[string]string {
	ms := make(map[string]string, c.plen)
	for i := 0; i < c.plen; i++ {
		ms[c.pnames[i]] = c.pvalues[i]
	}
	return ms
}

// URLParamNames returns the names of all the URL parameters.
func (c *Context) URLParamNames() []string { return c.pnames[:c.plen] }

// URLParamValues returns the values of all the URL parameters.
func (c *Context) URLParamValues() []string { return c.pvalues[:c.plen] }

//----------------------------------------------------------------------------
// Header
//----------------------------------------------------------------------------

// ReqHeader returns the header of the request.
func (c *Context) ReqHeader() http.Header { return c.req.Header }

// RespHeader returns the header of the response.
func (c *Context) RespHeader() http.Header { return c.res.Header() }

// Header is the alias of RespHeader to implement the interface http.ResponseWriter.
func (c *Context) Header() http.Header { return c.res.Header() }

// GetHeader returns the first value of the request header named name.
//
// Return "" if the header does not exist.
func (c *Context) GetHeader(name string, defaultValue ...string) string {
	if c.req.Header != nil {
		if vs, ok := c.req.Header[textproto.CanonicalMIMEHeaderKey(name)]; ok {
			return vs[0]
		}
	}

	if len(defaultValue) != 0 {
		return defaultValue[0]
	}

	return ""
}

// SetHeader sets the response header name to value.
func (c *Context) SetHeader(name, value string) { c.res.Header().Set(name, value) }

// AddHeader appends the value for the response header name.
func (c *Context) AddHeader(name, value string) { c.res.Header().Add(name, value) }

// DelHeader deletes the header named name from the response.
func (c *Context) DelHeader(name string) { c.res.Header().Del(name) }

// HasHeader reports whether the request header named name exists or not.
func (c *Context) HasHeader(name string) bool {
	_, ok := c.req.Header[textproto.CanonicalMIMEHeaderKey(name)]
	return ok
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

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.res, cookie)
}

//----------------------------------------------------------------------------
// Request Query
//----------------------------------------------------------------------------

// QueryParam returns the query param for the provided name.
func (c *Context) QueryParam(name string, defaultValue ...string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}

	if values, ok := c.query[name]; ok {
		return values[0]
	} else if len(defaultValue) != 0 {
		return defaultValue[0]
	}

	return ""
}

// QueryParams returns the query parameters as `url.Values`.
func (c *Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query
}

// QueryRawString returns the URL query string.
func (c *Context) QueryRawString() string {
	return c.req.URL.RawQuery
}

// HasQuery reports whether the query argument named name exists or not.
func (c *Context) HasQuery(name string) bool {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	_, ok := c.query[name]
	return ok
}

//----------------------------------------------------------------------------
// Request Form
//----------------------------------------------------------------------------

// FormValue returns the form field value for the provided name.
func (c *Context) FormValue(name string, defaultValue ...string) string {
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

// FormParams returns the form parameters as `url.Values`.
func (c *Context) FormParams() (url.Values, error) {
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
// Request Body
//----------------------------------------------------------------------------

// Body returns the reader of the request body.
func (c *Context) Body() io.ReadCloser { return c.req.Body }

// GetBody reads all the contents from the body and returns it as string.
func (c *Context) GetBody() (string, error) {
	buf := c.AcquireBuffer()
	_, err := CopyNBuffer(buf, c.req.Body, c.req.ContentLength, nil)
	body := buf.String()
	c.ReleaseBuffer(buf)
	return body, err
}

// GetBodyReader reads all the contents from the body to buffer and returns it.
//
// Notice: You should call ReleaseBuffer(buf) to release the buffer at last.
func (c *Context) GetBodyReader() (buf *bytes.Buffer, err error) {
	buf = c.AcquireBuffer()
	_, err = CopyNBuffer(buf, c.req.Body, c.req.ContentLength, nil)
	if err != nil {
		c.ReleaseBuffer(buf)
		return nil, err
	}
	return
}

//----------------------------------------------------------------------------
// Other Request Information
//----------------------------------------------------------------------------

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

// Method returns the method of the request.
func (c *Context) Method() string { return c.req.Method }

// Path returns the path of the request.
func (c *Context) Path() string { return c.req.URL.Path }

// Referer returns the Referer header of the request.
func (c *Context) Referer() string { return c.req.Referer() }

// UserAgent returns the User-Agent header of the request.
func (c *Context) UserAgent() string { return c.req.UserAgent() }

// RemoteAddr returns the remote address of the http connection.
func (c *Context) RemoteAddr() string { return c.req.RemoteAddr }

// RequestURI returns the URI of the request.
func (c *Context) RequestURI() string { return c.req.RequestURI }

// ContentLength return the length of the request body.
func (c *Context) ContentLength() int64 { return c.req.ContentLength }

// BasicAuth returns the username and password from the request.
func (c *Context) BasicAuth() (username, password string, ok bool) {
	return c.req.BasicAuth()
}

// Accept returns the content of the header Accept.
//
// If there is no the header Accept , it return nil.
//
// Notice:
//
//   1. It will sort the content by the q-factor weighting.
//   2. If the value is "<MIME_type>/*", it will be amended as "<MIME_type>/".
//      So you can use it to match the prefix.
//   3. If the value is "*/*", it will be amended as "".
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

	sort.SliceStable(accepts, func(i, j int) bool { return accepts[i].q < accepts[j].q })

	results := make([]string, len(accepts))
	for i := range accepts {
		results[i] = accepts[i].ct
	}
	return results
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *Context) Scheme() (scheme string) {
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

// ClientIP returns the real client's network address based on `X-Forwarded-For`
// or `X-Real-IP` request header. Or returns the remote address.
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
	for index := strings.IndexByte(ct, ';'); index > 0; index = strings.IndexByte(ct, ';') {
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
// Session Management
//----------------------------------------------------------------------------

// SetSessionManagement sets the session management to s.
func (c *Context) SetSessionManagement(s Session) { c.session = s }

// GetSession returns the session content by id from the backend store.
//
// If the session id does not exist, it returns ErrSessionNotExist.
func (c *Context) GetSession(id string) (v interface{}, err error) {
	if id == "" {
		return nil, ErrInvalidSession
	} else if c.session == nil {
		return nil, ErrNoSessionSupport
	} else if v, err = c.session.GetSession(id); err == nil {
		if v == nil {
			err = ErrSessionNotExist
		}
	}

	return
}

// SetSession sets the session to the backend store.
func (c *Context) SetSession(id string, value interface{}) (err error) {
	if id == "" || value == nil {
		return ErrInvalidSession
	} else if c.session == nil {
		return ErrNoSessionSupport
	}

	return c.session.SetSession(id, value)
}

// DelSession deletes the session from the backend store.
func (c *Context) DelSession(id string) (err error) {
	if id == "" {
		return ErrInvalidSession
	} else if c.session == nil {
		return ErrNoSessionSupport
	}

	return c.session.DelSession(id)
}

//----------------------------------------------------------------------------
// Validator
//----------------------------------------------------------------------------

// SetValidator sets the validator to v to validate the argument when Binding.
func (c *Context) SetValidator(v func(interface{}) error) { c.validate = v }

// Validate validates whether the argument v is valid.
func (c *Context) Validate(v interface{}) error { return c.validate(v) }

//----------------------------------------------------------------------------
// SetDefault
//----------------------------------------------------------------------------

// SetDefaulter sets the default setter to v set the default value.
func (c *Context) SetDefaulter(v func(interface{}) error) { c.defaulter = v }

// SetDefault calls the default setter to set the default value of v.
func (c *Context) SetDefault(v interface{}) error { return c.defaulter(v) }

//----------------------------------------------------------------------------
// Binder
//----------------------------------------------------------------------------

// SetBinder sets the binder to b to bind the request information to an object.
func (c *Context) SetBinder(b Binder) { c.binder = b }

// Bind binds the request information into the provided value v.
//
// The default binder does it based on Content-Type header.
//
// Notice: it will call the interfaces or functions in turn as follow:
//   1. Binder
//   2. SetDefault
//   3. Validator
func (c *Context) Bind(v interface{}) (err error) {
	if err = c.binder.Bind(c.req, v); err == nil {
		if err = c.defaulter(v); err == nil {
			err = c.validate(v)
		}
	}
	return
}

// SetQueryBinder sets the query binder to f to bind the url query to an object.
func (c *Context) SetQueryBinder(f func(interface{}, url.Values) error) { c.qbinder = f }

// BindQuery binds the request URL query into the provided value v.
//
// Notice: it will call the interfaces or functions in turn as follow:
//   1. QueryBinder
//   2. SetDefault
//   3. Validator
func (c *Context) BindQuery(v interface{}) (err error) {
	if err = c.qbinder(v, c.QueryParams()); err == nil {
		if err = c.defaulter(v); err == nil {
			err = c.validate(v)
		}
	}
	return
}

//----------------------------------------------------------------------------
// Renderer
//----------------------------------------------------------------------------

// SetRenderer sets the renderer to r to render the response to the peer.
func (c *Context) SetRenderer(r Renderer) { c.renderer = r }

// Render renders a template named name with data and sends a text/html response
// with status code.
func (c *Context) Render(name string, code int, data interface{}) error {
	if c.renderer == nil {
		return ErrRendererNotRegistered
	}
	return c.renderer.Render(c, name, code, data)
}

// RenderOk is short for c.Render(name, http.StatusOK, data).
func (c *Context) RenderOk(name string, data interface{}) error {
	return c.Render(name, http.StatusOK, data)
}

//----------------------------------------------------------------------------
// Set Repsonse
//----------------------------------------------------------------------------

// SetConnectionClose tell the server to close the connection.
func (c *Context) SetConnectionClose() {
	c.res.Header().Set(HeaderConnection, "close")
}

// SetContentType sets the Content-Type header of the response body to ct,
// but does nothing if contentType is "".
func (c *Context) SetContentType(ct string) {
	if ct != "" {
		SetContentType(c.res, ct)
	}
}

//----------------------------------------------------------------------------
// Send Repsonse
//----------------------------------------------------------------------------

// WriteHeader sends an HTTP response header with the provided status code,
// which implements the interface http.ResponseWriter.
func (c *Context) WriteHeader(statusCode int) { c.res.WriteHeader(statusCode) }

// Write writes the content to the peer, which implements the interface
// http.ResponseWriter.
//
// it will write the header firstly with 200 if the header is not sent.
func (c *Context) Write(b []byte) (int, error) { return c.res.Write(b) }

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) error { c.res.WriteHeader(code); return nil }

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

// Stream sends a streaming response with status code and content type.
func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.setContentTypeAndCode(code, contentType)
	_, err = io.CopyBuffer(c.res, r, make([]byte, 2048))
	return
}

// Blob sends a blob response with status code and content type.
func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.setContentTypeAndCode(code, contentType)
	_, err = c.res.Write(b)
	return
}

// BlobText sends a string blob response with status code and content type.
func (c *Context) BlobText(code int, contentType string, format string,
	args ...interface{}) (err error) {
	c.setContentTypeAndCode(code, contentType)
	if len(args) > 0 {
		_, err = fmt.Fprintf(c.res, format, args...)
	} else {
		_, err = c.res.WriteString(format)
	}
	return err
}

// Text sends a string response with status code.
func (c *Context) Text(code int, format string, args ...interface{}) error {
	return c.BlobText(code, MIMETextPlainCharsetUTF8, format, args...)
}

// Error sends an error response with status code.
func (c *Context) Error(code int, err error) HTTPServerError {
	return HTTPServerError{Code: code, Err: err}
}

// JSON sends a JSON response with status code.
func (c *Context) JSON(code int, v interface{}) error {
	c.setContentTypeAndCode(code, MIMEApplicationJSONCharsetUTF8)
	return json.NewEncoder(c.res).Encode(v)
}

// JSONPretty sends a pretty-print JSON with status code.
func (c *Context) JSONPretty(code int, v interface{}, indent string) error {
	c.setContentTypeAndCode(code, MIMEApplicationJSONCharsetUTF8)
	enc := json.NewEncoder(c.res)
	enc.SetIndent("", indent)
	return enc.Encode(v)
}

// JSONBlob sends a JSON blob response with status code.
func (c *Context) JSONBlob(code int, b []byte) error {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *Context) JSONP(code int, callback string, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONPBlob(code, callback, b)
}

// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
// to construct the JSONP payload.
func (c *Context) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.SetContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.res.WriteHeader(code)
	if _, err = c.res.WriteString(callback); err != nil {
		return
	} else if _, err = c.res.WriteString("("); err != nil {
		return
	} else if _, err = c.res.Write(b); err != nil {
		return
	}
	_, err = c.res.WriteString("):")
	return
}

// XML sends an XML response with status code.
func (c *Context) XML(code int, v interface{}) error {
	c.setContentTypeAndCode(code, MIMEApplicationXMLCharsetUTF8)
	if _, err := c.res.WriteString(xml.Header); err != nil {
		return err
	}
	return xml.NewEncoder(c.res).Encode(v)
}

// XMLPretty sends a pretty-print XML with status code.
func (c *Context) XMLPretty(code int, v interface{}, indent string) error {
	c.setContentTypeAndCode(code, MIMEApplicationXMLCharsetUTF8)
	if _, err := c.res.WriteString(xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(c.res)
	enc.Indent("", indent)
	return enc.Encode(v)
}

// XMLBlob sends an XML blob response with status code.
func (c *Context) XMLBlob(code int, b []byte) (err error) {
	c.setContentTypeAndCode(code, MIMEApplicationXMLCharsetUTF8)
	if _, err = c.res.WriteString(xml.Header); err != nil {
		return
	}
	_, err = c.res.Write(b)
	return
}

// HTML sends an HTTP response with status code.
func (c *Context) HTML(code int, html string) error {
	return c.BlobText(code, MIMETextHTMLCharsetUTF8, html)
}

// HTMLBlob sends an HTTP blob response with status code.
func (c *Context) HTMLBlob(code int, b []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

// File sends a response with the content of the file.
//
// If the file does not exist, it returns ErrNotFound.
//
// If not set the Content-Type, it will deduce it from the extension
// of the file name.
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
	disposition := fmt.Sprintf("%s; filename=%q", dispositionType, name)
	c.res.Header().Set(HeaderContentDisposition, disposition)
	return c.File(file)
}

// Attachment sends a response as attachment, prompting client to save the file.
//
// If the file does not exist, it returns ErrNotFound.
func (c *Context) Attachment(file string, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

// Inline sends a response as inline, opening the file in the browser.
//
// If the file does not exist, it returns ErrNotFound.
func (c *Context) Inline(file string, name string) error {
	return c.contentDisposition(file, name, "inline")
}
