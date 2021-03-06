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

// Route represents a route information.
type Route struct {
	ship    *Ship
	group   *RouteGroup
	host    string
	path    string
	name    string
	mdwares []Middleware
	headers []kvalues
	ctxdata interface{}
}

func newRoute(s *Ship, g *RouteGroup, prefix, host, path string,
	ctxdata interface{}, ms ...Middleware) *Route {
	if path == "" {
		panic("the route path must not be empty")
	} else if path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	return &Route{
		ship:    s,
		group:   g,
		host:    host,
		path:    strings.TrimSuffix(prefix, "/") + path,
		mdwares: append([]Middleware{}, ms...),
		ctxdata: ctxdata,
	}
}

// Route returns a new route, which is used to build and register the route.
//
// You should call Route.Method() or its short method to register it.
func (s *Ship) Route(path string) *Route {
	return newRoute(s, nil, s.Prefix, "", path, nil, s.mws...)
}

// New clones a new Route based on the current route.
func (r *Route) New() *Route {
	return &Route{
		ship:  r.ship,
		host:  r.host,
		path:  r.path,
		name:  r.name,
		group: r.group,

		mdwares: append([]Middleware{}, r.mdwares...),
		headers: append([]kvalues{}, r.headers...),
	}
}

// Ship returns the ship that the current route is associated with.
func (r *Route) Ship() *Ship { return r.ship }

// Group returns the group that the current route belongs to.
//
// Notice: it will return nil if the route is from ship.Route.
func (r *Route) Group() *RouteGroup { return r.group }

// Clone closes itself and returns a new one.
func (r *Route) Clone() *Route { nr := *r; return &nr }

// NoMiddlewares clears all the middlewares and returns itself.
func (r *Route) NoMiddlewares() *Route { r.mdwares = nil; return r }

// Name sets the route name.
func (r *Route) Name(name string) *Route { r.name = name; return r }

// Host sets the host of the route to host.
func (r *Route) Host(host string) *Route { r.host = host; return r }

// CtxData sets the context data on the route.
func (r *Route) CtxData(data interface{}) *Route { r.ctxdata = data; return r }

// Use adds some middlwares for the route.
func (r *Route) Use(middlewares ...Middleware) *Route {
	r.mdwares = append(r.mdwares, middlewares...)
	return r
}

// HasHeader checks whether the request contains the request header.
// If no, the request will be rejected.
//
// If the header value is given, it will be tested to match, for example
//
//     s := ship.New()
//     // The request must contains the header "Content-Type: application/json".
//     s.R("/path/to").HasHeader("Content-Type", "application/json").POST(handler)
//
func (r *Route) HasHeader(headerK string, headerV ...string) *Route {
	r.headers = append(r.headers, kvalues{http.CanonicalHeaderKey(headerK), headerV})
	return r
}

func (r *Route) buildHeaderMiddleware() Middleware {
	if len(r.headers) == 0 {
		return nil
	}

	return func(next Handler) Handler {
		return func(ctx *Context) error {
			for _, kv := range r.headers {
				if len(kv.Values) == 0 {
					if ctx.GetHeader(kv.Key) == "" {
						err := fmt.Errorf("missing the header '%s'", kv.Key)
						return ErrBadRequest.New(err)
					}
				} else {
					values := ctx.ReqHeader()[kv.Key]
					for _, v := range kv.Values {
						for _, rv := range values {
							if v == rv {
								return next(ctx)
							}
						}
					}
					err := fmt.Errorf("invalid header '%s: %v'", kv.Key, values)
					return ErrBadRequest.New(err)
				}
			}
			return next(ctx)
		}
	}
}

func (r *Route) toRouteInfo(name, host, path string, handler Handler, methods ...string) []RouteInfo {
	if len(methods) == 0 {
		return nil
	}

	middlewares := r.mdwares
	if m := r.buildHeaderMiddleware(); m != nil {
		middlewares = make([]Middleware, 0, len(r.mdwares)+1)
		middlewares = append(middlewares, r.mdwares...)
		middlewares = append(middlewares, m)
	}

	middlewaresLen := len(middlewares)
	if middlewaresLen > r.ship.MiddlewareMaxNum {
		panic(fmt.Errorf("the number of middlewares '%d' has exceeded the maximum '%d'",
			middlewaresLen, r.ship.MiddlewareMaxNum))
	}

	for i := middlewaresLen - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	routes := make([]RouteInfo, len(methods))
	for i, method := range methods {
		routes[i] = RouteInfo{
			Host:    host,
			Name:    name,
			Path:    path,
			Method:  method,
			Handler: handler,
			CtxData: r.ctxdata,
		}
	}
	return routes
}

func (r *Route) addRoute(name, host, path string, h Handler, ms ...string) {
	r.ship.AddRoutes(r.toRouteInfo(name, host, path, h, ms...)...)
}

// RouteInfo converts the context routes to []RouteInfo.
func (r *Route) RouteInfo(handler Handler, methods ...string) []RouteInfo {
	return r.toRouteInfo(r.name, r.host, r.path, handler, methods...)
}

// Method registers the routes with the handler and methods.
//
// It will panic with it if there is an error when adding the routes.
func (r *Route) Method(handler Handler, methods ...string) *Route {
	r.ship.AddRoutes(r.RouteInfo(handler, methods...)...)
	return r
}

// Any registers all the supported methods , which is short for
// r.Method(handler, "")
func (r *Route) Any(handler Handler) *Route {
	return r.Method(handler, "")
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

// Redirect is used to redirect the path to toURL.
//
// method is GET by default.
func (r *Route) Redirect(code int, toURL string, method ...string) *Route {
	rmethod := http.MethodGet
	if len(method) > 0 && method[0] != "" {
		rmethod = method[0]
	}

	return r.Method(func(ctx *Context) error {
		return ctx.Redirect(code, toURL)
	}, rmethod)
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

// StaticFS registers a route to serve a static filesystem.
func (r *Route) StaticFS(fs http.FileSystem) *Route {
	if strings.Contains(r.path, ":") || strings.Contains(r.path, "*") {
		panic(errors.New("URL parameters cannot be used when serving a static file"))
	}

	fileServer := http.StripPrefix(r.path, http.FileServer(fs))
	r.addRoute("", r.host, path.Join(r.path, "/*"), func(c *Context) error {
		fileServer.ServeHTTP(c.res, c.req)
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

type notDirFile struct{ http.File }

func (f notDirFile) Readdir(count int) ([]os.FileInfo, error) { return nil, nil }

/// ----------------------------------------------------------------------- ///

// Remove removes the route.
//
// If the method is "", it will remove all the routes associated with the path.
func (r *Route) Remove(method string) *Route {
	r.ship.DelRoutes(RouteInfo{
		Host:   r.host,
		Name:   r.name,
		Path:   r.path,
		Method: method,
	})
	return r
}

// RemoveAny is equal to r.Remove("").
func (r *Route) RemoveAny() *Route { return r.Remove("") }

// RemoveGET is equal to r.Remove(http.MethodGet).
func (r *Route) RemoveGET() *Route { return r.Remove(http.MethodGet) }

// RemovePUT is equal to r.Remove(http.MethodPut).
func (r *Route) RemovePUT() *Route { return r.Remove(http.MethodPut) }

// RemovePOST is equal to r.Remove(http.MethodPost).
func (r *Route) RemovePOST() *Route { return r.Remove(http.MethodPost) }

// RemoveHEAD is equal to r.Remove(http.MethodHead).
func (r *Route) RemoveHEAD() *Route { return r.Remove(http.MethodHead) }

// RemovePATCH is equal to r.Remove(http.MethodPatch).
func (r *Route) RemovePATCH() *Route { return r.Remove(http.MethodPatch) }

// RemoveDELETE is equal to r.Remove(http.MethodDelete).
func (r *Route) RemoveDELETE() *Route { return r.Remove(http.MethodDelete) }

// RemoveCONNECT is equal to r.Remove(http.MethodConnect).
func (r *Route) RemoveCONNECT() *Route { return r.Remove(http.MethodConnect) }

// RemoveOPTIONS is equal to r.Remove(http.MethodOptions).
func (r *Route) RemoveOPTIONS() *Route { return r.Remove(http.MethodOptions) }

// RemoveTRACE is equal to r.Remove(http.MethodTrace).
func (r *Route) RemoveTRACE() *Route { return r.Remove(http.MethodTrace) }
