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
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/xgfone/ship/binder"
	"github.com/xgfone/ship/core"
	"github.com/xgfone/ship/render"
	"github.com/xgfone/ship/router/echo"
	"github.com/xgfone/ship/utils"
)

var onceCall = sync.Once{}

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
//    Render(ctx Context, name string, code int, data interface{}) error
type Renderer = core.Renderer

// Matcher is used to check whether the request match some conditions.
type Matcher func(*http.Request) error

// Config is used to configure the router used by the default implementation.
type Config struct {
	// The name of the router, which is used when starting the http server.
	Name string

	// The route prefix, which is "" by default.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool

	// If true, it won't remove the trailing slash from the registered url path.
	KeepTrailingSlashPath bool

	// The size of the buffer initialized by the buffer pool.
	//
	// The default is 2KB.
	BufferSize int

	// The initializing size of the store, which is a map essentially,
	// used by the context.
	//
	// The default is 0. If you use the store, such as Get(), Set(), you should
	// set it to a appropriate value.
	ContextStoreSize int

	// The maximum number of the middlewares, which is 256 by default.
	MiddlewareMaxNum int

	// It is the default mapping to map the method into router. The default is
	//
	//     map[string]string{
	//         "Create": "POST",
	//         "Delete": "DELETE",
	//         "Update": "PUT",
	//         "Get":    "GET",
	//     }
	DefaultMethodMapping map[string]string

	// The signal set that built-in http server will wrap and handle.
	// The default is
	//
	//     []os.Signal{
	//         os.Interrupt,
	//         syscall.SIGTERM,
	//         syscall.SIGQUIT,
	//         syscall.SIGABRT,
	//         syscall.SIGINT,
	//     }
	//
	// In order to disable the signals, you can set it to []os.Signal{}.
	Signals []os.Signal

	// BindQuery binds the request query to v.
	BindQuery func(queries url.Values, v interface{}) error

	// The logger management, which is `NewNoLevelLogger(os.Stdout)` by default.
	// But you can appoint yourself customized Logger implementation.
	Logger Logger
	// Binder is used to bind the request data to the given value,
	// which is `NewBinder()` by default.
	// But you can appoint yourself customized Binder implementation
	Binder Binder
	// Rendered is used to render the response to the peer.
	//
	// The default is MuxRender, and adds some renderer, for example,
	// json, jsonpretty, xml, xmlpretty, etc, as follow.
	//
	//     renderer := NewMuxRender()
	//     renderer.Add("json", render.JSON())
	//     renderer.Add("jsonpretty", render.JSONPretty("    "))
	//     renderer.Add("xml", render.XML())
	//     renderer.Add("xmlpretty", render.XMLPretty("    "))
	//
	// So you can use it by the four ways:
	//
	//     renderer.Render(ctx, "json", 200, data)
	//     renderer.Render(ctx, "jsonpretty", 200, data)
	//     renderer.Render(ctx, "xml", 200, data)
	//     renderer.Render(ctx, "xmlpretty", 200, data)
	//
	// You can use the default, then add yourself renderer as follow.
	//
	///    router := New()
	//     mr := router.MuxRender()
	//     mr.Add("html", HtmlRenderer)
	//
	Renderer Renderer

	// Create a new router, which uses echo implementation by default.
	// But you can appoint yourself customized Router implementation.
	NewRouter func() Router

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

	if c.BufferSize <= 0 {
		c.BufferSize = 2048
	}

	if c.ContextStoreSize < 0 {
		c.ContextStoreSize = 0
	}

	if c.MiddlewareMaxNum <= 0 {
		c.MiddlewareMaxNum = 256
	}

	if c.DefaultMethodMapping == nil {
		c.DefaultMethodMapping = map[string]string{
			"Create": "POST",
			"Delete": "DELETE",
			"Update": "PUT",
			"Get":    "GET",
		}
	}

	if c.Signals == nil {
		c.Signals = []os.Signal{
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGABRT,
			syscall.SIGINT,
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

	if c.BindQuery == nil {
		c.BindQuery = binder.BindQuery
	}

	if c.Renderer == nil {
		mr := NewMuxRender()
		mr.Add("json", render.JSON())
		mr.Add("jsonpretty", render.JSONPretty("    "))
		mr.Add("xml", render.XML())
		mr.Add("xmlpretty", render.XMLPretty("    "))
		c.Renderer = mr
	}

	if c.NewRouter == nil {
		c.NewRouter = func() Router { return echo.NewRouter(c.MethodNotAllowedHandler, c.OptionsHandler) }
	}
}

// Ship is used to manage the router.
type Ship struct {
	config  Config
	ctxpool sync.Pool
	bufpool utils.BufferPool
	maxNum  int

	handler        Handler
	premiddlewares []Middleware
	middlewares    []Middleware

	router Router
	vhosts map[string]*Ship
	server *http.Server
	stopfs []func()
	links  []*Ship
	lock   *sync.RWMutex
}

// New returns a new Ship.
func New(config ...Config) *Ship {
	s := new(Ship)
	if len(config) > 0 {
		s.config = config[0]
	}

	s.config.init(s)
	s.handler = s.handleRequestMiddleware(NothingHandler())
	s.bufpool = utils.NewBufferPool(s.config.BufferSize)
	s.ctxpool.New = func() interface{} { return s.NewContext(nil, nil) }
	s.router = s.config.NewRouter()
	s.vhosts = make(map[string]*Ship)
	s.lock = new(sync.RWMutex)
	return s
}

func (s *Ship) setURLParamNum(num int) {
	if num > s.maxNum {
		s.maxNum = num
	}
}

// Config returns the inner config.
func (s *Ship) Config() Config {
	return s.config
}

// ResetConfig resets the config.
//
// You must not call it during the ship router is running.
func (s *Ship) ResetConfig(config Config) {
	config.init(s)
	s.config = config
}

// Clone returns a new Ship router with a new name by the current configuration.
//
// Notice: the new router will disable the signals and register the shutdown
// function into the parent Ship router.
func (s *Ship) Clone(name ...string) *Ship {
	config := s.config
	config.Signals = []os.Signal{}
	if len(name) > 0 && name[0] != "" {
		config.Name = name[0]
	}
	newShip := New(config)
	s.RegisterOnShutdown(func() { newShip.Shutdown(context.Background()) })
	return newShip
}

// Link links other to the current ship router, that's, other will be shutdown
// when the current router is shutdown. At last, return the current router.
//
// Notice: when calling other.Shutdown(), s will be shutdown.
func (s *Ship) Link(other *Ship) *Ship {
	s.links = append(s.links, other)
	return s
}

// LinkTo is equal to other.Link(s), but returns the current ship router s.
func (s *Ship) LinkTo(other *Ship) *Ship {
	other.Link(s)
	return s
}

// VHost returns a new ship used to manage the virtual host.
//
// For the different virtual host, you can register the same route.
//
// Notice: the new virtual host won't inherit anything except the configuration.
func (s *Ship) VHost(host string) *Ship {
	if s.vhosts == nil {
		panic(fmt.Errorf("the virtual host cannot create the virtual host"))
	}
	if s.vhosts[host] != nil {
		panic(fmt.Errorf("the virtual host '%s' has been added", host))
	}
	vhost := New(s.config)
	vhost.vhosts = nil
	s.vhosts[host] = vhost
	return vhost
}

// Logger returns the inner Logger
func (s *Ship) Logger() Logger {
	return s.config.Logger
}

// Renderer returns the inner Renderer.
func (s *Ship) Renderer() Renderer {
	return s.config.Renderer
}

// MuxRender check whether the inner Renderer is MuxRender.
//
// If yes, return it as "*MuxRender"; or return nil.
func (s *Ship) MuxRender() *MuxRender {
	if mr, ok := s.config.Renderer.(*MuxRender); ok {
		return mr
	}
	return nil
}

// NewContext news and returns a Context.
func (s *Ship) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return newContext(s, r, w, s.maxNum)
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) Context {
	c := s.ctxpool.Get().(*contextT)
	c.setReqResp(r, w)
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c Context) {
	if c != nil {
		c.(*contextT).reset()
		s.ctxpool.Put(c)
	}
}

// AcquireBuffer gets a Buffer from the pool.
func (s *Ship) AcquireBuffer() *bytes.Buffer {
	return s.bufpool.Get()
}

// ReleaseBuffer puts a Buffer into the pool.
func (s *Ship) ReleaseBuffer(buf *bytes.Buffer) {
	s.bufpool.Put(buf)
}

// Pre registers the Pre-middlewares, which are executed before finding the route.
func (s *Ship) Pre(middlewares ...Middleware) {
	s.premiddlewares = append(s.premiddlewares, middlewares...)

	handler := s.handleRequestMiddleware(NothingHandler())
	for i := len(s.premiddlewares) - 1; i >= 0; i-- {
		handler = s.premiddlewares[i](handler)
	}
	s.handler = handler
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
	return newGroup(s, s.router, s.config.Prefix, prefix, ms...)
}

// GroupWithoutMiddleware is the same as Group, but not inherit the middlewares of Ship.
func (s *Ship) GroupWithoutMiddleware(prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(middlewares))
	ms = append(ms, middlewares...)
	return newGroup(s, s.router, s.config.Prefix, prefix, ms...)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, s.router, s.config.Prefix, path, s.middlewares...)
}

// R is short for Ship#Route(path).
func (s *Ship) R(path string) *Route {
	return s.Route(path)
}

// Router returns the inner Router.
func (s *Ship) Router() Router {
	return s.router
}

// URL generates an URL from route name and provided parameters.
func (s *Ship) URL(name string, params ...interface{}) string {
	return s.router.URL(name, params...)
}

// Traverse traverses the registered route.
func (s *Ship) Traverse(f func(name string, method string, path string)) {
	s.router.Each(f)
}

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.vhosts != nil {
		if vhost := s.vhosts[r.Host]; vhost != nil {
			vhost.handleRequest(vhost.router, w, r)
			return
		}
	}

	s.handleRequest(s.router, w, r)
}

func (s *Ship) handleRequestMiddleware(next Handler) Handler {
	return func(ctx Context) error {
		c := ctx.(*contextT)
		h := c.router.Find(c.req.Method, c.req.URL.Path, c.pnames, c.pvalues)
		if h != nil {
			return h(ctx)
		}
		return s.config.NotFoundHandler(ctx)
	}
}

func (s *Ship) handleRequest(router Router, w http.ResponseWriter, r *http.Request) {
	ctx := s.AcquireContext(r, w).(*contextT)
	ctx.router = router
	err := s.handler(ctx)

	if err == nil {
		err = ctx.err
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
		if 400 <= code && code < 500 {
			msg = err.Error()
		} else if code >= 500 && ctx.IsDebug() {
			msg = err.Error()
		}

		ctx.Blob(code, ct, []byte(msg))

		if code >= 500 {
			if logger := ctx.Logger(); logger != nil {
				logger.Error("%s", err.Error())
			}
		}
		return
	}

	// For other errors, only log the error.
	if err != ErrSkip {
		ctx.NoContent(http.StatusInternalServerError)
		if logger := ctx.Logger(); logger != nil {
			logger.Error("%s", err.Error())
		}
	}
}

// Shutdown stops the HTTP server.
func (s *Ship) Shutdown(ctx context.Context) error {
	s.lock.RLock()
	server := s.server
	s.lock.RUnlock()

	if server == nil {
		return fmt.Errorf("the server has not been started")
	}
	return server.Shutdown(ctx)
}

// RegisterOnShutdown registers some functions to run
// when the http server is shut down.
func (s *Ship) RegisterOnShutdown(functions ...func()) {
	s.stopfs = append(s.stopfs, functions...)
}

// Start starts a HTTP server with addr.
//
// If tlsFile is not nil, it must be certFile and keyFile. That's,
//
//     router := ship.New()
//     rouetr.Start(addr, certFile, keyFile)
//
func (s *Ship) Start(addr string, tlsFiles ...string) error {
	var cert, key string
	if len(tlsFiles) == 2 && tlsFiles[0] != "" && tlsFiles[1] != "" {
		cert = tlsFiles[0]
		key = tlsFiles[1]
	}
	return s.startServer(&http.Server{Addr: addr}, cert, key)
}

// StartServer starts a HTTP server.
func (s *Ship) StartServer(server *http.Server) error {
	return s.startServer(server, "", "")
}

func (s *Ship) handleSignals(sigs ...os.Signal) {
	ss := make(chan os.Signal, 1)
	signal.Notify(ss, sigs...)
	for {
		<-ss
		s.server.Shutdown(context.Background())
		return
	}
}

func (s *Ship) stop() {
	for _, f := range s.stopfs {
		onceCall.Do(f)
	}
}

func (s *Ship) shutdown() {
	onceCall.Do(func() { s.Shutdown(context.Background()) })
}

func (s *Ship) startServer(server *http.Server, certFile, keyFile string) error {
	if s.vhosts == nil {
		return fmt.Errorf("forbid the virtual host to be started as a server")
	}

	server.ErrorLog = log.New(s.config.Logger.Writer(), "",
		log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if server.Handler == nil {
		server.Handler = s
	}

	// Handle the signal
	if len(s.config.Signals) > 0 {
		go s.handleSignals(s.config.Signals...)
	}

	for _, r := range s.links {
		server.RegisterOnShutdown(r.stop)
		r.RegisterOnShutdown(s.shutdown)
	}
	server.RegisterOnShutdown(s.stop)

	if s.config.Name == "" {
		s.config.Logger.Info("The HTTP Server is running on %s", server.Addr)
	} else {
		s.config.Logger.Info("The HTTP Server [%s] is running on %s",
			s.config.Name, server.Addr)
	}

	s.lock.Lock()
	s.server = server
	s.lock.Unlock()

	if certFile != "" && keyFile != "" {
		return server.ListenAndServeTLS(certFile, keyFile)
	}
	return server.ListenAndServe()
}
