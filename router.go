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
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config is used to configure the default router.
type Config struct {
	Prefix string

	Debug     bool
	CleanPath bool

	Logger   Logger
	Binder   Binder
	Renderer Renderer

	NewRoute     func() Route
	NewContext   func() Context
	NewURLParam  func(int) URLParam
	FilterOutput func([]byte) []byte

	HandleError func(ctx Context, err error)
	HandlePanic func(ctx Context, panicValue interface{})

	OptionsHandler          Handler
	NotFoundHandler         Handler
	MethodNotAllowedHandler Handler

	writerPool    *sync.Pool
	contextPool   *sync.Pool
	urlParamsPool *sync.Pool

	urlParamMaxNum int
}

type routeInfo struct {
	Name    string
	Path    string
	Method  string
	Handler Handler
}

type routerT struct {
	prefix string
	config *Config
	routes []routeInfo

	names map[string]Route
	trees map[string]Route

	beforeRouteHandler Handler
	afterRouteHandler  Handler
	beforeMiddlewares  []Middleware
	afterMiddlewares   []Middleware
	middlewares        []Middleware
}

func newRouter(config *Config, parent *routerT, prefix ...string) *routerT {
	var pre string
	if len(prefix) > 0 {
		pre = prefix[0]
	}

	router := routerT{
		prefix: pre,
		config: config,

		names: make(map[string]Route),
		trees: make(map[string]Route),

		beforeRouteHandler: NothingHandler,
		afterRouteHandler:  NothingHandler,
	}

	if parent != nil {
		router.beforeMiddlewares = parent.beforeMiddlewares[:]
		router.afterMiddlewares = parent.afterMiddlewares[:]
		router.middlewares = parent.middlewares[:]
	}

	return &router
}

// NewRouter returns a new Router.
func NewRouter(conf ...Config) Router {
	var config Config
	if len(conf) > 0 {
		config = conf[0]
	}

	if config.NewRoute == nil {
		config.NewRoute = NewRoute
	}
	if config.NewURLParam == nil {
		config.NewURLParam = NewURLParam
	}
	if config.NewContext == nil {
		config.NewContext = NewContext
	}
	if config.HandleError == nil {
		config.HandleError = HandleHTTPError
	}
	if config.Logger == nil {
		config.Logger = NewNoLevelLogger(os.Stdout)
	}
	if config.Binder == nil {
		config.Binder = NewBinder()
	}
	if config.NotFoundHandler == nil {
		config.NotFoundHandler = NotFoundHandler
	}
	config.writerPool = &sync.Pool{New: func() interface{} {
		return NewResponse(nil)
	}}
	config.contextPool = &sync.Pool{New: func() interface{} {
		return config.NewContext()
	}}
	config.urlParamsPool = &sync.Pool{New: func() interface{} {
		return config.NewURLParam(config.urlParamMaxNum)
	}}
	return newRouter(&config, nil, config.Prefix)
}

func (r *routerT) logErrorf(format string, args ...interface{}) {
	if r.config.Logger != nil {
		r.config.Logger.Error(format, args...)
	}
}

func (r *routerT) Before(ms ...Middleware) {
	r.beforeMiddlewares = append(r.beforeMiddlewares, ms...)

	var handler Handler = NothingHandler
	for i := len(r.beforeMiddlewares) - 1; i >= 0; i-- {
		handler = r.beforeMiddlewares[i].Handle(handler)
	}
	r.beforeRouteHandler = handler
}

func (r *routerT) After(ms ...Middleware) {
	r.afterMiddlewares = append(r.afterMiddlewares, ms...)

	var handler Handler = NothingHandler
	for i := len(r.afterMiddlewares) - 1; i >= 0; i-- {
		handler = r.afterMiddlewares[i].Handle(handler)
	}
	r.afterRouteHandler = handler
}

func (r *routerT) Use(ms ...Middleware) {
	r.middlewares = append(r.middlewares, ms...)
}

func (r *routerT) SubRouter(prefix ...string) Router {
	return newRouter(r.config, r, prefix...)
}

func (r *routerT) SubRouterNone(prefix ...string) Router {
	return newRouter(r.config, nil, prefix...)
}

func (r *routerT) Any(path string, handler Handler, name ...string) {
	r.Get(path, handler, name...)
	r.Put(path, handler, name...)
	r.Post(path, handler, name...)
	r.Head(path, handler, name...)
	r.Patch(path, handler, name...)
	r.Trace(path, handler, name...)
	r.Delete(path, handler, name...)
	r.Options(path, handler, name...)
	r.Connect(path, handler, name...)
}

func (r *routerT) Get(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodGet, path, handler, name...)
}

func (r *routerT) Put(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodPut, path, handler, name...)
}

func (r *routerT) Post(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodPost, path, handler, name...)
}

func (r *routerT) Head(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodHead, path, handler, name...)
}

func (r *routerT) Patch(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodPatch, path, handler, name...)
}

func (r *routerT) Trace(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodTrace, path, handler, name...)
}

func (r *routerT) Delete(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodDelete, path, handler, name...)
}

func (r *routerT) Options(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodOptions, path, handler, name...)
}

func (r *routerT) Connect(path string, handler Handler, name ...string) {
	r.addRoute(http.MethodConnect, path, handler, name...)
}

func (r *routerT) Methods(ms []string, p string, h Handler, name ...string) {
	for _, m := range ms {
		r.addRoute(m, p, h, name...)
	}
}

func (r *routerT) URL(name string, params URLParam) string {
	if route := r.names[name]; route != nil {
		return route.URL(name, params)
	}
	return ""
}

func (r *routerT) addRoute(method, path string, handler Handler, name ...string) {
	if path == "" {
		path = "/"
	}

	if path[0] != '/' {
		panic(fmt.Errorf("path must begin with '/' in path '%s'", path))
	} else if i := strings.Index(path, "//"); i != -1 {
		panic(fmt.Errorf("Bad path '%s' contains duplicate // at index:%s",
			path, strconv.Itoa(i)))
	}

	if handler == nil {
		panic(fmt.Errorf("Handler must not be nil"))
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i].Handle(handler)
	}

	method = strings.ToUpper(method)
	tree := r.trees[method]
	if tree == nil {
		tree = r.config.NewRoute()
		r.trees[method] = tree
	}

	var _name string
	if len(name) > 0 && name[0] != "" {
		_name = name[0]
	}

	if r.names[_name] != nil {
		panic(fmt.Errorf("the url name '%s' has been registered", _name))
	}

	num := tree.AddRoute(_name, method, path, handler)
	if num >= r.config.urlParamMaxNum {
		r.config.urlParamMaxNum = num + 1
	}

	if _name != "" {
		r.names[_name] = tree
	}

	info := routeInfo{Name: _name, Method: method, Path: path, Handler: handler}
	r.routes = append(r.routes, info)
}

func (r *routerT) Each(f func(name, method, path string, handler Handler)) {
	for i := range r.routes {
		f(r.routes[i].Name, r.routes[i].Method, r.routes[i].Path, r.routes[i].Handler)
	}
}

func (r *routerT) getURLParam() URLParam {
	return r.config.urlParamsPool.Get().(URLParam)
}

func (r *routerT) putURLParam(params URLParam) {
	if params != nil {
		params.Reset()
		r.config.urlParamsPool.Put(params)
	}
}

func (r *routerT) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	var ctx Context
	var params URLParam
	var handler Handler

	if r.config.HandlePanic != nil {
		defer func() {
			if err := recover(); err != nil {
				r.config.HandlePanic(ctx, err)
			}
		}()
	}

	path := req.URL.Path
	method := req.Method

	resp := r.config.writerPool.Get().(*Response).SetWriter(w).SetFilter(r.config.FilterOutput)
	ctx = r.config.contextPool.Get().(Context)
	ctx.Reset()
	ctx.SetRouter(r)
	ctx.SetReqResp(req, resp)
	if _, ok := ctx.(*contextT); !ok {
		ctx.SetDebug(r.config.Debug)
		ctx.SetLogger(r.config.Logger)
		ctx.SetBinder(r.config.Binder)
		ctx.SetRenderer(r.config.Renderer)
	}

	// Handle the pre-middlewares before routing.
	if err = r.beforeRouteHandler.Handle(ctx); err != nil {
		goto ERROR
	}

	if tree := r.trees[method]; tree != nil {
		if r.config.CleanPath {
			path = CleanPath(path)
		}

		if handler, params = tree.FindRoute(method, path, r.getURLParam); handler != nil {
			goto END
		}
	}

	// Handle OPTIONS requests
	if r.config.OptionsHandler != nil && method == http.MethodOptions {
		if path == "*" { // check server-wide OPTIONS
			for m := range r.trees {
				if m == http.MethodOptions {
					continue
				}
				w.Header().Add(HeaderAllow, m)
			}
		} else {
			for m, route := range r.trees {
				if m == method || m == http.MethodOptions {
					continue
				}
				if handler, params = route.FindRoute(m, path, r.getURLParam); handler != nil {
					w.Header().Add(HeaderAllow, m)
				}
			}
		}
		w.Header().Add(HeaderAllow, http.MethodOptions)
		handler = r.config.OptionsHandler
		goto END
	}

	// Check whether the method is not allowed.
	if r.config.MethodNotAllowedHandler != nil {
		var found bool
		for m, route := range r.trees {
			if m == method {
				continue
			}
			if handler, params = route.FindRoute(m, path, r.getURLParam); handler != nil {
				w.Header().Add(HeaderAllow, m)
				found = true
			}
		}
		if found {
			handler = r.config.MethodNotAllowedHandler
			goto END
		}
	}

	// Not Found
	handler = r.config.NotFoundHandler

END:
	if params != nil {
		ctx.SetRequest(SetURLParam(req, params))
		ctx.SetURLParam(params)
	}

	// Handle the post-middlewares after routing.
	if err = r.afterRouteHandler.Handle(ctx); err == nil {
		// Handle the request
		err = handler.Handle(ctx)
	}

ERROR:
	if err != nil {
		r.config.HandleError(ctx, err)
	}

	ctx.Reset()
	resp.Reset(nil)

	r.putURLParam(params)
	r.config.contextPool.Put(ctx)
	r.config.writerPool.Put(resp)
}
