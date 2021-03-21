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

	// If not nil, it will be locked and unlocked during access the routers.
	// So you can modify the routes concurrently and safely during running.
	//
	// Notice: if using the lock, you should also use the locked Router.
	//
	// Default: nil
	Lock *sync.RWMutex

	// The initialization capacity of Context.Data.
	//
	// Default: 0
	CtxDataSize int

	// The maximum size of the request body. And 0 represents no limit.
	//
	// Default: 0
	MaxBodySize int

	// The maximum number of the url paramters of the route.
	//
	// Default: 4
	MaxURLParamNum int

	// The maximum number of the middlewares.
	//
	// Default: 256
	MiddlewareMaxNum int

	// The prefix of the paths of all the routes.
	//
	// Default: ""
	Prefix string

	// The default handler when not finding the route.
	//
	// Default: NotFoundHandler()
	NotFound Handler

	// MethodMapping is used to map the struct method to register the routes.
	//
	// Default: DefaultMethodMapping.
	MethodMapping map[string]string

	// Filter the route when registering and unregistering it.
	//
	// Default: nil
	RouteFilter func(RouteInfo) bool

	// Modify the route before registering and unregistering it.
	// Default: nil
	RouteModifier func(RouteInfo) RouteInfo

	// RouteExecutor is the route executor, which is called after matching
	// the host and before finding the route. By default, it only calls
	// Context.Execute().
	//
	// For the context, the executor can only use the field RouteInfo.Host.
	RouteExecutor Handler

	// HandleError is used to handle the error at last
	// if the handler or middleware returns an error.
	//
	// Default: respond the error to the client if not responding.
	HandleError func(c *Context, err error)

	// Others is used to set the context.
	Logger     Logger                                      // Default: NewLoggerFromWriter(os.Stderr, "")
	Session    session.Session                             // Default: NewMemorySession()
	Binder     binder.Binder                               // Default: nil
	Renderer   render.Renderer                             // Default: nil
	Responder  func(c *Context, args ...interface{}) error // Default: nil
	BindQuery  func(interface{}, url.Values) error         // Default: binder.BindURLValues(v, vs, "query")
	SetDefault func(v interface{}) error                   // Default: SetStructFieldToDefault
	Validator  func(v interface{}) error                   // Default: nil

	defaultHost   string
	defaultRouter router.Router
	hostManager   *hostManager
	newReRouter   func() RegexpHostRouter
	newRouter     func() router.Router

	mws     []Middleware
	pmws    []Middleware
	handler Handler
	bpool   sync.Pool
	cpool   sync.Pool
}

// New returns a new Ship.
func New() *Ship {
	s := new(Ship)
	s.handler = s.executeRouter
	s.hostManager = newHostManager(nil)
	s.cpool.New = func() interface{} { return s.NewContext() }

	s.MaxURLParamNum = 4
	s.MiddlewareMaxNum = 256

	s.Runner = NewRunner("", s)
	s.Session = session.NewMemorySession()
	s.NotFound = NotFoundHandler()
	s.HandleError = s.handleErrorDefault
	s.SetDefault = SetStructFieldToDefault
	s.Validator = func(interface{}) error { return nil }
	s.BindQuery = func(v interface{}, vs url.Values) error {
		return binder.BindURLValues(v, vs, "query")
	}

	s.SetBufferSize(2048)
	s.SetLogger(NewLoggerFromWriter(os.Stderr, ""))
	s.SetNewRegexpHostRouter(NewRegexpHostRouter)
	s.SetNewRouter(func() router.Router { return echo.NewRouter(nil, nil) })

	return s
}

// Default returns a new ship with MuxBinder and MuxRenderer
// as the binder and renderer.
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

	return s
}

// Clone clones itself to a new one without routes, middlewares and the server.
// Meanwhile, it will reset the signals of the new Ship to nil.
func (s *Ship) Clone() *Ship {
	newShip := new(Ship)

	// Private
	newShip.newReRouter = s.newReRouter
	newShip.hostManager = newHostManager(s.newReRouter())
	newShip.cpool.New = func() interface{} { return newShip.NewContext() }
	newShip.handler = newShip.executeRouter

	// Public
	newShip.Prefix = s.Prefix
	newShip.NotFound = s.NotFound
	newShip.CtxDataSize = s.CtxDataSize
	newShip.HandleError = s.HandleError
	newShip.RouteFilter = s.RouteFilter
	newShip.RouteModifier = s.RouteModifier
	newShip.RouteExecutor = s.RouteExecutor
	newShip.MethodMapping = s.MethodMapping
	newShip.MaxURLParamNum = s.MaxURLParamNum
	newShip.MiddlewareMaxNum = s.MiddlewareMaxNum
	newShip.MaxBodySize = s.MaxBodySize

	// Context
	newShip.Binder = s.Binder
	newShip.Session = s.Session
	newShip.Renderer = s.Renderer
	newShip.BindQuery = s.BindQuery
	newShip.Validator = s.Validator
	newShip.Responder = s.Responder
	newShip.SetDefault = s.SetDefault

	if s.Runner != nil {
		newShip.Runner = NewRunner(s.Runner.Name, newShip)
		newShip.Runner.ConnState = s.Runner.ConnState
		newShip.Runner.Signals = nil
	}

	newShip.SetLogger(s.Logger)
	newShip.SetBufferSize(2048)
	newShip.SetNewRouter(s.newRouter)
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
	s.bpool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, size))
	}
	return s
}

// SetNewRouter resets the NewRouter to create the new router.
//
// It must be called before adding any route.
func (s *Ship) SetNewRouter(f func() router.Router) *Ship {
	s.defaultRouter = f()
	s.newRouter = f
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

// SetNewRegexpHostRouter is used to customize RegexpHostRouter.
func (s *Ship) SetNewRegexpHostRouter(f func() RegexpHostRouter) *Ship {
	s.hostManager.rhosts = f()
	s.newReRouter = f
	return s
}

//----------------------------------------------------------------------------
// Context & Buffer
//----------------------------------------------------------------------------

// NewContext news a Context.
func (s *Ship) NewContext() *Context {
	c := NewContext(s.MaxURLParamNum, s.CtxDataSize)
	c.SetSessionManagement(s.Session)
	c.SetNotFoundHandler(s.NotFound)
	c.SetBufferAllocator(s)
	c.SetQueryBinder(s.BindQuery)
	c.SetDefaulter(s.SetDefault)
	c.SetValidator(s.Validator)
	c.SetResponder(s.Responder)
	c.SetRenderer(s.Renderer)
	c.SetBinder(s.Binder)
	c.SetLogger(s.Logger)
	return c
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) *Context {
	c := s.cpool.Get().(*Context)
	c.req, c.res.ResponseWriter = r, w // c.SetReqRes(r, w)
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c *Context) { c.Reset(); s.cpool.Put(c) }

// AcquireBuffer gets a Buffer from the pool.
func (s *Ship) AcquireBuffer() *bytes.Buffer { return s.bpool.Get().(*bytes.Buffer) }

// ReleaseBuffer puts a Buffer into the pool.
func (s *Ship) ReleaseBuffer(buf *bytes.Buffer) { buf.Reset(); s.bpool.Put(buf) }

//----------------------------------------------------------------------------
// Route & RouteGroup
//----------------------------------------------------------------------------

// ResetMiddlewares resets the global middlewares to mdws.
func (s *Ship) ResetMiddlewares(mdws ...Middleware) *Ship {
	s.mws = append([]Middleware{}, mdws...)
	return s
}

// ResetPreMiddlewares resets the global pre-middlewares to mdws.
func (s *Ship) ResetPreMiddlewares(mdws ...Middleware) *Ship {
	s.pmws = append([]Middleware{}, mdws...)
	return s
}

// Pre registers the Pre-middlewares, which are executed before finding the route.
// then returns the origin ship router to write the chained router.
func (s *Ship) Pre(middlewares ...Middleware) *Ship {
	s.pmws = append(s.pmws, middlewares...)

	var handler Handler = s.executeRouter
	for i := len(s.pmws) - 1; i >= 0; i-- {
		handler = s.pmws[i](handler)
	}
	s.handler = handler

	return s
}

// Use registers the global middlewares and returns the origin ship router
// to write the chained router.
func (s *Ship) Use(middlewares ...Middleware) *Ship {
	s.mws = append(s.mws, middlewares...)
	return s
}

// Host returns a new sub-group with the virtual host.
func (s *Ship) Host(host string) *RouteGroup {
	return newRouteGroup(s, s.Prefix, "", host, s.mws...)
}

// Group returns a new sub-group.
func (s *Ship) Group(prefix string) *RouteGroup {
	return newRouteGroup(s, s.Prefix, prefix, "", s.mws...)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, nil, s.Prefix, "", path, nil, s.mws...)
}

// R is short for Route(path).
func (s *Ship) R(path string) *Route { return s.Route(path) }

// URLParamsMaxNum reports the maximum number of the parameters of all the URLs.
//
// Notice: it should be only called after adding all the urls.
//
// DEPRECATED!!! Please use MaxURLParamNum instead.
func (s *Ship) URLParamsMaxNum() int { return s.MaxURLParamNum }

func (s *Ship) getRoutes(host string, r router.Router, rs []RouteInfo) []RouteInfo {
	for _, r := range r.Routes() {
		ch := r.Handler.(RouteInfo)
		rs = append(rs, RouteInfo{
			Host:    host,
			Name:    r.Name,
			Path:    r.Path,
			Method:  r.Method,
			Handler: ch.Handler,
			CtxData: ch.CtxData,
		})
	}
	return rs
}

// Routes returns the information of all the routes.
func (s *Ship) Routes() (routes []RouteInfo) {
	s.rlock()
	nodefault := true
	routes = make([]RouteInfo, 0, s.hostManager.Sum+1)
	s.hostManager.Each(func(host string, router router.Router) {
		routes = s.getRoutes(host, router, routes)
		if nodefault && host == s.defaultHost {
			nodefault = false
		}
	})
	if nodefault {
		routes = s.getRoutes(s.defaultHost, s.defaultRouter, routes)
	}
	s.runlock()
	return
}

// Routers returns the routers with their host.
func (s *Ship) Routers() (routers map[string]router.Router) {
	s.rlock()
	if _len := s.hostManager.Len() + 1; _len == 1 {
		routers = map[string]router.Router{s.defaultHost: s.defaultRouter}
	} else {
		routers = make(map[string]router.Router, _len)
		routers[s.defaultHost] = s.defaultRouter
		s.hostManager.Each(func(host string, router router.Router) {
			routers[host] = router
		})
	}
	s.runlock()
	return
}

// Router returns the Router implementation by the host name.
//
// If host is empty, return the default router.
func (s *Ship) Router(host string) (r router.Router) {
	s.rlock()
	if host == "" {
		r = s.defaultRouter
	} else if r = s.hostManager.Router(host); r == nil && host == s.defaultHost {
		r = s.defaultRouter
	}
	s.runlock()
	return
}

// SetDefaultRouter resets the default router with the host domain.
//
// If no host router matches the request host, use the default router
// to find the route handler to handle the request.
func (s *Ship) SetDefaultRouter(host string, router router.Router) {
	if router == nil {
		panic("Ship.SetDefaultRouter: router must not be nil")
	}
	s.lock()
	s.defaultHost, s.defaultRouter = host, router
	s.unlock()
}

// GetDefaultRouter returns the default host domain and router.
//
// For the default default router, the host is "".
func (s *Ship) GetDefaultRouter() (host string, router router.Router) {
	s.rlock()
	host, router = s.defaultHost, s.defaultRouter
	s.runlock()
	return
}

// AddHost adds and returns the new host router.
//
// If existed, return it and do nothing. If router is nil, new one firstly.
func (s *Ship) AddHost(host string, r router.Router) (router.Router, error) {
	if host == "" {
		return nil, errors.New("the host must not be empty")
	} else if r == nil {
		r = s.newRouter()
	}

	s.lock()
	r, err := s.hostManager.Add(host, r)
	s.unlock()
	return r, err
}

// DelHost deletes the host router.
func (s *Ship) DelHost(host string) {
	if host != "" {
		s.lock()
		s.hostManager.Del(host)
		s.unlock()
	}
}

// AddRoutes registers a set of the routes.
func (s *Ship) AddRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.AddRoute(ri); err != nil {
			panic(err)
		}
	}
}

// DelRoutes deletes a set of the registered routes.
func (s *Ship) DelRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.DelRoute(ri); err != nil {
			panic(err)
		}
	}
}

// AddRoute registers the route.
//
// Only "Path", "Method" and "Handler" are mandatory, and others are optional.
func (s *Ship) AddRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if err != nil {
		return RouteError{RouteInfo: ri, Err: err}
	} else if !ok {
		return
	} else if ri.Handler == nil {
		return RouteError{RouteInfo: ri, Err: errors.New("handler must not be nil")}
	}

	var router router.Router
	s.lock()
	if ri.Host == "" {
		router = s.defaultRouter
	} else if router = s.hostManager.Router(ri.Host); router == nil {
		if ri.Host == s.defaultHost {
			router = s.defaultRouter
		} else {
			router, err = s.hostManager.Add(ri.Host, s.newRouter())
		}
	}
	s.unlock()

	if err != nil {
		return RouteError{RouteInfo: ri, Err: err}
	} else if n, e := router.Add(ri.Name, ri.Method, ri.Path, ri); e != nil {
		err = RouteError{RouteInfo: ri, Err: e}
	} else if n > s.MaxURLParamNum {
		router.Del(ri.Name, ri.Method, ri.Path)
		err = RouteError{RouteInfo: ri, Err: errors.New("too many url params")}
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

// DelRoute deletes the registered route, which only needs "Host", "Name",
// "Path" and "Method", and others are ignored.
//
// If Name is not empty, lookup the path by it instead of Path.
// If Method is empty, deletes all the routes associated with the path.
func (s *Ship) DelRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if !ok || err != nil {
		return
	}

	var r router.Router
	s.lock()
	if ri.Host == "" {
		r = s.defaultRouter
	} else if r = s.hostManager.Router(ri.Host); r == nil && ri.Host == s.defaultHost {
		r = s.defaultRouter
	}
	s.unlock()

	if r != nil {
		if err = r.Del(ri.Name, ri.Method, ri.Path); err != nil {
			err = RouteError{RouteInfo: ri, Err: err}
		}
	}

	return
}

//----------------------------------------------------------------------------
// Handle Request
//----------------------------------------------------------------------------

func (s *Ship) handleErrorDefault(ctx *Context, err error) {
	if !ctx.res.Wrote {
		switch e := err.(type) {
		case HTTPServerError:
			ctx.BlobText(e.Code, e.CT, e.Error())
		default:
			ctx.NoContent(http.StatusInternalServerError)
		}
	}
}

func (s *Ship) executeRouter(c *Context) error {
	if s.RouteExecutor != nil {
		return s.RouteExecutor(c)
	}

	h, n := c.router.Find(c.req.Method, c.req.URL.Path, c.pnames, c.pvalues)
	if h == nil {
		return c.notFound(c)
	}

	c.plen, c.RouteInfo = n, h.(RouteInfo)
	return c.RouteInfo.Handler(c)
}

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if s.MaxBodySize > 0 && req.ContentLength > int64(s.MaxBodySize) {
		resp.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	var host string
	var router router.Router

	if s.Lock == nil {
		if s.hostManager.Sum == 0 || req.Host == "" {
			host, router = s.defaultHost, s.defaultRouter
		} else if host, router = s.hostManager.Match(req.Host); host == "" {
			host, router = s.defaultHost, s.defaultRouter
		}
	} else {
		s.Lock.RLock()
		if s.hostManager.Sum == 0 || req.Host == "" {
			host, router = s.defaultHost, s.defaultRouter
		} else if host, router = s.hostManager.Match(req.Host); host == "" {
			host, router = s.defaultHost, s.defaultRouter
		}
		s.Lock.RUnlock()
	}

	// Optimize the function call, which is equal to
	//
	//   ctx := s.AcquireContext(req, resp)
	//   ctx.SetRouter(router)
	//   ctx.RouteInfo.Host = host
	//
	ctx := s.cpool.Get().(*Context)
	ctx.req, ctx.res.ResponseWriter = req, resp // ctx.SetReqRes(req, resp)
	ctx.router = router                         // ctx.SetRouter(router)
	ctx.RouteInfo.Host = host

	// s.executeRouter(ctx)
	switch err := s.handler(ctx); err {
	case nil, ErrSkip:
	default:
		s.HandleError(ctx, err)
	}

	// Optimize the function call, which is equal to
	//
	//   s.ReleaseContext(ctx)
	//
	ctx.Reset()
	s.cpool.Put(ctx)
}
