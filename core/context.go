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

package core

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Context stands for a context interface.
//
// This interface will be used by the built-in implementation of this project.
// And at the moment this project does not support the customized implementation.
type Context interface {
	// Report whether the response has been sent.
	IsResponse() bool

	// Find the registered router handler by the method and path of the request.
	//
	// Return nil if not found.
	FindHandler(method string, path string) Handler

	NotFoundHandler() Handler

	AcquireBuffer() *bytes.Buffer
	ReleaseBuffer(*bytes.Buffer)

	Request() *http.Request
	Response() http.ResponseWriter
	SetResponse(http.ResponseWriter)
	SetResponded(bool)
	SetConnectionClose() // Tell the server to close the connection.

	// These may be passed the error between the handlers.
	Error() error
	HasError() bool
	SetError(err error)

	IsTLS() bool
	IsDebug() bool
	IsAjax() bool
	IsWebSocket() bool

	Header(name string) (value string)
	SetHeader(name string, value string)

	// URL Parameter. We remove the URL prefix for the convenience.
	Param(name string) (value string)
	Params() map[string]string // Return the key-value map of the url parameters
	ParamNames() []string      // Return the list of the url parameter names
	ParamValues() []string     // Return the list of the url parameter values

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
	Accept() []string
	Host() string
	Method() string
	Scheme() string
	RealIP() string
	Charset() string
	RemoteAddr() string
	RequestURI() string
	ContentType() string // It should not contain the charset.
	ContentLength() int64
	GetBody() (string, error)
	// You should call Context.ReleaseBuffer(buf) to release the buffer at last.
	GetBodyReader() (buf *bytes.Buffer, err error)
	SetContentType(string)

	QueryParam(name string) (value string)
	QueryParams() url.Values
	QueryRawString() string

	FormParams() (url.Values, error)
	FormValue(name string) (value string)
	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)

	Cookies() []*http.Cookie
	Cookie(name string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)

	// If the session id does not exist, it maybe return (nil, nil).
	//
	// Notice: for the same session id, the context maybe optimize GetSession
	// by the cache, which will call the backend store only once.
	GetSession(id string) (interface{}, error)
	// id must not be "".
	//
	// value should not be nil. If nil, however, it will tell the context
	// that the session id is missing, and the context should not forward
	// the request to the underlying session store when calling GetSession.
	SetSession(id string, value interface{}) error
	// id must not be "".
	DelSession(id string) error

	// Get and Set are used to store the key-value information about the context.
	Store() map[string]interface{}
	Get(key string) (value interface{})
	Set(key string, value interface{})
	Del(key string)
	// Store maybe use a map to store the key-value, so it asks the system
	// to allocate the memory. If the interim context value is too few and
	// you don't want the system to allocate the memory for the map,
	// the context supplies three context variables for you and you can consider
	// them as the context register to use.
	Key1() (value interface{})
	Key2() (value interface{})
	Key3() (value interface{})
	SetKey1(value interface{})
	SetKey2(value interface{})
	SetKey3(value interface{})

	// You can set a handler then call it across the functions, which is used to
	// handle the various arguments. For example,
	//
	//    responder := func(ctx Context, args ...interface{}) error {
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
	//                return ctx.String(v0, fmt.Sprintf("%v", args[1]))
	//            }
	//        }
	//        return ctx.NoContent(http.StatusInternalServerError)
	//    }
	//
	//    sethandler := func(next Handler) Handler {
	//        return func(ctx Context) error {
	//            ctx.SetHandler(responder)
	//            return next(ctx)
	//        }
	//    }
	//
	//    router := New()
	//    router.Use(sethandler)
	//    router.Route("/path1").GET(func(c Context) error { return c.Handle() })
	//    router.Route("/path2").GET(func(c Context) error { return c.Handle(200) })
	//    router.Route("/path3").GET(func(c Context) error { return c.Handle("Hello, World") })
	//    router.Route("/path4").GET(func(c Context) error { return c.Handle(200, "Hello, World") })
	//
	SetHandler(func(ctx Context, args ...interface{}) error)
	Handle(args ...interface{}) error

	Logger() Logger
	URL(name string, params ...interface{}) string

	Bind(v interface{}) error
	BindQuery(v interface{}) error

	Write([]byte) (int, error)
	Render(name string, code int, data interface{}) error

	NoContent(code int) error
	Redirect(code int, toURL string) error
	String(code int, data string) error
	Blob(code int, contentType string, b []byte) error

	HTML(code int, html string) error
	HTMLBlob(code int, b []byte) error

	JSON(code int, i interface{}) error
	JSONBlob(code int, b []byte) error
	JSONPretty(code int, i interface{}, indent string) error

	JSONP(code int, callback string, i interface{}) error
	JSONPBlob(code int, callback string, b []byte) error

	XML(code int, i interface{}) error
	XMLBlob(code int, b []byte) error
	XMLPretty(code int, i interface{}, indent string) error

	File(file string) error
	Inline(file string, name string) error
	Attachment(file string, name string) error
	Stream(code int, contentType string, r io.Reader) error
}
