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
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
)

// Route represents a route information.
type Route struct {
	ship    *Ship
	path    string
	name    string
	router  Router
	mdwares []Middleware

	matchers []Matcher
	headers  []string
	schemes  []string

	matcherM func([]Matcher) Middleware
}

func newRoute(s *Ship, router Router, prefix, path string, m ...Middleware) *Route {
	if !s.config.KeepTrailingSlashPath {
		path = strings.TrimSuffix(path, "/")
	}

	if len(path) == 0 {
		if len(prefix) == 0 {
			path = "/"
		}
	} else if path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	ms := make([]Middleware, 0, len(m))
	return &Route{
		ship: s,
		path: prefix + path,

		router:  router,
		mdwares: append(ms, m...),
	}
}

// New clones a new Route based on the current route.
func (r *Route) New() *Route {
	return &Route{
		ship:    r.ship,
		path:    r.path,
		name:    r.name,
		router:  r.router,
		mdwares: append([]Middleware{}, r.mdwares...),

		matchers: append([]Matcher{}, r.matchers...),
		headers:  append([]string{}, r.headers...),
		schemes:  append([]string{}, r.schemes...),
	}
}

// Name sets the route name.
func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

// Use adds some middlwares for the route.
func (r *Route) Use(middlewares ...Middleware) *Route {
	r.mdwares = append(r.mdwares, middlewares...)
	return r
}

// Match adds the matchers of the request to check whether the request matches
// these conditions.
//
// These matchers will be executes as the middlewares.
func (r *Route) Match(matchers ...Matcher) *Route {
	r.matchers = append(r.matchers, matchers...)
	return r
}

// MatchMiddleware sets the matcher middleware.
//
// The default implementation will execute those matchers in turn.
// If a certain matcher returns an error, it will return a HTTPError
// with 404 and the error.
func (r *Route) MatchMiddleware(f func([]Matcher) Middleware) *Route {
	r.matcherM = f
	return r
}

// Header checks whether the request contains the request header.
// If no, the request will be rejected.
//
// If the header value is given, it will be tested to match.
//
// Example
//
//     s := ship.New()
//     // The request must contains the header "Content-Type: application/json".
//     s.R("/path/to").HasHeader("Content-Type", "application/json").POST(handler)
//
// Notice: it is implemented by using Matcher.
func (r *Route) Header(headerK string, headerV ...string) *Route {
	var value string
	if len(headerV) > 0 {
		value = headerV[0]
	}
	r.headers = append(r.headers, http.CanonicalHeaderKey(headerK), value)
	return r
}

func (r *Route) buildHeadersMatcher() Matcher {
	if len(r.headers) == 0 {
		return nil
	}

	return func(req *http.Request) error {
		for i, _len := 0, len(r.headers); i < _len; i += 2 {
			key, value := r.headers[i], r.headers[i+1]
			if value != "" {
				if req.Header.Get(key) != value {
					return fmt.Errorf("missing the header '%s: %s'", key, value)
				}
			} else {
				if req.Header.Get(key) == "" {
					return fmt.Errorf("missing the header '%s'", key)
				}
			}
		}
		return nil
	}
}

// HasSchemes checks whether the request is one of the schemes.
// If no, the request will be rejected.
//
// Example
//
//     s := ship.New()
//     // We only handle https and wss, others will be rejected.
//     s.R("/path/to").HasSchemes("https", "wss").POST(handler)
//
// Notice: it is implemented by using Matcher.
func (r *Route) HasSchemes(schemes ...string) *Route {
	_len := len(schemes)
	if _len == 0 {
		return r
	}
	for i := 0; i < _len; i++ {
		schemes[i] = strings.ToLower(schemes[i])
	}
	r.schemes = append(r.schemes, schemes...)
	return r
}

func (r *Route) buildSchemesMatcher() Matcher {
	if len(r.schemes) == 0 {
		return nil
	}

	return func(req *http.Request) error {
		scheme := req.URL.Scheme
		for _, s := range r.schemes {
			if s == scheme {
				return nil
			}
		}
		return fmt.Errorf("not support the scheme '%s'", scheme)
	}
}

func (r *Route) buildMatcherMiddleware() Middleware {
	ms := []Matcher{
		r.buildHeadersMatcher(),
		r.buildSchemesMatcher(),
	}

	matchers := make([]Matcher, 0, len(r.matchers)+4)
	matchers = append(matchers, r.matchers...)
	for _, m := range ms {
		if m != nil {
			matchers = append(matchers, m)
		}
	}

	if len(matchers) == 0 {
		return nil
	}

	if r.matcherM != nil {
		return r.matcherM(matchers)
	}

	return func(next Handler) Handler {
		return func(ctx Context) (err error) {
			req := ctx.Request()
			for _, matcher := range matchers {
				if err = matcher(req); err != nil {
					return ErrNotFound.SetInnerError(err)
				}
			}
			return next(ctx)
		}
	}
}

func (r *Route) addRoute(name, path string, handler Handler, methods ...string) *Route {
	if handler == nil {
		panic(errors.New("handler must not be nil"))
	}

	if len(methods) == 0 {
		panic(errors.New("the route requires methods"))
	}

	if len(path) == 0 || path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	if i := strings.Index(path, "//"); i != -1 {
		panic(fmt.Errorf("bad path '%s' contains duplicate // at index:%d", path, i))
	}

	middlewares := r.mdwares
	matcherMiddleware := r.buildMatcherMiddleware()
	if matcherMiddleware != nil {
		middlewares = make([]Middleware, 0, len(r.mdwares)+1)
		middlewares = append(middlewares, r.mdwares...)
		middlewares = append(middlewares, matcherMiddleware)
	}

	middlewaresLen := len(middlewares)
	if middlewaresLen > r.ship.config.MiddlewareMaxNum {
		panic(fmt.Errorf("the number of middlewares '%d' has exceeded the maximum '%d'",
			middlewaresLen, r.ship.config.MiddlewareMaxNum))
	}

	for i := middlewaresLen - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	for i := range methods {
		n := r.router.Add(name, path, strings.ToUpper(methods[i]), handler)
		r.ship.setURLParamNum(n)
	}

	return r
}

// Method sets the methods and registers the route.
//
// If methods is nil, it will register all the supported methods for the route.
//
// Notice: The method must be called at last.
func (r *Route) Method(handler Handler, methods ...string) *Route {
	r.addRoute(r.name, r.path, handler, methods...)
	return r
}

// Any registers all the supported methods , which is short for
// r.Method(handler, "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE" )
func (r *Route) Any(handler Handler) *Route {
	return r.Method(handler, http.MethodConnect, http.MethodGet,
		http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodTrace)
}

// CONNECT is the short for r.Method(handler, "CONNECT").
func (r *Route) CONNECT(handler Handler) *Route {
	return r.Method(handler, http.MethodConnect)
}

// OPTIONS is the short for r.Method(handler, "OPTIONS").
func (r *Route) OPTIONS(handler Handler) *Route {
	return r.Method(handler, http.MethodOptions)
}

// HEAD is the short for r.Method(handler, "HEAD").
func (r *Route) HEAD(handler Handler) *Route {
	return r.Method(handler, http.MethodHead)
}

// PATCH is the short for r.Method(handler, "PATCH").
func (r *Route) PATCH(handler Handler) *Route {
	return r.Method(handler, http.MethodPatch)
}

// TRACE is the short for r.Method(handler, "TRACE").
func (r *Route) TRACE(handler Handler) *Route {
	return r.Method(handler, http.MethodTrace)
}

// GET is the short for r.Method(handler, "GET").
func (r *Route) GET(handler Handler) *Route {
	return r.Method(handler, http.MethodGet)
}

// PUT is the short for r.Method(handler, "PUT").
func (r *Route) PUT(handler Handler) *Route {
	return r.Method(handler, http.MethodPut)
}

// POST is the short for r.Method(handler, "POST").
func (r *Route) POST(handler Handler) *Route {
	return r.Method(handler, http.MethodPost)
}

// DELETE is the short for r.Method(handler, "DELETE").
func (r *Route) DELETE(handler Handler) *Route {
	return r.Method(handler, http.MethodDelete)
}

// Map registers a group of methods with handlers, which is equal to
//
//     for method, handler := range method2handlers {
//         r.Method(handler, method)
//     }
func (r *Route) Map(method2handlers map[string]Handler) *Route {
	for method, handler := range method2handlers {
		r.Method(handler, method)
	}
	return r
}

// MapType registers the methods of a type as the routes.
//
// By default, mapping is Ship.Config.DefaultMethodMapping if not given.
//
// Example
//
//    type TestType struct{}
//    func (t TestType) Create(ctx ship.Context) error { return nil }
//    func (t TestType) Delete(ctx ship.Context) error { return nil }
//    func (t TestType) Update(ctx ship.Context) error { return nil }
//    func (t TestType) Get(ctx ship.Context) error    { return nil }
//    func (t TestType) Has(ctx ship.Context) error    { return nil }
//    func (t TestType) NotHandler()                   {}
//
//    router := ship.New()
//    router.Route("/path/to").MapType(TestType{})
//
// It's equal to the operation as follow:
//
//    router.Route("/v1/testtype/get").Name("testtype_get").GET(ts.Get)
//    router.Route("/v1/testtype/update").Name("testtype_update").PUT(ts.Update)
//    router.Route("/v1/testtype/create").Name("testtype_create").POST(ts.Create)
//    router.Route("/v1/testtype/delete").Name("testtype_delete").DELETE(ts.Delete)
//
// If you don't like the default mapping policy, you can give the customized
// mapping by the last argument, the key of which is the name of the method
// of the type, and the value of that is the request method, such as GET, POST,
// etc. Notice that the method type must be compatible with
//
//    func (Context) error
//
// Notice: the name of type and method will be converted to the lower.
func (r *Route) MapType(tv interface{}, mapping ...map[string]string) *Route {
	if tv == nil {
		panic(errors.New("the type value must no be nil"))
	}

	value := reflect.ValueOf(tv)
	methodMaps := r.ship.config.DefaultMethodMapping
	if len(mapping) > 0 {
		methodMaps = mapping[0]
	}

	var err error
	errType := reflect.TypeOf(&err).Elem()
	prefix := r.path
	if prefix == "/" {
		prefix = ""
	}

	_type := value.Type()
	typeName := strings.ToLower(_type.Name())
	for i := _type.NumMethod() - 1; i >= 0; i-- {
		method := _type.Method(i)
		mtype := method.Type

		// func (s StructType) Handler(ctx Context) error
		if mtype.NumIn() != 2 || mtype.NumOut() != 1 {
			continue
		}
		if _, ok := reflect.New(mtype.In(1)).Interface().(*Context); !ok {
			continue
		}
		if !mtype.Out(0).Implements(errType) {
			continue
		}

		// r.addRoute(r.name, r.path, handler, methods...)
		if reqMethod := methodMaps[method.Name]; reqMethod != "" {
			methodName := strings.ToLower(method.Name)
			path := fmt.Sprintf("%s/%s/%s", prefix, typeName, methodName)

			name := fmt.Sprintf("%s_%s", typeName, methodName)
			r.addRoute(name, path, func(ctx Context) error {
				vs := method.Func.Call([]reflect.Value{value, reflect.ValueOf(ctx)})
				return vs[0].Interface().(error)
			}, reqMethod)
		}
	}

	return r
}

func (r *Route) serveFileMetadata(ctx Context, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError).SetInnerError(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError).SetInnerError(err)
	} else if fi.IsDir() {
		return ctx.NotFoundHandler()(ctx)
	}

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return NewHTTPError(http.StatusInternalServerError).SetInnerError(err)
	}

	ctx.SetHeader(HeaderEtag, fmt.Sprintf("%x", h.Sum(nil)))
	ctx.SetHeader(HeaderContentLength, fmt.Sprintf("%d", fi.Size()))
	return ctx.NoContent(http.StatusOK)
}

// StaticFile registers a route for a static file, which supports the HEAD method
// to get the its length and the GET method to download it.
func (r *Route) StaticFile(filePath string) *Route {
	if strings.Contains(r.path, ":") || strings.Contains(r.path, "*") {
		panic(errors.New("URL parameters cannot be used when serving a static file"))
	}

	r.addRoute("", r.path, func(ctx Context) error { return ctx.File(filePath) }, http.MethodGet)
	r.addRoute("", r.path, func(ctx Context) error { return r.serveFileMetadata(ctx, filePath) }, http.MethodHead)
	return r
}

// StaticFS registers a route to serve a static filesystem.
func (r *Route) StaticFS(fs http.FileSystem) *Route {
	if strings.Contains(r.path, ":") || strings.Contains(r.path, "*") {
		panic(errors.New("URL parameters cannot be used when serving a static file"))
	}

	fileServer := http.StripPrefix(r.path, http.FileServer(fs))
	rpath := path.Join(r.path, "/*filepath")

	r.addRoute("", rpath, func(ctx Context) error {
		filepath := ctx.Param("filepath")
		if _, err := fs.Open(filepath); err != nil {
			return ctx.NotFoundHandler()(ctx)
		}
		fileServer.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}, http.MethodHead, http.MethodGet)

	return r
}

// Static is the same as StaticFS, but listing the files for a directory.
func (r *Route) Static(dirpath string) *Route {
	return r.StaticFS(newOnlyFileFS(dirpath))
}

func newOnlyFileFS(root string) http.FileSystem {
	return onlyFileFS{fs: http.Dir(root)}
}

type onlyFileFS struct {
	fs http.FileSystem
}

func (fs onlyFileFS) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return notDirFile{f}, nil
}

type notDirFile struct {
	http.File
}

func (f notDirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
