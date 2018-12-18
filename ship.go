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
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/xgfone/ship/binder"
	"github.com/xgfone/ship/core"
	"github.com/xgfone/ship/router/echo"
)

// Router is the alias of core.Router, which is used to manage the routes.
//
// Methods:
//   URL(name string, params ...interface{}) string
//   Add(name string, path string, method string, handler Handler) (paramNum int)
//   Find(method string, path string, pnames []string, pvalues []string) (handler Handler)
//   Each(func(name string, method string, path string))
type Router = core.Router

// Binder is the alias of core.Binder, which is used to bind the request
// to v.
//
// Methods:
//   Bind(ctx Context, v interface{}) error
type Binder = core.Binder

// Renderer is the alias of core.Renderer, which is used to render the response.
//
// Methods:
//    Render(ctx Context, w io.Writer, name string, code int, data interface{}) error
type Renderer = core.Renderer

// Config is used to configure the router used by the default implementation.
type Config struct {
	// The route prefix, which is "" by default.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool

	// If true, it won't remove the trailing slash from the registered url path.
	KeepTrailingSlashPath bool

	// It is the default mapping to map the method into router. The default is
	//
	//     map[string]string{
	//         "Create": "POST",
	//         "Delete": "DELETE",
	//         "Update": "PUT",
	//         "Get":    "GET",
	//     }
	DefaultMethodMapping map[string]string

	// The router management, which uses echo implementation by default.
	// But you can appoint yourself customized Router implementation.
	Router Router

	// The logger management, which is `NewNoLevelLogger(os.Stdout)` by default.
	// But you can appoint yourself customized Logger implementation.
	Logger Logger
	// Binder is used to bind the request data to the given value,
	// which is `NewBinder()` by default.
	// But you can appoint yourself customized Binder implementation
	Binder Binder
	// Rendered is used to render the response to the peer, which has no
	// the default implementation.
	Renderer Renderer

	// Handle the error at last.
	//
	// The default will send the response to the peer if the error is a HTTPError.
	// Or only log it. So the handler and the middleware return a HTTPError,
	// instead of sending the response to the peer.
	HandleError func(Context, error)

	// You can appoint the NotFound handler. The default is NotFoundHandler().
	NotFoundHandler Handler

	// OPTIONS and MethodNotAllowed handler, which are used for the default router.
	OptionsHandler          Handler
	MethodNotAllowedHandler Handler
}

func (c *Config) init(s *Ship) {
	c.Prefix = strings.TrimSuffix(c.Prefix, "/")

	if c.DefaultMethodMapping == nil {
		c.DefaultMethodMapping = map[string]string{
			"Create": "POST",
			"Delete": "DELETE",
			"Update": "PUT",
			"Get":    "GET",
		}
	}

	if c.Logger == nil {
		c.Logger = NewNoLevelLogger(os.Stdout)
	}

	if c.NotFoundHandler == nil {
		c.NotFoundHandler = NotFoundHandler()
	}

	if c.HandleError == nil {
		c.HandleError = s.handleError
	}

	if c.Binder == nil {
		c.Binder = binder.NewBinder()
	}

	if c.Router == nil {
		c.Router = echo.NewRouter(c.MethodNotAllowedHandler, c.OptionsHandler)
	}
}

// Ship is used to manage the router.
type Ship struct {
	config  Config
	ctxpool sync.Pool
	maxNum  int

	prehandler     Handler
	premiddlewares []Middleware
	middlewares    []Middleware
}

// New returns a new Ship.
func New(config ...Config) *Ship {
	s := new(Ship)
	if len(config) > 0 {
		s.config = config[0]
	}

	s.config.init(s)
	s.prehandler = NothingHandler()
	s.ctxpool.New = func() interface{} { return s.NewContext(nil, nil) }
	return s
}

// Logger returns the inner Logger
func (s *Ship) Logger() Logger {
	return s.config.Logger
}

// NewContext news and returns a Context.
func (s *Ship) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return newContext(s, r, w, s.maxNum)
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) Context {
	c := s.ctxpool.Get().(*context)
	c.setReqResp(r, w)
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c Context) {
	if c != nil {
		c.(*context).reset()
		s.ctxpool.Put(c)
	}
}

// Pre registers the Pre-middlewares, which are executed before finding the route.
func (s *Ship) Pre(middlewares ...Middleware) {
	s.premiddlewares = append(s.premiddlewares, middlewares...)

	handler := NothingHandler()
	for i := len(s.premiddlewares) - 1; i >= 0; i-- {
		handler = s.premiddlewares[i](handler)
	}
	s.prehandler = handler
}

// Use registers the global middlewares.
func (s *Ship) Use(middlewares ...Middleware) {
	s.middlewares = append(s.middlewares, middlewares...)
}

// Group returns a new sub-group.
func (s *Ship) Group(prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(s.middlewares)+len(middlewares))
	ms = append(ms, s.middlewares...)
	ms = append(ms, middlewares...)
	return newGroup(s, s.config.Prefix, prefix, ms...)
}

// GroupNone is the same as Group, but not inherit the middlewares of Ship.
func (s *Ship) GroupNone(prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(middlewares))
	ms = append(ms, middlewares...)
	return newGroup(s, s.config.Prefix, prefix, ms...)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, s.config.Prefix, path, s.middlewares...)
}

// R is short for Ship#Route(path).
func (s *Ship) R(path string) *Route {
	return s.Route(path)
}

func (s *Ship) addRoute(name, path string, methods []string, handler Handler, mws ...Middleware) {
	if handler == nil {
		panic(errors.New("handler must not be nil"))
	}

	if len(methods) == 0 {
		panic(errors.New("method must not be empty"))
	}

	if len(path) == 0 || path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	if i := strings.Index(path, "//"); i != -1 {
		panic(fmt.Errorf("bad path '%s' contains duplicate // at index:%d", path, i))
	}

	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}

	for i := range methods {
		if n := s.config.Router.Add(name, path, strings.ToUpper(methods[i]), handler); n > s.maxNum {
			s.maxNum = n
		}
	}
}

// Router returns the inner Router.
func (s *Ship) Router() Router {
	return s.config.Router
}

// URL generates an URL from route name and provided parameters.
func (s *Ship) URL(name string, params ...interface{}) string {
	return s.config.Router.URL(name, params...)
}

// Traverse traverses the registered route.
func (s *Ship) Traverse(f func(name string, method string, path string)) {
	s.config.Router.Each(f)
}

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var ctx = s.AcquireContext(r, w).(*context)

	if err = s.prehandler(ctx); err == nil {
		h := s.config.Router.Find(r.Method, r.URL.Path, ctx.pnames, ctx.pvalues)
		if h == nil {
			err = s.config.NotFoundHandler(ctx)
		} else {
			err = h(ctx)
		}
	}

	if err != nil {
		s.config.HandleError(ctx, err)
	}
	s.ReleaseContext(ctx)
}

func (s *Ship) handleError(ctx Context, err error) {
	// Handle the HTTPError, and send the response
	if he, ok := err.(HTTPError); ok {
		code := he.Code()
		ct := he.ContentType()
		msg := he.Message()
		if ctx.IsDebug() {
			msg = err.Error()
		} else if len(msg) == 0 {
			msg = http.StatusText(code)
		}

		if ie := he.InnerError(); ie != nil {
			err = fmt.Errorf("%s, %s", err.Error(), ie.Error())
		}

		ctx.Blob(code, ct, []byte(msg))
	}

	if err != ErrSkip {
		// For other errors, only log the error.
		if logger := ctx.Logger(); logger != nil {
			logger.Error("%s", err.Error())
		}
	}
}
