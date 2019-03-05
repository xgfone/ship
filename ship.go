// Copyright 2018 xgfone
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
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/xgfone/ship/router/echo"
	"github.com/xgfone/ship/utils"
)

// Router stands for a router management.
type Router interface {
	// Generate a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add a route with name, path, method and handler,
	// and return the number of the parameters if there are the parameters
	// in the route. Or return 0.
	//
	// If the name has been added for the same path, it should be allowed.
	// Or it should panic.
	//
	// If the router does not support the parameter, it should panic.
	//
	// Notice: for keeping consistent, the parameter should start with ":"
	// or "*". ":" stands for a single parameter, and "*" stands for
	// a wildcard parameter.
	Add(name string, path string, method string, handler interface{}) (paramNum int)

	// Find a route handler by the method and path of the request.
	//
	// Return nil if the route does not exist.
	//
	// If the route has more than one parameter, the name and value
	// of the parameters should be stored `pnames` and `pvalues` respectively.
	Find(method string, path string, pnames []string, pvalues []string) (handler interface{})

	// Traverse each route.
	Each(func(name string, method string, path string))
}

// Resetter is an Reset interface.
type Resetter interface {
	Reset()
}

type stopT struct {
	once sync.Once
	f    func()
}

func (s *stopT) run() {
	s.once.Do(s.f)
}

var defaultSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGINT,
}

var defaultMethodMapping = map[string]string{
	"Create": "POST",
	"Delete": "DELETE",
	"Update": "PUT",
	"Get":    "GET",
}

// Ship is an app to be used to manage the router.
type Ship struct {
	/// Configuration Options
	name   string
	debug  bool
	prefix string

	logger   Logger
	binder   Binder
	session  Session
	renderer Renderer
	signals  []os.Signal

	bufferSize            int
	ctxDataSize           int
	middlewareMaxNum      int
	enableCtxHTTPContext  bool
	keepTrailingSlashPath bool
	defaultMethodMapping  map[string]string

	notFoundHandler         Handler
	optionsHandler          Handler
	methodNotAllowedHandler Handler

	isDefaultRouter bool
	disableErrorLog bool

	newRouter   func() Router
	newCtxData  func(*Context) Resetter
	handleError func(*Context, error)
	ctxHandler  func(*Context, ...interface{}) error
	bindQuery   func(interface{}, url.Values) error

	/// Inner settings
	ctxpool sync.Pool
	bufpool utils.BufferPool

	maxNum int
	router Router

	handler        Handler
	premiddlewares []Middleware
	middlewares    []Middleware

	links  []*Ship
	vhosts map[string]*Ship

	server *http.Server
	stopfs []*stopT
	once1  sync.Once // For shutdown
	once2  sync.Once // For stop
	done   chan struct{}
	lock   sync.RWMutex

	connState func(net.Conn, http.ConnState)
}

// New returns a new Ship.
func New(options ...Option) *Ship {
	s := new(Ship)

	/// Initialize the default configuration.
	s.logger = NewNoLevelLogger(os.Stderr)
	s.session = NewMemorySession()
	s.signals = defaultSignals
	mb := NewMuxBinder()
	mb.Add(MIMEApplicationJSON, JSONBinder())
	mb.Add(MIMETextXML, XMLBinder())
	mb.Add(MIMEApplicationXML, XMLBinder())
	mb.Add(MIMEMultipartForm, FormBinder())
	mb.Add(MIMEApplicationForm, FormBinder())
	s.binder = mb
	mr := NewMuxRenderer()
	mr.Add("json", JSONRenderer())
	mr.Add("jsonpretty", JSONPrettyRenderer("    "))
	mr.Add("xml", XMLRenderer())
	mr.Add("xmlpretty", XMLPrettyRenderer("    "))
	s.renderer = mr

	s.bufferSize = 2048
	s.middlewareMaxNum = 256
	s.defaultMethodMapping = defaultMethodMapping

	s.notFoundHandler = NotFoundHandler()

	s.handleError = s.handleErrorDefault
	s.bindQuery = func(v interface{}, d url.Values) error {
		return BindURLValues(v, d, "query")
	}
	s.newRouter = s.defaultNewRouter
	s.isDefaultRouter = true

	/// Initialize the inner variables.
	s.ctxpool.New = func() interface{} { return s.NewContext(nil, nil) }
	s.bufpool = utils.NewBufferPool(s.bufferSize)
	s.router = s.newRouter()
	s.handler = s.handleRequestRoute
	s.vhosts = make(map[string]*Ship)
	s.done = make(chan struct{}, 1)

	return s.Configure(options...)
}

func (s *Ship) defaultNewRouter() Router {
	var handleMethodNotAllowed, handleOptions func([]string) interface{}
	if s.methodNotAllowedHandler != nil {
		handleMethodNotAllowed = toRouterHandler(s.methodNotAllowedHandler)
	}
	if s.optionsHandler != nil {
		handleOptions = toRouterHandler(s.optionsHandler)
	}
	return echo.NewRouter(handleMethodNotAllowed, handleOptions)
}

func (s *Ship) clone() *Ship {
	newShip := Ship{
		// Configurations
		name:   s.name,
		debug:  s.debug,
		prefix: s.prefix,

		logger:   s.logger,
		binder:   s.binder,
		session:  s.session,
		renderer: s.renderer,
		signals:  s.signals,

		bufferSize:            s.bufferSize,
		ctxDataSize:           s.ctxDataSize,
		middlewareMaxNum:      s.middlewareMaxNum,
		keepTrailingSlashPath: s.keepTrailingSlashPath,
		defaultMethodMapping:  s.defaultMethodMapping,

		notFoundHandler:         s.notFoundHandler,
		optionsHandler:          s.optionsHandler,
		methodNotAllowedHandler: s.methodNotAllowedHandler,

		isDefaultRouter: s.isDefaultRouter,

		newRouter:   s.newRouter,
		newCtxData:  s.newCtxData,
		handleError: s.handleError,
		ctxHandler:  s.ctxHandler,
		bindQuery:   s.bindQuery,

		// Inner variables
		bufpool: utils.NewBufferPool(s.bufferSize),
		router:  s.newRouter(),
		handler: s.handleRequestRoute,
		vhosts:  make(map[string]*Ship),
		done:    make(chan struct{}, 1),
	}

	newShip.ctxpool.New = func() interface{} { return newShip.NewContext(nil, nil) }
	return &newShip
}

func (s *Ship) setURLParamNum(num int) {
	if num > s.maxNum {
		s.maxNum = num
	}
}

// Configure configures the Ship.
//
// Notice: the method must be called before starting the http server.
func (s *Ship) Configure(options ...Option) *Ship {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.server != nil {
		panic(fmt.Errorf("the http server has been started"))
	}

	for _, opt := range options {
		opt(s)
	}
	return s
}

// Clone returns a new Ship router with a new name by the current configuration.
//
// Notice: the new router will disable the signals and register the shutdown
// function into the parent Ship router.
func (s *Ship) Clone(name ...string) *Ship {
	newShip := s.clone()
	newShip.signals = []os.Signal{}
	if len(name) > 0 && name[0] != "" {
		newShip.name = name[0]
	}
	s.RegisterOnShutdown(newShip.shutdown)
	return newShip
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

	vhost := s.clone()
	vhost.vhosts = nil
	s.vhosts[host] = vhost
	return vhost
}

// Link links other to the current router, that's, only if either of the two
// routers is shutdown, another is also shutdown.
//
// Return the current router.
func (s *Ship) Link(other *Ship) *Ship {
	// Avoid to add each other repeatedly.
	for i := range s.links {
		if other == s.links[i] {
			return s
		}
	}

	s.links = append(s.links, other)
	other.links = append(other.links, s)
	return s
}

// NewContext news and returns a Context.
func (s *Ship) NewContext(r *http.Request, w http.ResponseWriter) *Context {
	return newContext(s, r, w, s.maxNum)
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) *Context {
	c := s.ctxpool.Get().(*Context)
	c.setReqResp(r, w)
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c *Context) {
	if c != nil {
		c.reset()
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

// Logger returns the inner Logger
func (s *Ship) Logger() Logger {
	return s.logger
}

// Renderer returns the inner Renderer.
func (s *Ship) Renderer() Renderer {
	return s.renderer
}

// MuxRenderer check whether the inner Renderer is MuxRenderer.
//
// If yes, return it as "*MuxRenderer"; or return nil.
func (s *Ship) MuxRenderer() *MuxRenderer {
	if mr, ok := s.renderer.(*MuxRenderer); ok {
		return mr
	}
	return nil
}

// Binder returns the inner Binder.
func (s *Ship) Binder() Binder {
	return s.binder
}

// MuxBinder check whether the inner Binder is MuxBinder.
//
// If yes, return it as "*MuxBinder"; or return nil.
func (s *Ship) MuxBinder() *MuxBinder {
	if mb, ok := s.binder.(*MuxBinder); ok {
		return mb
	}
	return nil
}

// Pre registers the Pre-middlewares, which are executed before finding the route.
// then returns the origin ship router to write the chained router.
func (s *Ship) Pre(middlewares ...Middleware) *Ship {
	s.premiddlewares = append(s.premiddlewares, middlewares...)

	handler := s.handleRequestRoute
	for i := len(s.premiddlewares) - 1; i >= 0; i-- {
		handler = s.premiddlewares[i](handler)
	}
	s.handler = handler

	return s
}

// Use registers the global middlewares and returns the origin ship router
// to write the chained router.
func (s *Ship) Use(middlewares ...Middleware) *Ship {
	s.middlewares = append(s.middlewares, middlewares...)
	return s
}

// Group returns a new sub-group.
func (s *Ship) Group(prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(s.middlewares)+len(middlewares))
	ms = append(ms, s.middlewares...)
	ms = append(ms, middlewares...)
	return newGroup(s, s.router, s.prefix, prefix, ms...)
}

// GroupWithoutMiddleware is the same as Group, but not inherit the middlewares of Ship.
func (s *Ship) GroupWithoutMiddleware(prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(middlewares))
	ms = append(ms, middlewares...)
	return newGroup(s, s.router, s.prefix, prefix, ms...)
}

// RouteWithoutMiddleware is the same as Route, but not inherit the middlewares of Ship.
func (s *Ship) RouteWithoutMiddleware(path string) *Route {
	return newRoute(s, s.router, s.prefix, path)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, s.router, s.prefix, path, s.middlewares...)
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
			vhost.serveHTTP(w, r)
			return
		}
	}
	s.serveHTTP(w, r)
}

func (s *Ship) handleRequestRoute(c *Context) error {
	if h := c.findHandler(c.req.Method, c.req.URL.Path); h != nil {
		return h(c)
	}
	return c.NotFoundHandler()(c)
}

func (s *Ship) serveHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := s.AcquireContext(r, w)
	ctx.router = s.router
	err := s.handler(ctx)

	if err == nil {
		err = ctx.Err
	}
	if err != nil {
		s.handleError(ctx, err)
	}
	s.ReleaseContext(ctx)
}

func (s *Ship) handleErrorDefault(ctx *Context, err error) {
	switch err {
	case nil, ErrSkip:
		return
	}

	if !ctx.IsResponded() {
		switch e := err.(type) {
		case HTTPError:
			if e.Code < 500 {
				if e.Msg == "" {
					if e.Err == nil {
						ctx.Blob(e.Code, e.CT, nil)
					} else {
						ctx.Blob(e.Code, e.CT, []byte(e.Err.Error()))
					}
				} else if e.Err == nil {
					ctx.Blob(e.Code, e.CT, []byte(e.Msg))
				} else {
					ctx.Blob(e.Code, e.CT, []byte(fmt.Sprintf("msg='%s', err='%s'", e.Msg, e.Err)))
				}
				return
			}
			ctx.Blob(e.Code, e.CT, []byte(e.Msg))
		default:
			ctx.NoContent(http.StatusInternalServerError)
			goto END
		}
	}

END:
	if !s.disableErrorLog {
		s.logger.Error("%s", err)
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

// RegisterOnShutdown registers some functions to run when the http server is
// shut down.
func (s *Ship) RegisterOnShutdown(functions ...func()) *Ship {
	s.lock.Lock()
	for _, f := range functions {
		s.stopfs = append(s.stopfs, &stopT{once: sync.Once{}, f: f})
	}
	s.lock.Unlock()
	return s
}

// SetConnStateHandler sets a handler to monitor the change of the connection
// state, which is used by the HTTP server.
func (s *Ship) SetConnStateHandler(h func(net.Conn, http.ConnState)) *Ship {
	s.lock.Lock()
	s.connState = h
	s.lock.Unlock()
	return s
}

// Start starts a HTTP server with addr.
//
// If tlsFile is not nil, it must be certFile and keyFile. That's,
//
//     router := ship.New()
//     rouetr.Start(addr, certFile, keyFile)
//
func (s *Ship) Start(addr string, tlsFiles ...string) *Ship {
	var cert, key string
	if len(tlsFiles) == 2 && tlsFiles[0] != "" && tlsFiles[1] != "" {
		cert = tlsFiles[0]
		key = tlsFiles[1]
	}
	s.startServer(&http.Server{Addr: addr}, cert, key)
	return s
}

// StartServer starts a HTTP server.
func (s *Ship) StartServer(server *http.Server) {
	s.startServer(server, "", "")
}

func (s *Ship) handleSignals(sigs ...os.Signal) {
	ss := make(chan os.Signal, 1)
	signal.Notify(ss, sigs...)
	for {
		<-ss
		s.shutdown()
		return
	}
}

func (s *Ship) runStop() {
	s.lock.RLock()
	defer s.lock.RUnlock()
	defer close(s.done)
	for _, r := range s.stopfs {
		r.run()
	}
}

func (s *Ship) stop() {
	s.once2.Do(s.runStop)
}

func (s *Ship) shutdown() {
	s.once1.Do(func() { s.Shutdown(context.Background()) })
}

func (s *Ship) startServer(server *http.Server, certFile, keyFile string) {
	defer s.shutdown()

	if s.vhosts == nil {
		s.logger.Error("forbid the virtual host to be started as a server")
		return
	}

	server.ErrorLog = log.New(s.logger.Writer(), "",
		log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if server.Handler == nil {
		server.Handler = s
	}

	// Handle the signal
	if len(s.signals) > 0 {
		go s.handleSignals(s.signals...)
	}

	for _, r := range s.links {
		s.RegisterOnShutdown(r.shutdown)
		r.RegisterOnShutdown(s.shutdown)
	}
	server.RegisterOnShutdown(s.stop)

	if server.ConnState == nil && s.connState != nil {
		server.ConnState = s.connState
	}

	var format string
	if s.name == "" {
		format = "The HTTP Server is shutdown"
		s.logger.Info("The HTTP Server is running on %s", server.Addr)
	} else {
		format = fmt.Sprintf("The HTTP Server [%s] is shutdown", s.name)
		s.logger.Info("The HTTP Server [%s] is running on %s",
			s.name, server.Addr)
	}

	s.lock.Lock()
	if s.server != nil {
		s.logger.Error(format + ": the server has been started")
		return
	}
	s.server = server
	s.lock.Unlock()

	var err error
	if certFile != "" && keyFile != "" {
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err == http.ErrServerClosed {
		s.logger.Info(format)
	} else {
		s.logger.Error(format+": %s", err)
	}
}

// Wait waits until all the registered shutdown functions have finished.
func (s *Ship) Wait() {
	<-s.done
}
