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

// RouteGroup is a route group, that's, it manages a set of routes.
type RouteGroup struct {
	ship    *Ship
	host    string
	prefix  string
	mdwares []Middleware
	ctxdata interface{}
}

func newRouteGroup(s *Ship, pprefix, prefix, host string, mws ...Middleware) *RouteGroup {
	if prefix = strings.TrimSuffix(prefix, "/"); len(prefix) == 0 {
		prefix = "/"
	} else if prefix[0] != '/' {
		panic(fmt.Errorf("prefix '%s' must start with '/'", prefix))
	}

	return &RouteGroup{
		ship:    s,
		host:    host,
		prefix:  strings.TrimSuffix(pprefix, "/") + prefix,
		mdwares: append([]Middleware{}, mws...),
	}
}

// Ship returns the ship that the current group belongs to.
func (g *RouteGroup) Ship() *Ship { return g.ship }

// Clone clones itself and returns a new one.
func (g *RouteGroup) Clone() *RouteGroup { rg := *g; return &rg }

// Host sets the host of the route group to host.
func (g *RouteGroup) Host(host string) *RouteGroup { g.host = host; return g }

// Use adds some middlwares for the group and returns the origin group
// to write the chained router.
func (g *RouteGroup) Use(middlewares ...Middleware) *RouteGroup {
	g.mdwares = append(g.mdwares, middlewares...)
	return g
}

// Group returns a new sub-group.
func (g *RouteGroup) Group(prefix string, middlewares ...Middleware) *RouteGroup {
	return newRouteGroup(g.ship, g.prefix, prefix, g.host, append(g.mdwares, middlewares...)...)
}

// CtxData sets the context data on the route group.
func (g *RouteGroup) CtxData(data interface{}) *RouteGroup {
	g.ctxdata = data
	return g
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (g *RouteGroup) Route(path string) *Route {
	return newRoute(g.ship, g, g.prefix, g.host, path, g.ctxdata, g.mdwares...)
}

// R is short for Group#Route(path).
func (g *RouteGroup) R(path string) *Route { return g.Route(path) }

// NoMiddlewares clears all the middlewares and returns itself.
func (g *RouteGroup) NoMiddlewares() *RouteGroup { g.mdwares = nil; return g }

// AddRoutes adds the routes by RouteInfo.
func (g *RouteGroup) AddRoutes(ris ...RouteInfo) *RouteGroup {
	for i, ri := range ris {
		ris[i].Path = g.Route(ri.Path).path
	}
	g.ship.AddRoutes(ris...)
	return g
}

// DelRoutes deletes the routes by RouteInfo.
func (g *RouteGroup) DelRoutes(ris ...RouteInfo) *RouteGroup {
	for i, ri := range ris {
		ris[i].Path = g.Route(ri.Path).path
	}
	g.ship.DelRoutes(ris...)
	return g
}
