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

// Name sets the route name.
func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

// Headers adds some header matches.
//
// If the headers of a certain request don't contain these headers,
// it will return ship.config.NotFoundHandler.
//
// Example
//
//     s := ship.New()
//     s.R("/path/to").Headers("Content-Type", "application/json").POST(handler)
//
func (r *Route) Headers(headers ...string) *Route {
	_len := len(headers)
	if _len == 0 {
		return r
	} else if _len%2 != 0 {
		panic(errors.New("the number of the headers must be even"))
	}

	for i := 0; i < _len; i += 2 {
		headers[i] = http.CanonicalHeaderKey(headers[i])
	}

	return r.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			header := ctx.Request().Header
			for i := 0; i < _len; i += 2 {
				if header.Get(headers[i]) != headers[i+1] {
					return r.ship.config.NotFoundHandler(ctx)
				}
			}
			return next(ctx)
		}
	})
}

// Schemes adds some scheme matches.
//
// If the scheme of a certain request is not in these schemes,
// it will return ship.config.NotFoundHandler.
//
// Example
//
//     s := ship.New()
//     s.R("/path/to").Schemes("https", "wss").POST(handler)
//
func (r *Route) Schemes(schemes ...string) *Route {
	_len := len(schemes)
	if _len == 0 {
		return r
	}
	for i := 0; i < _len; i++ {
		schemes[i] = strings.ToLower(schemes[i])
	}

	return r.Use(func(next Handler) Handler {
		return func(ctx Context) error {
			scheme := ctx.Request().URL.Scheme
			for i := 0; i < _len; i++ {
				if schemes[i] == scheme {
					return next(ctx)
				}
			}
			return r.ship.config.NotFoundHandler(ctx)
		}
	})
}

// Use adds some middlwares for the route.
func (r *Route) Use(middlewares ...Middleware) *Route {
	r.mdwares = append(r.mdwares, middlewares...)
	return r
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

	for i := len(r.mdwares) - 1; i >= 0; i-- {
		handler = r.mdwares[i](handler)
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
func (r *Route) StaticFile(filePath string) {
	if strings.Contains(r.path, ":") || strings.Contains(r.path, "*") {
		panic(errors.New("URL parameters cannot be used when serving a static file"))
	}

	r.addRoute("", r.path, func(ctx Context) error { return ctx.File(filePath) }, http.MethodGet)
	r.addRoute("", r.path, func(ctx Context) error { return r.serveFileMetadata(ctx, filePath) }, http.MethodHead)
}
