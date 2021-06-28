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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/xgfone/ship/v4/binder"
	"github.com/xgfone/ship/v4/router"
	"github.com/xgfone/ship/v4/router/echo"
)

// DefaultShip is the default global ship.
var DefaultShip = Default()

// Router is the alias of router.Router.
type Router = router.Router

// Ship is an app to be used to manage the router.
type Ship struct {
	*Runner

	// Lock is used to access the host routers concurrently and thread-safely.
	//
	// Notice: It doesn't ensure that it's safe to access the routes
	// in a certain router concurrently and thread-safely.
	// But you maybe use the locked Router, such as router.LockRouter.
	//
	// Default: NewNoopRWLocker()
	Lock RWLocker

	// The initialization capacity of Context.Data.
	//
	// Default: 0
	CtxDataInitCap int

	// The maximum number of the url paramters of the route.
	//
	// Default: 4
	URLParamMaxNum int

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

	// Filter the route when registering and unregistering it.
	//
	// Default: nil
	RouteFilter func(RouteInfo) bool

	// Modify the route before registering and unregistering it.
	// Default: nil
	RouteModifier func(RouteInfo) RouteInfo

	// RouterExecutor is the router executor, which is called after matching
	// the host and before finding the route. By default, it only calls
	// Context.Execute().
	//
	// For the context, the executor can only use the field RouteInfo.Host.
	RouterExecutor Handler

	// HandleError is used to handle the error at last
	// if the handler or middleware returns an error.
	//
	// Default: respond the error to the client if not responding.
	HandleError func(c *Context, err error)

	// Others is used to set the context.
	Session    Session                                     // Default: NewMemorySession()
	Logger     Logger                                      // Default: NewLoggerFromWriter(os.Stderr, "")
	Binder     Binder                                      // Default: nil
	Renderer   Renderer                                    // Default: nil
	Responder  func(c *Context, args ...interface{}) error // Default: nil
	BindQuery  func(interface{}, url.Values) error         // Default: BindURLValues(v, vs, "query")
	SetDefault func(v interface{}) error                   // Default: SetStructFieldToDefault
	Validator  func(v interface{}) error                   // Default: nil

	defaultHost   string
	defaultRouter Router
	hostManager   *hostManager
	newReRouter   func() RegexpHostRouter
	newRouter     func() Router

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

	s.URLParamMaxNum = 4
	s.MiddlewareMaxNum = 256

	s.Lock = NewNoopRWLocker()
	s.Runner = NewRunner("", s)
	s.Session = NewMemorySession()
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
	s.SetNewRouter(func() Router { return echo.NewRouter(nil) })

	return s
}

// Default returns a new ship with MuxBinder and MuxRenderer
// as the binder and renderer.
func Default() *Ship {
	mb := NewMuxBinder()
	mb.Add(MIMEApplicationJSON, JSONBinder())
	mb.Add(MIMETextXML, XMLBinder())
	mb.Add(MIMEApplicationXML, XMLBinder())
	mb.Add(MIMEMultipartForm, FormBinder(MaxMemoryLimit))
	mb.Add(MIMEApplicationForm, FormBinder(MaxMemoryLimit))

	s := New()
	s.Binder = mb
	s.Renderer = NewMuxRenderer()

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
	newShip.Lock = NewNoopRWLocker()
	newShip.Prefix = s.Prefix
	newShip.NotFound = s.NotFound
	newShip.HandleError = s.HandleError
	newShip.RouteFilter = s.RouteFilter
	newShip.RouteModifier = s.RouteModifier
	newShip.RouterExecutor = s.RouterExecutor
	newShip.CtxDataInitCap = s.CtxDataInitCap
	newShip.URLParamMaxNum = s.URLParamMaxNum
	newShip.MiddlewareMaxNum = s.MiddlewareMaxNum

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
func (s *Ship) SetNewRouter(f func() Router) *Ship {
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
	c := NewContext(s.URLParamMaxNum, s.CtxDataInitCap)
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

// Routers returns the routers with their host.
func (s *Ship) Routers() (routers map[string]Router) {
	s.Lock.RLock()
	if _len := s.hostManager.Sum + 1; _len == 1 {
		routers = map[string]Router{s.defaultHost: s.defaultRouter}
	} else {
		routers = make(map[string]Router, _len)
		routers[s.defaultHost] = s.defaultRouter
		s.hostManager.Range(func(host string, router Router) {
			routers[host] = router
		})
	}
	s.Lock.RUnlock()
	return
}

// Router returns the Router implementation by the host name.
//
// If host is empty, return the default router.
func (s *Ship) Router(host string) (r Router) {
	s.Lock.RLock()
	if host == "" {
		r = s.defaultRouter
	} else if r = s.hostManager.Router(host); r == nil && host == s.defaultHost {
		r = s.defaultRouter
	}
	s.Lock.RUnlock()
	return
}

// SetDefaultRouter resets the default router with the host domain.
// If the router is nil, it lookups the router by the host, and uses
// the default host router instead if failing to lookup it.
//
// When matching the route, if no host router matches the request host,
// the default router will be used to find the route to handle the request.
func (s *Ship) SetDefaultRouter(host string, router Router) {
	s.Lock.Lock()
	s.defaultHost = host
	if router != nil {
		s.defaultRouter = router
	} else if router = s.hostManager.Router(host); router != nil {
		s.defaultRouter = router
	}
	s.Lock.Unlock()
}

// GetDefaultRouter returns the default host domain and router.
//
// For the default default router, the host is "".
func (s *Ship) GetDefaultRouter() (host string, router Router) {
	s.Lock.RLock()
	host, router = s.defaultHost, s.defaultRouter
	s.Lock.RUnlock()
	return
}

// AddHost adds the router with the host and returns it if it does not exist;
// or, do nothing and return the existed router.
//
// If router is nil, new one firstly.
func (s *Ship) AddHost(host string, r Router) (Router, error) {
	if host == "" {
		return nil, errors.New("the host must not be empty")
	} else if r == nil {
		r = s.newRouter()
	}

	s.Lock.Lock()
	r, err := s.hostManager.Add(host, r)
	s.Lock.Unlock()
	return r, err
}

// DelHost deletes the host router.
//
// If the host is empty or the host router does not exist, do nothing.
func (s *Ship) DelHost(host string) {
	if host != "" {
		s.Lock.Lock()
		s.hostManager.Del(host)
		s.Lock.Unlock()
	}
}

// Hosts returns all the hosts except for the default.
func (s *Ship) Hosts() (hosts []string) {
	s.Lock.RLock()
	hosts = make([]string, 0, s.hostManager.Sum)
	s.hostManager.Range(func(h string, _ Router) { hosts = append(hosts, h) })
	s.Lock.RUnlock()
	return
}

//----------------------------------------------------------------------------
// Handle Request
//----------------------------------------------------------------------------

func (s *Ship) handleErrorDefault(ctx *Context, err error) {
	if !ctx.res.Wrote {
		if se, ok := err.(HTTPServerError); !ok {
			ctx.NoContent(http.StatusInternalServerError)
		} else if se.CT == "" {
			ctx.BlobText(se.Code, MIMETextPlain, se.Error())
		} else {
			ctx.BlobText(se.Code, se.CT, se.Error())
		}
	}
}

func (s *Ship) executeRouter(c *Context) error {
	if s.RouterExecutor != nil {
		return s.RouterExecutor(c)
	}

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

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var host string
	var router Router

	s.Lock.RLock()
	if s.hostManager.Sum == 0 || req.Host == "" {
		host, router = s.defaultHost, s.defaultRouter
	} else if host, router = s.hostManager.Match(req.Host); host == "" {
		host, router = s.defaultHost, s.defaultRouter
	}
	s.Lock.RUnlock()

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
