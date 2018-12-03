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
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Logger stands for a logger.
type Logger interface {
	Debug(foramt string, args ...interface{})
	Info(foramt string, args ...interface{})
	Warn(foramt string, args ...interface{})
	Error(foramt string, args ...interface{})
}

// HTTPError stands for a HTTP error.
type HTTPError interface {
	Code() int
	Message() string
	Error() string
	InnerError() error
	SetInnerError(error) HTTPError
}

// URLParam is the interface of request scoped variables tracked by sea.
type URLParam interface {
	Reset()

	Get(name string) (value string)
	Set(name string, value string)
	Each(func(name string, value string))
}

// Binder is the interface to bind the value to v from ctx.
type Binder interface {
	Bind(ctx Context, v interface{}) error
}

// Renderer is the interface to render the response.
type Renderer interface {
	Render(ctx Context, w io.Writer, code int, name string, data interface{}) error
}

// Context stands for a request & response context.
type Context interface {
	Reset()

	Router() Router
	SetRouter(router Router)

	Request() *http.Request
	Response() http.ResponseWriter
	SetRequest(req *http.Request)
	SetReqResp(req *http.Request, resp http.ResponseWriter)

	IsDebug() bool
	SetDebug(debug bool)

	Logger() Logger
	SetLogger(logger Logger)

	Bind(v interface{}) error
	SetBinder(binder Binder)

	Render(code int, name string, data interface{}) error
	SetRenderer(renderer Renderer)

	// Manage the key-value in the context.
	Get(key string) interface{}
	Set(key string, value interface{})

	IsTLS() bool
	IsWebSocket() bool

	Scheme() string
	RealIP() string

	URLParam() URLParam
	URLParamByName(name string) string
	SetURLParam(params URLParam)

	QueryParam(name string) string
	QueryParams() url.Values
	QueryString() string

	FormFile(name string) (*multipart.FileHeader, error)
	FormValue(name string) string

	FormParams() (url.Values, error)
	MultipartForm() (*multipart.Form, error)

	Cookie(name string) (*http.Cookie, error)
	Cookies() []*http.Cookie
	SetCookie(cookie *http.Cookie)

	URL(name string, params URLParam) string
	Redirect(code int, toURL string) error

	HTML(code int, htmlData string) error
	HTMLBlob(code int, htmlData []byte) error

	JSON(code int, data interface{}) error
	JSONBlob(code int, data []byte) error
	JSONPretty(code int, data interface{}, indent string) error
	JSONP(code int, callback string, data interface{}) error
	JSONPBlob(code int, callback string, data []byte) error

	XML(code int, data interface{}) error
	XMLBlob(code int, data []byte) error
	XMLPretty(code int, data interface{}, indent string) error

	File(file string) error
	Inline(file string, name string) error
	Attachment(file string, name string) error
	Stream(code int, contentType string, data io.Reader) error

	NoContent(code int) error
	String(code int, data string) error
	Blob(code int, contentType string, data []byte) error
}

// Handler is a handler of the HTTP request.
type Handler func(Context) error

// Middleware stands for a middleware.
type Middleware func(Handler) Handler

// Route is used to manage the route.
type Route interface {
	AddRoute(name string, method string, path string, handler Handler) (paramMaxNum int)
	FindRoute(method string, path string, newURLParam func() URLParam) (Handler, URLParam)
	URL(name string, params URLParam) string
}

// Router stands for a router.
type Router interface {
	Before(middlewares ...Middleware)
	After(middlewares ...Middleware)
	Use(middlewares ...Middleware)

	SubRouter(prefix ...string) Router
	SubRouterNone(prefix ...string) Router

	Any(path string, handler Handler, name ...string)
	Get(path string, handler Handler, name ...string)
	Put(path string, handler Handler, name ...string)
	Post(path string, handler Handler, name ...string)
	Head(path string, handler Handler, name ...string)
	Patch(path string, handler Handler, name ...string)
	Trace(path string, handler Handler, name ...string)
	Delete(path string, handler Handler, name ...string)
	Options(path string, handler Handler, name ...string)
	Connect(path string, handler Handler, name ...string)
	Methods(methods []string, path string, handler Handler, name ...string)
	Each(func(name string, method string, path string, handler Handler))

	URL(name string, params URLParam) string

	ServeHTTP(resp http.ResponseWriter, req *http.Request)
}

// ToHTTPHandler converts the Handler to http.Handler
//
// Notice: Debug(), Logger(), Binder(), Render() and URLParam() can't be used
// until executing the following settings:
//
//    ctx.SetDebug(bool)
//    ctx.SetLogger(Logger)
//    ctx.SetBinder(Binder)
//    ctx.SetRenderer(Renderer)
//    ctx.SetURLParam(URLParam)
func ToHTTPHandler(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext()
		ctx.SetReqResp(r, w)
		h(ctx)
	})
}

// FromHTTPHandler converts http.Handler to Handler.
func FromHTTPHandler(h http.HandlerFunc) Handler {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

// FromHTTPHandlerFunc converts http.HandlerFunc to Handler.
func FromHTTPHandlerFunc(h http.HandlerFunc) Handler {
	return func(ctx Context) error {
		h(ctx.Response(), ctx.Request())
		return nil
	}
}

// NothingHandler does nothing.
func NothingHandler(ctx Context) error { return nil }

// NotFoundHandler is the default NotFound handler.
func NotFoundHandler(ctx Context) error {
	http.NotFound(ctx.Response(), ctx.Request())
	return nil
}

// MethodNotAllowedHandler is the default MethodNotAllowed handler.
func MethodNotAllowedHandler(ctx Context) error {
	ctx.Response().WriteHeader(http.StatusMethodNotAllowed)
	return nil
}

// OptionsHandler is the default OPTIONS handler.
func OptionsHandler(ctx Context) error {
	ctx.Response().WriteHeader(http.StatusOK)
	return nil
}

// HandlePanic wraps and logs the panic information.
func HandlePanic(ctx Context, err interface{}) {
	logger := ctx.Logger()
	if logger != nil {
		logger.Error("panic: %v", err)
	}
}

// HandleHTTPError handles the HTTP error.
func HandleHTTPError(ctx Context, err error) {
	var code = http.StatusInternalServerError
	var msg string

	if he, ok := err.(HTTPError); ok {
		code = he.Code()
		msg = he.Message()
		if ie := he.InnerError(); ie != nil {
			err = fmt.Errorf("%s, %s", err.Error(), ie.Error())
		}
	} else if ctx.IsDebug() {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}

	// Send response
	if resp, ok := ctx.Response().(*Response); ok && !resp.Committed {
		if ctx.Request().Method == http.MethodHead { // Issue #608
			resp.WriteHeader(code)
			err = nil
		} else {
			if resp.Header().Get(MIMETextPlain) == "" {
				resp.Header().Set(MIMETextPlain, msg)
			}
			resp.WriteHeader(code)
			_, err = resp.Write([]byte(msg))
		}
		if err != nil && ctx.Logger() != nil {
			ctx.Logger().Error("%s", err.Error())
		}
	}
}
