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
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
)

type kvalues struct {
	Key    string
	Values []string
}

// RouteBuilder is used to build a route.
type RouteBuilder struct {
	ship    *Ship
	group   *RouteGroupBuilder
	path    string
	name    string
	data    interface{}
	mdwares []Middleware
}

func newRouteBuilder(s *Ship, g *RouteGroupBuilder, prefix, path string,
	data interface{}, ms ...Middleware) *RouteBuilder {
	if path == "" {
		panic("the route path must not be empty")
	} else if path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	return &RouteBuilder{
		ship:    s,
		group:   g,
		path:    strings.TrimSuffix(prefix, "/") + path,
		mdwares: append([]Middleware{}, ms...),
		data:    data,
	}
}

// Route returns a new route builder.
func (s *Ship) Route(path string) *RouteBuilder {
	return newRouteBuilder(s, nil, s.Prefix, path, nil, s.mws...)
}

// Ship returns the ship that the current route is associated with.
func (r *RouteBuilder) Ship() *Ship { return r.ship }

// Clone clones a new route builder.
func (r *RouteBuilder) Clone() *RouteBuilder {
	return &RouteBuilder{
		data:    r.data,
		ship:    r.ship,
		path:    r.path,
		name:    r.name,
		group:   r.group,
		mdwares: append([]Middleware{}, r.mdwares...),
	}
}

// Group returns the route group builder that the current route belongs to,
// which maybe return nil.
func (r *RouteBuilder) Group() *RouteGroupBuilder { return r.group }

// Use appends some middlwares.
func (r *RouteBuilder) Use(middlewares ...Middleware) *RouteBuilder {
	r.mdwares = append(r.mdwares, middlewares...)
	return r
}

// ResetMiddlewares resets the middlewares to ms.
func (r *RouteBuilder) ResetMiddlewares(ms ...Middleware) *RouteBuilder {
	r.mdwares = append([]Middleware{}, ms...)
	return r
}

// Name sets the route name.
func (r *RouteBuilder) Name(name string) *RouteBuilder {
	r.name = name
	return r
}

// Data sets the context data.
func (r *RouteBuilder) Data(data interface{}) *RouteBuilder {
	r.data = data
	return r
}

func (r *RouteBuilder) newRoutes(name, path string, handler Handler,
	methods ...string) []Route {
	if len(methods) == 0 {
		return nil
	}

	middlewaresLen := len(r.mdwares)
	if middlewaresLen > r.ship.MiddlewareMaxNum {
		panic(fmt.Errorf("the number of middlewares '%d' has exceeded the maximum '%d'",
			middlewaresLen, r.ship.MiddlewareMaxNum))
	}

	for i := middlewaresLen - 1; i >= 0; i-- {
		handler = r.mdwares[i](handler)
	}

	routes := make([]Route, len(methods))
	for i, method := range methods {
		routes[i] = Route{
			Name:    name,
			Path:    path,
			Method:  method,
			Handler: handler,
			Data:    r.data,
		}
	}
	return routes
}

func (r *RouteBuilder) addRoute(name, path string, h Handler, ms ...string) {
	r.ship.AddRoutes(r.newRoutes(name, path, h, ms...)...)
}

// Routes builds and returns the routes.
func (r *RouteBuilder) Routes(handler Handler, methods ...string) []Route {
	return r.newRoutes(r.name, r.path, handler, methods...)
}

// Method registers the routes with the handler and methods.
//
// It will panic with it if there is an error when adding the routes.
func (r *RouteBuilder) Method(handler Handler, methods ...string) *RouteBuilder {
	r.ship.AddRoutes(r.Routes(handler, methods...)...)
	return r
}

// Any registers all the supported methods , which is short for
// r.Method(handler, "")
func (r *RouteBuilder) Any(handler Handler) *RouteBuilder {
	return r.Method(handler, "")
}

// CONNECT is the short for r.Method(handler, "CONNECT").
func (r *RouteBuilder) CONNECT(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodConnect)
}

// OPTIONS is the short for r.Method(handler, "OPTIONS").
func (r *RouteBuilder) OPTIONS(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodOptions)
}

// HEAD is the short for r.Method(handler, "HEAD").
func (r *RouteBuilder) HEAD(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodHead)
}

// PATCH is the short for r.Method(handler, "PATCH").
func (r *RouteBuilder) PATCH(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodPatch)
}

// TRACE is the short for r.Method(handler, "TRACE").
func (r *RouteBuilder) TRACE(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodTrace)
}

// GET is the short for r.Method(handler, "GET").
func (r *RouteBuilder) GET(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodGet)
}

// PUT is the short for r.Method(handler, "PUT").
func (r *RouteBuilder) PUT(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodPut)
}

// POST is the short for r.Method(handler, "POST").
func (r *RouteBuilder) POST(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodPost)
}

// DELETE is the short for r.Method(handler, "DELETE").
func (r *RouteBuilder) DELETE(handler Handler) *RouteBuilder {
	return r.Method(handler, http.MethodDelete)
}

// Redirect is used to redirect the request to toURL with 301, 302, 307 or 308.
func (r *RouteBuilder) Redirect(code int, toURL string) *RouteBuilder {
	var methods []string
	switch code {
	case 301, 302:
		methods = []string{http.MethodGet, http.MethodHead}
	case 307, 308:
		methods = []string{http.MethodPost}
	default:
		panic("redirect only support 301, 302, 307 or 308")
	}

	return r.Method(func(ctx *Context) error {
		return ctx.Redirect(code, toURL)
	}, methods...)
}

// Map registers a group of methods with handlers, which is equal to
//
//     for method, handler := range method2handlers {
//         r.Method(handler, method)
//     }
//
func (r *RouteBuilder) Map(method2handlers map[string]Handler) *RouteBuilder {
	for method, handler := range method2handlers {
		r.Method(handler, method)
	}
	return r
}

// StaticFS registers a route to serve a static filesystem.
func (r *RouteBuilder) StaticFS(fs http.FileSystem) *RouteBuilder {
	if strings.Contains(r.path, ":") || strings.Contains(r.path, "*") {
		panic(errors.New("URL parameters cannot be used when serving a static file"))
	}

	fileServer := http.StripPrefix(r.path, http.FileServer(fs))
	handler := func(c *Context) error {
		fileServer.ServeHTTP(c.res, c.req)
		return nil
	}
	r.addRoute("", path.Join(r.path, "/"), handler, http.MethodHead, http.MethodGet)
	r.addRoute("", path.Join(r.path, "/*"), handler, http.MethodHead, http.MethodGet)

	return r
}

// Static is the same as StaticFS, but listing the files for a directory.
func (r *RouteBuilder) Static(dirpath string) *RouteBuilder {
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
	return noDirFile{f}, nil
}

type noDirFile struct{ http.File }

func (f noDirFile) Readdir(count int) ([]os.FileInfo, error) { return nil, nil }

/// ----------------------------------------------------------------------- ///

// Remove removes the route.
//
// If the method is "", it will remove all the routes associated with the path.
func (r *RouteBuilder) Remove(method string) *RouteBuilder {
	r.ship.DelRoutes(Route{Name: r.name, Path: r.path, Method: method})
	return r
}

// RemoveAny is equal to r.Remove("").
func (r *RouteBuilder) RemoveAny() *RouteBuilder {
	return r.Remove("")
}

// RemoveGET is equal to r.Remove(http.MethodGet).
func (r *RouteBuilder) RemoveGET() *RouteBuilder {
	return r.Remove(http.MethodGet)
}

// RemovePUT is equal to r.Remove(http.MethodPut).
func (r *RouteBuilder) RemovePUT() *RouteBuilder {
	return r.Remove(http.MethodPut)
}

// RemovePOST is equal to r.Remove(http.MethodPost).
func (r *RouteBuilder) RemovePOST() *RouteBuilder {
	return r.Remove(http.MethodPost)
}

// RemoveHEAD is equal to r.Remove(http.MethodHead).
func (r *RouteBuilder) RemoveHEAD() *RouteBuilder {
	return r.Remove(http.MethodHead)
}

// RemovePATCH is equal to r.Remove(http.MethodPatch).
func (r *RouteBuilder) RemovePATCH() *RouteBuilder {
	return r.Remove(http.MethodPatch)
}

// RemoveDELETE is equal to r.Remove(http.MethodDelete).
func (r *RouteBuilder) RemoveDELETE() *RouteBuilder {
	return r.Remove(http.MethodDelete)
}

// RemoveCONNECT is equal to r.Remove(http.MethodConnect).
func (r *RouteBuilder) RemoveCONNECT() *RouteBuilder {
	return r.Remove(http.MethodConnect)
}

// RemoveOPTIONS is equal to r.Remove(http.MethodOptions).
func (r *RouteBuilder) RemoveOPTIONS() *RouteBuilder {
	return r.Remove(http.MethodOptions)
}

// RemoveTRACE is equal to r.Remove(http.MethodTrace).
func (r *RouteBuilder) RemoveTRACE() *RouteBuilder {
	return r.Remove(http.MethodTrace)
}
