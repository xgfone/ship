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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/xgfone/ship/v5/binder"
	"github.com/xgfone/ship/v5/router"
	"github.com/xgfone/ship/v5/router/echo"
)

var noop = func(interface{}) error { return nil }

// DefaultShip is the default global ship.
var DefaultShip = Default()

// Router is the alias of router.Router.
type Router = router.Router

// Ship is an app to be used to manage the router.
type Ship struct {
	// Name is the name of the ship.
	//
	// Default: ""
	Name string

	// Prefix is the default prefix of the paths of all the routes.
	//
	// Default: ""
	Prefix string

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

	// Router is the route manager to manage all the routes.
	//
	// Default: echo.NewRouter(nil)
	Router Router

	// The default handler when not finding the route.
	//
	// Default: NotFoundHandler()
	NotFound Handler

	// Filter the route if returning true when registering and unregistering it.
	//
	// Default: nil
	RouteFilter func(Route) bool

	// Modify the route before registering and unregistering it.
	//
	// Default: nil
	RouteModifier func(Route) Route

	// HandleError is used to handle the error at last
	// if the handler or middleware returns an error.
	//
	// Default: respond the error to the client if not responding.
	HandleError func(c *Context, err error)

	// Context Settings.
	Session    Session                                     // Default: NewMemorySession()
	Logger     Logger                                      // Default: NewLoggerFromWriter(os.Stderr, "")
	Binder     Binder                                      // Default: nil
	Renderer   Renderer                                    // Default: nil
	Responder  func(c *Context, args ...interface{}) error // Default: nil
	BindQuery  func(dst interface{}, src url.Values) error // Default: BindURLValues(dst, src, "query")
	SetDefault func(v interface{}) error                   // Default: SetStructFieldToDefault
	Validator  func(v interface{}) error                   // Default: nil

	mws     []Middleware
	pmws    []Middleware
	handler Handler
	bpool   sync.Pool
	cpool   sync.Pool
}

// New returns a new Ship.
func New() *Ship {
	s := &Ship{
		Router: echo.NewRouter(&echo.Config{RemoveTrailingSlash: true}),
	}

	s.handler = s.handleRequest
	s.cpool.New = func() interface{} { return s.NewContext() }

	s.URLParamMaxNum = 4
	s.MiddlewareMaxNum = 256

	s.Logger = NewLoggerFromWriter(os.Stderr, "")
	s.Session = NewMemorySession()
	s.NotFound = NotFoundHandler()
	s.HandleError = s.handleErrorDefault
	s.SetDefault = SetStructFieldToDefault
	s.Validator = noop
	s.BindQuery = func(v interface{}, vs url.Values) error {
		return binder.BindURLValues(v, vs, "query")
	}

	s.SetBufferSize(2048)
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

// Clone clones itself to a new one without routes and middlewares.
// Meanwhile, it will reset the signals of the new Ship to nil.
func (s *Ship) Clone() *Ship {
	newShip := &Ship{
		Router: echo.NewRouter(&echo.Config{RemoveTrailingSlash: true}),
	}

	// Private
	newShip.cpool.New = func() interface{} { return newShip.NewContext() }
	newShip.handler = newShip.handleRequest

	// Public
	newShip.Logger = s.Logger
	newShip.Prefix = s.Prefix
	newShip.NotFound = s.NotFound
	newShip.HandleError = s.HandleError
	newShip.RouteFilter = s.RouteFilter
	newShip.RouteModifier = s.RouteModifier
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

	newShip.SetBufferSize(2048)
	return newShip
}

// SetBufferSize resets the size of the buffer. The default is 2048.
func (s *Ship) SetBufferSize(size int) {
	s.bpool.New = func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, size))
	}
}

//----------------------------------------------------------------------------
// Context & Buffer
//----------------------------------------------------------------------------

// NewContext news a Context.
func (s *Ship) NewContext() *Context {
	c := newContext(s.URLParamMaxNum, s.CtxDataInitCap)
	c.buffer = s
	c.logger = s.Logger
	c.router = s.Router
	c.session = s.Session
	c.notFound = s.NotFound
	c.binder = s.Binder
	c.qbinder = s.BindQuery
	c.renderer = s.Renderer
	c.responder = s.Responder

	if s.SetDefault == nil {
		c.defaulter = noop
	} else {
		c.defaulter = s.SetDefault
	}

	if s.Validator == nil {
		c.validate = noop
	} else {
		c.validate = s.Validator
	}

	return c
}

// AcquireContext gets a Context from the pool.
func (s *Ship) AcquireContext(r *http.Request, w http.ResponseWriter) *Context {
	c := s.cpool.Get().(*Context)
	c.req, c.res.ResponseWriter = r, w
	return c
}

// ReleaseContext puts a Context into the pool.
func (s *Ship) ReleaseContext(c *Context) {
	c.Reset()
	s.cpool.Put(c)
}

// AcquireBuffer gets a Buffer from the pool.
func (s *Ship) AcquireBuffer() *bytes.Buffer {
	return s.bpool.Get().(*bytes.Buffer)
}

// ReleaseBuffer puts a Buffer into the pool.
func (s *Ship) ReleaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	s.bpool.Put(buf)
}

//----------------------------------------------------------------------------
// Middleware
//----------------------------------------------------------------------------

// ResetMiddlewares resets the global middlewares to mdws.
func (s *Ship) ResetMiddlewares(mdws ...Middleware) {
	s.mws = append([]Middleware{}, mdws...)
}

// ResetPreMiddlewares resets the global pre-middlewares to mdws.
func (s *Ship) ResetPreMiddlewares(mdws ...Middleware) {
	s.pmws = append([]Middleware{}, mdws...)
}

// Pre registers the pre-middlewares, which are executed before finding the route.
func (s *Ship) Pre(middlewares ...Middleware) {
	s.pmws = append(s.pmws, middlewares...)
	s.handler = s.handleRequest
	for i := len(s.pmws) - 1; i >= 0; i-- {
		s.handler = s.pmws[i](s.handler)
	}
}

// Use registers the global middlewares, which must be registered
// before adding the routes using these middlewares.
func (s *Ship) Use(middlewares ...Middleware) {
	s.mws = append(s.mws, middlewares...)
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

func (s *Ship) handleRequest(c *Context) error {
	h, n := c.router.Match(c.req.URL.Path, c.req.Method, c.pnames, c.pvalues)
	if h == nil {
		return c.notFound(c)
	}

	c.plen = n
	switch r := h.(type) {
	case Route:
		c.Route = r
	case Handler:
		c.Route.Handler = r
	default:
		panic(fmt.Errorf("unknown handler type '%T'", h))
	}

	return c.Route.Handler(c)
}

// ServeHTTP implements the interface http.Handler.
func (s *Ship) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := s.cpool.Get().(*Context)
	ctx.req, ctx.res.ResponseWriter = req, resp

	switch err := s.handler(ctx); err {
	case nil, ErrSkip:
	default:
		s.HandleError(ctx, err)
	}

	ctx.Reset()
	s.cpool.Put(ctx)
}
