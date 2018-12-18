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
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Route represents a route information.
type Route struct {
	ship    *Ship
	path    string
	name    string
	mdwares []Middleware
}

func newRoute(s *Ship, prefix, path string, m ...Middleware) *Route {
	if !s.config.KeepTrailingSlashPath {
		path = strings.TrimSuffix(path, "/")
	}

	if len(path) == 0 {
		path = "/"
	} else if path[0] != '/' {
		panic(fmt.Errorf("path '%s' must start with '/'", path))
	}

	ms := make([]Middleware, 0, len(m))
	return &Route{
		ship: s,
		path: prefix + path,

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
//     s.R("/path/to", handler).Headers("Content-Type", "application/json").POST()
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
//     s.R("/path/to", handler).Schemes("https", "wss").POST()
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

// Method sets the methods and registers the route.
//
// If methods is nil, it will register all the supported methods for the route.
//
// Notice: The method must be called at last.
func (r *Route) Method(handler Handler, methods ...string) *Route {
	if len(methods) == 0 {
		panic(errors.New("the route requires methods"))
	}
	r.ship.addRoute(r.name, r.path, methods, handler, r.mdwares...)
	return r
}

// Any registers all the supported methods , which is short for
// r.Method(handler, "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE" )
func (r *Route) Any(handler Handler) *Route {
	return r.Method(handler, http.MethodConnect, http.MethodGet,
		http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodTrace)
}

// CONNECT is the short for r.Method("CONNECT").
func (r *Route) CONNECT(handler Handler) *Route {
	return r.Method(handler, http.MethodConnect)
}

// OPTIONS is the short for r.Method("OPTIONS").
func (r *Route) OPTIONS(handler Handler) *Route {
	return r.Method(handler, http.MethodOptions)
}

// HEAD is the short for r.Method("HEAD").
func (r *Route) HEAD(handler Handler) *Route {
	return r.Method(handler, http.MethodHead)
}

// PATCH is the short for r.Method("PATCH").
func (r *Route) PATCH(handler Handler) *Route {
	return r.Method(handler, http.MethodPatch)
}

// TRACE is the short for r.Method("TRACE").
func (r *Route) TRACE(handler Handler) *Route {
	return r.Method(handler, http.MethodTrace)
}

// GET is the short for r.Method("GET").
func (r *Route) GET(handler Handler) *Route {
	return r.Method(handler, http.MethodGet)
}

// PUT is the short for r.Method("PUT").
func (r *Route) PUT(handler Handler) *Route {
	return r.Method(handler, http.MethodPut)
}

// POST is the short for r.Method("POST").
func (r *Route) POST(handler Handler) *Route {
	return r.Method(handler, http.MethodPost)
}

// DELETE is the short for r.Method("DELETE").
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
