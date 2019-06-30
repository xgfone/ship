// Copyright 2019 xgfone
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
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/xgfone/go-tools/function"
	"github.com/xgfone/go-tools/io2"
)

var contenttypes = map[string][]string{}

// AddContentTypeToSlice add a rule to convert contentType to contentTypeSlice.
//
// When calling `Context#SetContentType(contentType)` to set the response header
// Content-Type, it will use contentTypeSlice to avoid to allocate the memory.
// See the function `ToContentTypes(contentType)`.
func AddContentTypeToSlice(contentType string, contentTypeSlice []string) {
	if contentType == "" {
		panic(fmt.Errorf("the Content-Type is empty"))
	}
	if len(contentTypeSlice) == 0 {
		panic(fmt.Errorf("the Content-Type slice is empty"))
	}
	contenttypes[contentType] = contentTypeSlice
}

// ToContentTypes converts the Content-Type to the Content-Type slice.
func ToContentTypes(contentType string) []string {
	return toContentTypes(contentType)
}

func toContentTypes(contentType string) []string {
	switch contentType {
	case MIMEApplicationJSON:
		return MIMEApplicationJSONs
	case MIMEApplicationJSONCharsetUTF8:
		return MIMEApplicationJSONCharsetUTF8s
	case MIMEApplicationJavaScript:
		return MIMEApplicationJavaScripts
	case MIMEApplicationJavaScriptCharsetUTF8:
		return MIMEApplicationJavaScriptCharsetUTF8s
	case MIMEApplicationXML:
		return MIMEApplicationXMLs
	case MIMEApplicationXMLCharsetUTF8:
		return MIMEApplicationXMLCharsetUTF8s
	case MIMETextXML:
		return MIMETextXMLs
	case MIMETextXMLCharsetUTF8:
		return MIMETextXMLCharsetUTF8s
	case MIMEApplicationForm:
		return MIMEApplicationForms
	case MIMEApplicationProtobuf:
		return MIMEApplicationProtobufs
	case MIMEApplicationMsgpack:
		return MIMEApplicationMsgpacks
	case MIMETextHTML:
		return MIMETextHTMLs
	case MIMETextHTMLCharsetUTF8:
		return MIMETextHTMLCharsetUTF8s
	case MIMETextPlain:
		return MIMETextPlains
	case MIMETextPlainCharsetUTF8:
		return MIMETextPlainCharsetUTF8s
	case MIMEMultipartForm:
		return MIMEMultipartForms
	case MIMEOctetStream:
		return MIMEOctetStreams
	default:
		if ss := contenttypes[contentType]; ss != nil {
			return ss
		}
		return []string{contentType}
	}
}

var (
	indexPage  = "index.html"
	emptyStrS  = [256]string{}
	emptyValue = emptyType(0)
)

type emptyType uint8

// MaxMemoryLimit is the maximum memory.
var MaxMemoryLimit int64 = 32 << 20 // 32MB

type contextKeyT int

var contextKey contextKeyT

func setContext(ctx *Context) {
	if ctx.req != nil {
		ctx.req = ctx.req.WithContext(context.WithValue(context.TODO(), contextKey, ctx))
	}
}

// GetContext gets the Context from the http Request.
//
// Notice: you must enable it by SetEnableCtxHTTPContext(true).
func GetContext(req *http.Request) *Context {
	if v := req.Context().Value(contextKey); v != nil {
		return v.(*Context)
	}
	return nil
}

type responder struct {
	resp http.ResponseWriter
	ctx  *Context
}

func newResponder(ctx *Context, resp http.ResponseWriter) responder {
	return responder{ctx: ctx, resp: resp}
}

func (r *responder) reset(resp http.ResponseWriter) {
	r.resp = resp
}

func (r responder) Header() http.Header {
	return r.resp.Header()
}

func (r responder) Write(p []byte) (int, error) {
	if !r.ctx.wrote {
		r.ctx.wrote = true
	}
	return r.resp.Write(p)
}

func (r responder) WriteString(s string) (int, error) {
	return io.WriteString(r.resp, s)
}

// WriteHeader implements http.ResponseWriter#WriteHeader().
func (r responder) WriteHeader(code int) {
	r.resp.WriteHeader(code)
	r.ctx.wrote = true
	r.ctx.code = code
}

// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r responder) Flush() {
	r.resp.(http.Flusher).Flush()
}

// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r responder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.resp.(http.Hijacker).Hijack()
}

// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r responder) CloseNotify() <-chan bool {
	return r.resp.(http.CloseNotifier).CloseNotify()
}

// Context represetns a request and response context.
type Context struct {
	// you can use it to pass the error between the handlers or the middlewares.
	//
	// Notice: when the new request is coming, it will be reset to nil.
	Err error

	// ReqCtxData is the data what each request has, the lifecycle of which is
	// the same as this context, that's, when creating this context, it will
	// call `newCtxData()`, which is set by `SetNewCtxData()` as the option of
	// the ship router, to creating ReqCtxData. When finishing the request,
	// it will be reset by the context and put into the pool with this context.
	ReqCtxData Resetter

	// Data is used to store many the key-value pairs about the context.
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

	ship  *Ship
	code  int
	wrote bool

	req    *http.Request
	resp   responder
	query  url.Values
	router Router

	handler func(*Context, ...interface{}) error
	pnames  []string
	pvalues []string

	sessionK string
	sessionV interface{}
}

// NewContext returns a new context.
func newContext(s *Ship, req *http.Request, resp http.ResponseWriter, maxParam int) *Context {
	var pnames, pvalues []string
	if maxParam > 0 {
		pnames = make([]string, maxParam)
		pvalues = make([]string, maxParam)
	}

	ctx := &Context{
		ship:    s,
		code:    200,
		pnames:  pnames,
		pvalues: pvalues,
		Data:    make(map[string]interface{}, s.ctxDataSize),
	}
	ctx.setReqResp(req, resp)

	if s.newCtxData != nil {
		ctx.ReqCtxData = s.newCtxData(ctx)
	}

	return ctx
}

func (c *Context) reset() {
	c.Err = nil
	c.Key1 = nil
	c.Key2 = nil
	c.Key3 = nil
	c.ClearData()

	if c.ReqCtxData != nil {
		c.ReqCtxData.Reset()
	}

	c.code = 200
	c.wrote = false

	c.req = nil
	c.resp.reset(nil)
	c.query = nil
	c.router = nil

	c.handler = nil
	c.resetURLParam()

	c.sessionK = ""
	c.sessionV = nil
}

func (c *Context) resetURLParam() {
	copy(c.pnames, emptyStrS[:len(c.pnames)])
	copy(c.pvalues, emptyStrS[:len(c.pvalues)])
}

func (c *Context) setReqResp(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.resp = newResponder(c, w)
	if c.ship.enableCtxHTTPContext {
		setContext(c)
	}
}

// ClearData clears the data.
func (c *Context) ClearData() {
	for key := range c.Data {
		delete(c.Data, key)
	}
}

// FindHandler finds the registered router handler by the method and
// path of the request.
//
// Return nil if not found.
func (c *Context) FindHandler(method, path string) Handler {
	c.resetURLParam()
	return c.findHandler(method, path)
}

func (c *Context) findHandler(method, path string) Handler {
	switch h := c.router.Find(method, path, c.pnames, c.pvalues).(type) {
	case Handler:
		return h
	case func(ctx *Context) error:
		return Handler(h)
	case nil:
		return nil
	default:
		panic(fmt.Errorf("unknown handler type '%T'", h))
	}
}

// NotFoundHandler returns the configured NotFound handler.
func (c *Context) NotFoundHandler() Handler {
	return c.ship.notFoundHandler
}

// URL generates an URL by route name and provided parameters.
func (c *Context) URL(name string, params ...interface{}) string {
	return c.ship.URL(name, params...)
}

// Logger returns the logger implementation.
func (c *Context) Logger() Logger {
	return c.ship.logger
}

// Router returns the router.
func (c *Context) Router() Router {
	return c.router
}

// AcquireBuffer acquires a buffer.
//
// Notice: you should call ReleaseBuffer() to release it.
func (c *Context) AcquireBuffer() *bytes.Buffer {
	return c.ship.AcquireBuffer()
}

// ReleaseBuffer releases a buffer into the pool.
func (c *Context) ReleaseBuffer(buf *bytes.Buffer) {
	c.ship.ReleaseBuffer(buf)
}

// SetHandler sets a context handler in order to call it across the functions
// by the method Handle(), which is used to handle the various arguments.
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
//                return ctx.String(http.StatusOK, v)
//            }
//        case 2:
//            switch v0 := args[0].(type) {
//            case int:
//                return ctx.String(v0, "%v", args[1])
//            }
//        }
//        return ctx.NoContent(http.StatusInternalServerError)
//    }
//
//    sethandler := func(next Handler) Handler {
//        return func(ctx *Context) error {
//            ctx.SetHandler(responder)
//            return next(ctx)
//        }
//    }
//
//    router := New()
//    router.Use(sethandler)
//    router.Route("/path1").GET(func(c *Context) error { return c.Handle() })
//    router.Route("/path2").GET(func(c *Context) error { return c.Handle(200) })
//    router.Route("/path3").GET(func(c *Context) error { return c.Handle("Hello, World") })
//    router.Route("/path4").GET(func(c *Context) error { return c.Handle(200, "Hello, World") })
//
func (c *Context) SetHandler(h func(*Context, ...interface{}) error) {
	c.handler = h
}

// Handle calls the context handler.
//
// Return ErrNoHandler if the context handler or the global handler is not set.
func (c *Context) Handle(args ...interface{}) error {
	if c.handler != nil {
		return c.handler(c, args...)
	} else if c.ship.ctxHandler != nil {
		return c.ship.ctxHandler(c, args...)
	}
	return ErrNoHandler
}

// Request returns the inner Request.
func (c *Context) Request() *http.Request {
	return c.req
}

// Response returns the inner http.ResponseWriter.
func (c *Context) Response() http.ResponseWriter {
	return newResponder(c, c.resp.resp)
}

// StatusCode returns the status code of the response.
//
// Notice: it's used by the middleware, such as Logger in general.
func (c *Context) StatusCode() int {
	return c.code
}

// IsResponded reports whether the response is sent.
func (c *Context) IsResponded() bool {
	return c.wrote
}

// SetResponded sets the response to be sent or not.
func (c *Context) SetResponded(yes bool) {
	c.wrote = yes
}

// SetResponse resets the response to resp, which will ignore nil.
func (c *Context) SetResponse(resp http.ResponseWriter) {
	if resp != nil {
		c.resp.reset(resp)
	}
}

// SetConnectionClose tell the server to close the connection.
func (c *Context) SetConnectionClose() {
	c.resp.Header().Set(HeaderConnection, "close")
}

// Param returns the parameter value in the url path by name.
func (c *Context) Param(name string) string {
	for i, _len := 0, len(c.pnames); i < _len; i++ {
		switch v := c.pnames[i]; v {
		case "":
			return ""
		case name:
			return c.pvalues[i]
		}
	}
	return ""
}

// Params returns all the parameters as the key-value map in the url path.
func (c *Context) Params() map[string]string {
	_len := len(c.pnames)
	ms := make(map[string]string, _len)
	for i := 0; i < _len; i++ {
		if c.pnames[i] == "" {
			break
		}
		ms[c.pnames[i]] = c.pvalues[i]
	}
	return ms
}

// ParamNames returns all the names of the URL parameters.
func (c *Context) ParamNames() []string {
	i, _len := 0, len(c.pnames)
	for ; i < _len; i++ {
		if c.pnames[i] == "" {
			return c.pnames[:i]
		}
	}
	if i == 0 {
		return []string{}
	}
	return c.pnames
}

// ParamValues returns all the names of the URL parameters.
func (c *Context) ParamValues() []string {
	i, _len := 0, len(c.pnames)
	for ; i < _len; i++ {
		if c.pnames[i] == "" {
			return c.pvalues[:i]
		}
	}
	if i == 0 {
		return []string{}
	}
	return c.pvalues
}

// ParamToStruct scans the url parameters to a pointer v to the struct.
//
// For the struct, the argument name is the field name by default. But you can
// change it by the tag "url", such as `url:"name"`. The tag `url:"-"`, however,
// will ignore this field.
func (c *Context) ParamToStruct(v interface{}) error {
	if v == nil {
		return errors.New("the argument is nil")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return errors.New("the argument is not a pointer")
	} else if value = value.Elem(); value.Kind() != reflect.Struct {
		return errors.New("the argument is not a pointer to struct")
	}

	vtype := value.Type()
	for i, num := 0, value.NumField(); i < num; i++ {
		fieldv := value.Field(i)
		fieldt := vtype.Field(i)

		name := fieldt.Name
		if n := fieldt.Tag.Get("url"); n != "" {
			if n == "-" {
				continue
			}
			name = n
		}

		// Check whether the field can be set.
		if !fieldv.CanSet() {
			continue
		}

		if fieldv.Kind() != reflect.Ptr {
			fieldv = fieldv.Addr()
		}

		if v := c.Param(name); v != "" {
			if err := function.SetValue(fieldv.Interface(), v); err != nil {
				return err
			}
		}
	}

	return nil
}

// Header is the alias of GetHeader.
func (c *Context) Header(name string) string {
	return c.GetHeader(name)
}

// GetHeader returns the first value of the request header named name.
//
// Return "" if the header does not exist.
func (c *Context) GetHeader(name string) string {
	return c.req.Header.Get(name)
}

// SetHeader sets the response header name to value.
func (c *Context) SetHeader(name, value string) {
	c.resp.Header().Set(name, value)
}

// IsDebug reports whether to enable the debug mode.
func (c *Context) IsDebug() bool {
	return c.ship.debug
}

// IsTLS reports whether HTTP connection is TLS or not.
func (c *Context) IsTLS() bool {
	return c.req.TLS != nil
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

// IsAjax reports whether the request is ajax or not.
func (c *Context) IsAjax() bool {
	return c.req.Header.Get(HeaderXRequestedWith) == "XMLHttpRequest"
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

// Host returns the host of the request.
func (c *Context) Host() string {
	return c.req.Host
}

// Method returns the method of the request.
func (c *Context) Method() string {
	return c.req.Method
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

// RealIP returns the client's network address based on `X-Forwarded-For`
// or `X-Real-IP` request header.
func (c *Context) RealIP() string {
	if ip := c.req.Header.Get(HeaderXForwardedFor); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := c.req.Header.Get(HeaderXRealIP); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.req.RemoteAddr)
	return ra
}

// RemoteAddr returns the remote address of the http connection.
func (c *Context) RemoteAddr() string {
	return c.req.RemoteAddr
}

// RequestURI returns the URI of the request.
func (c *Context) RequestURI() string {
	return c.req.RequestURI
}

// Charset returns the charset of the request content.
//
// Return "" if there is no charset.
func (c *Context) Charset() string {
	ct := c.req.Header.Get(HeaderContentType)
	index := strings.IndexByte(ct, ';')
	for ; index > 0; index = strings.IndexByte(ct, ';') {
		ct = ct[index:]
		if index = strings.IndexByte(ct, '='); index > 0 {
			if strings.HasSuffix(ct[:index], "charset") {
				return ct[index+1:]
			}
			ct = ct[index+1:]
		}
	}
	return ""
}

// ContentType returns the Content-Type of the request without the charset.
func (c *Context) ContentType() (ct string) {
	ct = c.req.Header.Get(HeaderContentType)
	if index := strings.IndexAny(ct, " ;"); index > 0 {
		ct = ct[:index]
	}
	return
}

// ContentLength return the length of the request body.
func (c *Context) ContentLength() int64 {
	return c.req.ContentLength
}

// SetContentTypes sets the Content-Type of the response body to more than one.
func (c *Context) SetContentTypes(contentTypes []string) {
	c.resp.Header()[HeaderContentType] = contentTypes
}

// SetContentType sets the Content-Type of the response body to contentType,
// but does nothing if contentType is "".
//
// Notice: In order to avoid the memory allocation to improve performance,
// it will look up the corresponding Content-Type slice constant firstly,
// or generate one by the argument contentType, then set the response header
// Content-Type to the Content-Type slice. Howevre, you can call
// SetContentTypes() to set it to avoid the memory allocation, and pass ""
// to the response function, for example, JSON(), String(), Blob(), etc.
//
// For the pre-defined Content-Type slices, please see
// https://godoc.org/github.com/xgfone/ship/#pkg-variables.
func (c *Context) SetContentType(contentType string) {
	if contentType != "" {
		c.SetContentTypes(toContentTypes(contentType))
	}
}

// GetBody reads all the contents from the body and returns it as string.
func (c *Context) GetBody() (string, error) {
	buf := c.AcquireBuffer()
	err := io2.ReadNWriter(buf, c.req.Body, c.req.ContentLength)
	body := buf.String()
	c.ReleaseBuffer(buf)
	return body, err
}

// GetBodyReader reads all the contents from the body to buffer and returns it.
//
// Notice: You should call ReleaseBuffer(buf) to release the buffer at last.
func (c *Context) GetBodyReader() (buf *bytes.Buffer, err error) {
	buf = c.AcquireBuffer()
	err = io2.ReadNWriter(buf, c.req.Body, c.req.ContentLength)
	if err != nil {
		c.ReleaseBuffer(buf)
		return nil, err
	}
	return
}

// QueryParam returns the query param for the provided name.
func (c *Context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.req.URL.Query()
	}
	return c.query.Get(name)
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

// FormValue returns the form field value for the provided name.
func (c *Context) FormValue(name string) string {
	return c.req.FormValue(name)
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
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.req.FormFile(name)
	return fh, err
}

// MultipartForm returns the multipart form.
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.req.ParseMultipartForm(MaxMemoryLimit)
	return c.req.MultipartForm, err
}

// Cookie returns the named cookie provided in the request.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *Context) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.resp, cookie)
}

// GetSession returns the session content by id from the backend store.
//
// If the session id does not exist, it maybe return (nil, nil).
//
// Notice: for the same session id, the context maybe optimize GetSession
// by the cache, which will call the backend store only once.
func (c *Context) GetSession(id string) (v interface{}, err error) {
	if id == "" {
		return nil, ErrSessionNotExist
	}
	if c.sessionK == id {
		switch c.sessionV {
		case nil, emptyValue:
			return nil, ErrSessionNotExist
		}
		return c.sessionV, nil
	}

	if c.ship.session == nil {
		return nil, ErrNoSessionSupport
	}
	if v, err = c.ship.session.GetSession(id); err == nil {
		c.sessionK = id
		if v == nil {
			err = ErrSessionNotExist
			c.sessionV = emptyValue
		} else {
			c.sessionV = v
		}
	}

	return
}

// SetSession sets the session to the backend store.
//
// id must not be "".
//
// value should not be nil. If nil, however, it will tell the context
// that the session id is missing, and the context should not forward
// the request to the underlying session store when calling GetSession.
func (c *Context) SetSession(id string, value interface{}) (err error) {
	if c.ship.session == nil {
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
	if err = c.ship.session.SetSession(id, value); err != nil {
		return err
	}
	c.sessionK = id
	c.sessionV = value
	return nil
}

// DelSession deletes the session from the backend store.
//
// id must not be "".
func (c *Context) DelSession(id string) (err error) {
	if id == "" {
		return ErrInvalidSession
	}
	if c.ship.session == nil {
		return ErrNoSessionSupport
	}
	if err = c.ship.session.DelSession(id); err != nil {
		return
	}
	if c.sessionK == id {
		c.sessionV = nil
	}
	return nil
}

// Bind binds the request information into the provided value v.
//
// The default binder does it based on Content-Type header.
func (c *Context) Bind(v interface{}) error {
	return c.ship.binder.Bind(c, v)
}

// BindQuery binds the request URL query into the provided value v.
func (c *Context) BindQuery(v interface{}) error {
	return c.ship.bindQuery(v, c.QueryParams())
}

// Render renders a template named name with data and sends a text/html response
// with status code.
func (c *Context) Render(name string, code int, data interface{}) error {
	if c.ship.renderer == nil {
		return ErrRendererNotRegistered
	}
	return c.ship.renderer.Render(c, name, code, data)
}

// Write writes the content to the peer.
//
// it will write the header firstly if the header is not sent.
func (c *Context) Write(b []byte) (int, error) {
	return c.resp.Write(b)
}

// NoContent sends a response with no body and a status code.
func (c *Context) NoContent(code int) error {
	c.resp.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, toURL string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.resp.Header().Set(HeaderLocation, toURL)
	c.resp.WriteHeader(code)
	return nil
}

// Stream sends a streaming response with status code and content type.
func (c *Context) Stream(code int, contentType string, r io.Reader) (err error) {
	c.SetContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = io.Copy(c.resp, r)
	return
}

// Blob sends a blob response with status code and content type.
func (c *Context) Blob(code int, contentType string, b []byte) (err error) {
	c.SetContentType(contentType)
	c.resp.WriteHeader(code)
	_, err = c.resp.Write(b)
	return
}

// BlobString sends a string blob response with status code and content type.
func (c *Context) BlobString(code int, contentType string, format string, args ...interface{}) (err error) {
	c.SetContentType(contentType)
	c.resp.WriteHeader(code)
	if len(args) > 0 {
		_, err = fmt.Fprintf(c.resp, format, args...)
	} else {
		_, err = io.WriteString(c.resp, format)
	}
	return err
}

// String sends a string response with status code.
func (c *Context) String(code int, format string, args ...interface{}) (err error) {
	return c.BlobString(code, MIMETextPlainCharsetUTF8, format, args...)
}

// Error sends an error response with status code.
func (c *Context) Error(code int, err error) error {
	return c.String(code, "%s", err.Error())
}

// JSON sends a JSON response with status code.
func (c *Context) JSON(code int, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
}

// JSONPretty sends a pretty-print JSON with status code.
func (c *Context) JSONPretty(code int, i interface{}, indent string) error {
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.JSONBlob(code, b)
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
func (c *Context) XML(code int, i interface{}) error {
	b, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLPretty sends a pretty-print XML with status code.
func (c *Context) XMLPretty(code int, i interface{}, indent string) error {
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return err
	}
	return c.XMLBlob(code, b)
}

// XMLBlob sends an XML blob response with status code.
func (c *Context) XMLBlob(code int, b []byte) (err error) {
	c.SetContentType(MIMEApplicationXMLCharsetUTF8)
	c.resp.WriteHeader(code)
	if _, err = c.resp.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.resp.Write(b)
	return
}

// HTML sends an HTTP response with status code.
func (c *Context) HTML(code int, html string) error {
	return c.BlobString(code, MIMETextHTMLCharsetUTF8, html)
}

// HTMLBlob sends an HTTP blob response with status code.
func (c *Context) HTMLBlob(code int, b []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

// File sends a response with the content of the file.
//
// If the file does not exist, it returns ErrNotFound.
func (c *Context) File(file string) (err error) {
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

func (c *Context) contentDisposition(file, name, dispositionType string) error {
	c.resp.Header().Set(HeaderContentDisposition,
		fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

// Attachment sends a response as attachment, prompting client to save the
// file.
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
