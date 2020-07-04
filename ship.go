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
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/xgfone/ship/v3/binder"
	"github.com/xgfone/ship/v3/render"
	"github.com/xgfone/ship/v3/router"
	"github.com/xgfone/ship/v3/router/echo"
	"github.com/xgfone/ship/v3/session"
)

// DefaultMethodMapping is the default method mapping of the route.
var DefaultMethodMapping = map[string]string{
	"Create": "POST",
	"Delete": "DELETE",
	"Update": "PUT",
	"Get":    "GET",
}

// DefaultShip is the default global ship.
var DefaultShip = Default()

// Ship is an app to be used to manage the router.
type Ship struct {
	*Runner

	/// Context
	CtxDataSize int // The initialization size of Context.Data.

	/// Route, Handler and Middleware
	Prefix           string
	NotFound         Handler
	RouteFilter      func(RouteInfo) bool
	RouteModifier    func(RouteInfo) RouteInfo
	MethodMapping    map[string]string // The default is DefaultMethodMapping.
	MiddlewareMaxNum int               // Default is 256

	// If not nil, it will be locked and unlocked during access the routers.
	// So you can modify the routes concurrently and safely during running.
	//
	// Notice: if using the lock, you should also use the locked Router.
	Lock *sync.RWMutex

	// Others
	Logger      Logger
	Binder      binder.Binder
	Session     session.Session
	Renderer    render.Renderer
	BindQuery   func(interface{}, url.Values) error
	Responder   func(c *Context, args ...interface{}) error
	HandleError func(c *Context, err error)

	urlMaxNum   int32
	bufferPool  sync.Pool
	contextPool sync.Pool

	router    router.Router
	newRouter func() router.Router
	hrouters  map[string]router.Router

	handler        Handler
	middlewares    []Middleware
	premiddlewares []Middleware
}

// New returns a new Ship.
func New() *Ship {
	s := new(Ship)

	s.Runner = NewRunner("", s)
	s.Session = session.NewMemorySession()
	s.NotFound = NotFoundHandler()
	s.HandleError = s.handleErrorDefault
	s.MiddlewareMaxNum = 256

	s.SetBufferSize(2048)
	s.SetLogger(NewLoggerFromWriter(os.Stderr, ""))
	s.SetNewRouter(func() router.Router { return echo.NewRouter(nil, nil) })

	s.contextPool.New = func() interface{} { return s.NewContext() }
	s.hrouters = make(map[string]router.Router, 4)
	s.handler = s.handleRoute

	return s
}

// Default returns a new ship with default configuration, which will set Binder,
// Renderer and BindQuery to MuxBinder, MuxRenderer and BindURLValues based on
// New().
func Default() *Ship {
	mb := binder.NewMuxBinder()
	mb.Add(MIMEApplicationJSON, binder.JSONBinder())
	mb.Add(MIMETextXML, binder.XMLBinder())
	mb.Add(MIMEApplicationXML, binder.XMLBinder())
	mb.Add(MIMEMultipartForm, binder.FormBinder(MaxMemoryLimit))
	mb.Add(MIMEApplicationForm, binder.FormBinder(MaxMemoryLimit))

	mr := render.NewMuxRenderer()
	mr.Add("json", render.JSONRenderer())
	mr.Add("jsonpretty", render.JSONPrettyRenderer())
	mr.Add("xml", render.XMLRenderer())
	mr.Add("xmlpretty", render.XMLPrettyRenderer())

	s := New()
	s.Binder = mb
	s.Renderer = mr
	s.BindQuery = func(v interface{}, vs url.Values) error {
		return binder.BindURLValues(v, vs, "query")
	}

	return s
}

// Clone clones itself to a new one without routes, middlewares and the server.
// Meanwhile, it will reset the signals of the new Ship to nil.
func (s *Ship) Clone() *Ship {
	newShip := new(Ship)

	// Private
	newShip.handler = newShip.handleRoute
	newShip.hrouters = make(map[string]router.Router, 4)
	newShip.contextPool.New = func() interface{} { return newShip.NewContext() }

	// Public
	newShip.CtxDataSize = s.CtxDataSize
	newShip.Prefix = s.Prefix
	newShip.NotFound = s.NotFound
	newShip.RouteFilter = s.RouteFilter
	newShip.RouteModifier = s.RouteModifier
	newShip.MethodMapping = s.MethodMapping
	newShip.MiddlewareMaxNum = s.MiddlewareMaxNum
	newShip.Binder = s.Binder
	newShip.Session = s.Session
	newShip.Renderer = s.Renderer
	newShip.BindQuery = s.BindQuery
	newShip.Responder = s.Responder
	newShip.HandleError = s.HandleError

	newShip.SetBufferSize(2048)
	newShip.SetNewRouter(s.newRouter)

	if s.Runner != nil {
		newShip.Runner = NewRunner(s.Runner.Name, newShip)
		newShip.Runner.ConnState = s.Runner.ConnState
		newShip.Runner.Signals = nil
	}

	newShip.SetLogger(s.Logger)
	return newShip
}

func (s *Ship) lock() {
	if s.Lock != nil {
		s.Lock.Lock()
	}
}

func (s *Ship) unlock() {
	if s.Lock != nil {
		s.Lock.Unlock()
	}
}

func (s *Ship) rlock() {
	if s.Lock != nil {
		s.Lock.RLock()
	}
}

func (s *Ship) runlock() {
	if s.Lock != nil {
		s.Lock.RUnlock()
	}
}

//----------------------------------------------------------------------------
// Settings
//----------------------------------------------------------------------------

// SetBufferSize resets the size of the buffer.
func (s *Ship) SetBufferSize(size int) *Ship {
	s.bufferPool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, size))
	}
	return s
}

// SetNewRouter resets the NewRouter to create the new router.
//
// It must be called before adding any route.
func (s *Ship) SetNewRouter(f func() router.Router) *Ship {
	s.newRouter = f
	s.router = f()
	return s
}

// SetLogger sets the logger of Ship and Runner to logger.
func (s *Ship) SetLogger(logger Logger) *Ship {
	s.Logger = logger
	if s.Runner != nil {
		s.Runner.Logger = logger
	}
	return s
}

//----------------------------------------------------------------------------
// Context & Buffer
//----------------------------------------------------------------------------

// NewContext news a Context.
func (s *Ship) NewContext() *Context {
	c := NewContext(s.URLParamsMaxNum(), s.CtxDataSize)
	c.SetSessionManagement(s.Session)
	c.SetNotFoundHandler(s.NotFound)
	c.SetBufferAllocator(s)
	c.SetQueryBinder(s.BindQuery)
	c.SetResponder(s.Responder)
	c.SetRenderer(s.Renderer)
	c.SetBinder(s.Binder)
	c.SetLogger(s.Logger)
	return c
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) *Context {
	c := s.contextPool.Get().(*Context)
	c.SetReqRes(r, w)
	num := s.URLParamsMaxNum()
	if len(c.pnames) < num {
		c.pnames = make([]string, num)
		c.pvalues = make([]string, num)
	}
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c *Context) { c.Reset(); s.contextPool.Put(c) }

// AcquireBuffer gets a Buffer from the pool.
func (s *Ship) AcquireBuffer() *bytes.Buffer {
	return s.bufferPool.Get().(*bytes.Buffer)
}

// ReleaseBuffer puts a Buffer into the pool.
func (s *Ship) ReleaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	s.bufferPool.Put(buf)
}

//----------------------------------------------------------------------------
// Route & RouteGroup
//----------------------------------------------------------------------------

// Pre registers the Pre-middlewares, which are executed before finding the route.
// then returns the origin ship router to write the chained router.
func (s *Ship) Pre(middlewares ...Middleware) *Ship {
	s.premiddlewares = append(s.premiddlewares, middlewares...)

	handler := s.handleRoute
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

// Host returns a new sub-group with the virtual host.
func (s *Ship) Host(host string) *RouteGroup {
	return newRouteGroup(s, s.Prefix, "", host, s.middlewares...)
}

// Group returns a new sub-group.
func (s *Ship) Group(prefix string) *RouteGroup {
	return newRouteGroup(s, s.Prefix, prefix, "", s.middlewares...)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, nil, s.Prefix, "", path, s.middlewares...)
}

// R is short for Route(path).
func (s *Ship) R(path string) *Route { return s.Route(path) }

// URLParamsMaxNum reports the maximum number of the parameters of all the URLs.
//
// Notice: it should be only called after adding all the urls.
func (s *Ship) URLParamsMaxNum() int {
	return int(atomic.LoadInt32(&s.urlMaxNum))
}

// Routes returns the inforatiom of all the routes.
func (s *Ship) Routes() (routes []RouteInfo) {
	routes = make([]RouteInfo, 0, 64)
	for _, r := range s.router.Routes() {
		routes = append(routes, RouteInfo{
			Name:    r.Name,
			Path:    r.Path,
			Method:  r.Method,
			Handler: r.Handler.(Handler),
		})
	}

	s.rlock()
	defer s.runlock()
	for host, router := range s.hrouters {
		for _, r := range router.Routes() {
			routes = append(routes, RouteInfo{
				Host:    host,
				Name:    r.Name,
				Path:    r.Path,
				Method:  r.Method,
				Handler: r.Handler.(Handler),
			})
		}
	}

	return
}

// Routers returns the routers with their host.
//
// For the main router, the host is "".
func (s *Ship) Routers() (routers map[string]router.Router) {
	s.rlock()
	if _len := len(s.hrouters); _len == 0 {
		routers = map[string]router.Router{"": s.router}
	} else {
		routers = make(map[string]router.Router, _len*2)
		routers[""] = s.router
		for host, router := range s.hrouters {
			routers[host] = router
		}
	}
	s.runlock()
	return
}

// Router returns the Router implementation by the host name.
//
// If host is empty, return the main router.
func (s *Ship) Router(host string) router.Router {
	if host == "" {
		return s.router
	}

	s.rlock()
	r := s.hrouters[host]
	s.runlock()
	return r
}

// AddRoutes registers a set of the routes.
func (s *Ship) AddRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.AddRoute(ri); err != nil {
			panic(err)
		}
	}
}

// AddRoute registers the route, which uses the global middlewares to wrap
// the handler. If you don't want to use any middleware, you can do it by
//    s.Group("").NoMiddlewares().AddRoutes(ri)
//
// Notice: "Name" and "Host" are optional, "Router" will be ignored.
// and others are mandatory.
func (s *Ship) AddRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if !ok || err != nil {
		return
	} else if ri.Handler == nil {
		return errors.New("handler must not be nil")
	} else if ri.Method == "" {
		return errors.New("the route requires methods")
	}

	s.lock()
	defer s.unlock()

	router := s.router
	if ri.Host != "" {
		if r, ok := s.hrouters[ri.Host]; ok {
			router = r
		} else {
			router = s.newRouter()
			s.hrouters[ri.Host] = router
		}
	}

	num, err := router.Add(ri.Name, ri.Method, ri.Path, ri.Handler)
	if err != nil {
		err = RouteError{RouteInfo: ri, Err: err}
	} else if maxnum := s.URLParamsMaxNum(); num > maxnum {
		atomic.StoreInt32(&s.urlMaxNum, int32(num))
	}

	return
}

func (s *Ship) checkRouteInfo(ri *RouteInfo) (ok bool, err error) {
	ri.Method = strings.ToUpper(ri.Method)
	if s.RouteModifier != nil {
		*ri = s.RouteModifier(*ri)
	}

	if s.RouteFilter != nil && s.RouteFilter(*ri) {
		return
	}

	if err = ri.checkPath(); err == nil {
		ok = true
	}

	return
}

// DelRoute deletes the registered route.
//
// Only need "Name", "Path", "Method", but only "Path" is required
// and others are ignored.
//
// If Name is not empty, lookup the path by it instead of Path.
// If Method is empty, deletes all the routes associated with the path.
func (s *Ship) DelRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if !ok || err != nil {
		return
	}

	s.lock()
	defer s.unlock()

	router := s.router
	if ri.Host != "" {
		router = s.hrouters[ri.Host]
	}

	if router != nil {
		err = router.Del(ri.Name, ri.Method, ri.Path)
	}

	return
}

// DelRoutes deletes a set of the registered routes.
func (s *Ship) DelRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.DelRoute(ri); err != nil {
			panic(err)
		}
	}
}

//----------------------------------------------------------------------------
// Handle Request
//----------------------------------------------------------------------------

func (s *Ship) handleErrorDefault(ctx *Context, err error) {
	if !ctx.IsResponded() {
		switch e := err.(type) {
		case HTTPError:
			ctx.BlobText(e.Code, e.CT, e.GetMsg())
			if e.Code < 500 {
				return
			}
			err = e.Err
		default:
			ctx.NoContent(http.StatusInternalServerError)
		}
	}
}

func (s *Ship) handleRoute(c *Context) error { return c.Execute() }

func (s *Ship) routing(router router.Router, w http.ResponseWriter, r *http.Request) {
	ctx := s.AcquireContext(r, w)
	ctx.SetRouter(router)
	switch err := s.handler(ctx); err {
	case nil, ErrSkip:
	default:
		s.HandleError(ctx, err)
	}
	s.ReleaseContext(ctx)
}

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := s.router
	s.rlock()
	if r, ok := s.hrouters[r.Host]; ok {
		router = r
	}
	s.runlock()
	s.routing(router, w, r)
}
