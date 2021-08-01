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
	"fmt"
	"strings"
)

// RouteGroupBuilder is a route group to build a set of routes.
type RouteGroupBuilder struct {
	ship    *Ship
	prefix  string
	data    interface{}
	mdwares []Middleware
}

func newRouteGroup(s *Ship, pprefix, prefix string,
	mws ...Middleware) *RouteGroupBuilder {
	if prefix = strings.TrimSuffix(prefix, "/"); len(prefix) == 0 {
		prefix = "/"
	} else if prefix[0] != '/' {
		panic(fmt.Errorf("prefix '%s' must start with '/'", prefix))
	}

	return &RouteGroupBuilder{
		ship:    s,
		prefix:  strings.TrimSuffix(pprefix, "/") + prefix,
		mdwares: append([]Middleware{}, mws...),
	}
}

// Group returns a new route sub-group with the group prefix.
func (s *Ship) Group(prefix string) *RouteGroupBuilder {
	return newRouteGroup(s, s.Prefix, prefix, s.mws...)
}

// Ship returns the ship that the current group belongs to.
func (g *RouteGroupBuilder) Ship() *Ship { return g.ship }

// Clone clones itself and returns a new one.
func (g *RouteGroupBuilder) Clone() *RouteGroupBuilder {
	return &RouteGroupBuilder{
		ship:    g.ship,
		data:    g.data,
		prefix:  g.prefix,
		mdwares: append([]Middleware{}, g.mdwares...),
	}
}

// Use appends some middlwares into the group.
func (g *RouteGroupBuilder) Use(middlewares ...Middleware) *RouteGroupBuilder {
	g.mdwares = append(g.mdwares, middlewares...)
	return g
}

// Group returns a new route sub-group.
func (g *RouteGroupBuilder) Group(prefix string, middlewares ...Middleware) *RouteGroupBuilder {
	mws := make([]Middleware, 0, len(g.mdwares)+len(middlewares))
	mws = append(mws, g.mdwares...)
	mws = append(mws, middlewares...)
	return newRouteGroup(g.ship, g.prefix, prefix, mws...)
}

// Data sets the context data.
func (g *RouteGroupBuilder) Data(data interface{}) *RouteGroupBuilder {
	g.data = data
	return g
}

// Route returns a new route, which is used to build and register the route.
//
// You should call Method() or its short method to register it.
func (g *RouteGroupBuilder) Route(path string) *RouteBuilder {
	return newRouteBuilder(g.ship, g, g.prefix, path, g.data, g.mdwares...)
}

// ResetMiddlewares resets the middlewares of the group to ms.
func (g *RouteGroupBuilder) ResetMiddlewares(ms ...Middleware) *RouteGroupBuilder {
	g.mdwares = append([]Middleware{}, ms...)
	return g
}

// AddRoutes registers a set of the routes.
//
// It will panic with it if there is an error when adding the routes.
func (g *RouteGroupBuilder) AddRoutes(routes ...Route) *RouteGroupBuilder {
	for i, r := range routes {
		routes[i].Path = g.Route(r.Path).path
	}
	g.ship.AddRoutes(routes...)
	return g
}

// DelRoutes deletes a set of the registered routes.
//
// It will panic with it if there is an error when deleting the routes.
func (g *RouteGroupBuilder) DelRoutes(routes ...Route) *RouteGroupBuilder {
	for i, r := range routes {
		routes[i].Path = g.Route(r.Path).path
	}
	g.ship.DelRoutes(routes...)
	return g
}
