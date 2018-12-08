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
	"net/http"
)

// Route represents a route information.
type Route struct {
	ship    *Ship
	path    string
	name    string
	prefix  string
	handler Handler
	mdwares []Middleware
}

func newRoute(s *Ship, prefix, path string, handler Handler, m ...Middleware) *Route {
	ms := make([]Middleware, 0, len(m))
	return &Route{
		ship:   s,
		path:   path,
		prefix: prefix,

		handler: handler,
		mdwares: append(ms, m...),
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

// Method sets the methods and registers the route.
//
// If methods is nil, it will register all the supported methods for the route.
//
// Notice: The method must be called at last.
func (r *Route) Method(methods ...string) {
	if len(r.name) > 0 && len(methods) > 1 {
		panic(errors.New("the named route only appoint one method"))
	}
	r.ship.addRoute(r.name, "", r.path, methods, r.handler, r.mdwares...)
}

// Any registers all the supported methods , which is short for
// r.Method("GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE" )
func (r *Route) Any() {
	r.Method(http.MethodConnect, http.MethodGet, http.MethodHead,
		http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodOptions, http.MethodTrace)
}

// CONNECT is the short for r.Method("CONNECT").
func (r *Route) CONNECT() {
	r.Method(http.MethodConnect)
}

// OPTIONS is the short for r.Method("OPTIONS").
func (r *Route) OPTIONS() {
	r.Method(http.MethodOptions)
}

// HEAD is the short for r.Method("HEAD").
func (r *Route) HEAD() {
	r.Method(http.MethodHead)
}

// PATCH is the short for r.Method("PATCH").
func (r *Route) PATCH() {
	r.Method(http.MethodPatch)
}

// TRACE is the short for r.Method("TRACE").
func (r *Route) TRACE() {
	r.Method(http.MethodTrace)
}

// GET is the short for r.Method("GET").
func (r *Route) GET() {
	r.Method(http.MethodGet)
}

// PUT is the short for r.Method("PUT").
func (r *Route) PUT() {
	r.Method(http.MethodPut)
}

// POST is the short for r.Method("POST").
func (r *Route) POST() {
	r.Method(http.MethodPost)
}

// DELETE is the short for r.Method("DELETE").
func (r *Route) DELETE() {
	r.Method(http.MethodDelete)
}
