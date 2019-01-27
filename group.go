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
	"strings"
)

// Group is a router group, that's, it manages a set of routes.
type Group struct {
	ship    *Ship
	router  Router
	prefix  string
	mdwares []Middleware
}

func newGroup(s *Ship, router Router, pprefix, prefix string, middlewares ...Middleware) *Group {
	prefix = strings.TrimSuffix(prefix, "/")
	if len(prefix) == 0 {
		panic(errors.New("the prefix must not be empty"))
	}
	if prefix[0] != '/' {
		panic(fmt.Errorf("prefix '%s' must start with '/'", prefix))
	}

	ms := make([]Middleware, 0, len(middlewares))
	return &Group{
		ship:    s,
		router:  router,
		prefix:  pprefix + prefix,
		mdwares: append(ms, middlewares...),
	}
}

// Use adds some middlwares for the group and returns the origin group to write
// the chained router.
func (g *Group) Use(middlewares ...Middleware) *Group {
	g.mdwares = append(g.mdwares, middlewares...)
	return g
}

// Group returns a new sub-group.
func (g *Group) Group(prefix string, middlewares ...Middleware) *Group {
	return newGroup(g.ship, g.router, g.prefix, prefix, append(g.mdwares, middlewares...)...)
}

// GroupWithoutMiddleware is the same as Group, but not inherit the middlewares of the parent.
func (g *Group) GroupWithoutMiddleware(prefix string, middlewares ...Middleware) *Group {
	return newGroup(g.ship, g.router, g.prefix, prefix, middlewares...)
}

// RouteWithoutMiddleware is the same as Group, but not inherit the middlewares of the group.
func (g *Group) RouteWithoutMiddleware(path string) *Route {
	return newRoute(g.ship, g.router, g.prefix, path)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (g *Group) Route(path string) *Route {
	return newRoute(g.ship, g.router, g.prefix, path, g.mdwares...)
}

// R is short for Group#Route(path).
func (g *Group) R(path string) *Route {
	return g.Route(path)
}
