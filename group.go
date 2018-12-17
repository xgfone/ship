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
	"strings"
)

// Group is a router group, that's, it manages a set of routes.
type Group struct {
	ship    *Ship
	prefix  string
	mdwares []Middleware
}

func newGroup(s *Ship, prefix string, middlewares ...Middleware) *Group {
	ms := make([]Middleware, 0, len(middlewares))
	return &Group{
		ship:    s,
		prefix:  strings.TrimRight(prefix, "/"),
		mdwares: append(ms, middlewares...),
	}
}

func (g *Group) newGroup(inherit bool, prefix string, ms ...Middleware) *Group {
	if len(prefix) == 0 || prefix[0] != '/' {
		panic(fmt.Errorf("prefix '%s' must start with '/'", prefix))
	}

	if inherit {
		ms = append(ms, g.mdwares...)
	}
	return newGroup(g.ship, g.prefix+prefix, ms...)
}

// Use adds some middlwares for the group.
func (g *Group) Use(middlewares ...Middleware) {
	g.mdwares = append(g.mdwares, middlewares...)
}

// Group returns a new sub-group.
func (g *Group) Group(prefix string, middlewares ...Middleware) *Group {
	return g.newGroup(true, prefix, middlewares...)
}

// GroupNone is the same as Group, but not inherit the middlewares of the parent.
func (g *Group) GroupNone(prefix string, middlewares ...Middleware) *Group {
	return g.newGroup(false, prefix, middlewares...)
}

// Route returns a new route, then you can customize and register it.
//
// You must call Route.Method() or its short method.
func (g *Group) Route(path string, handler Handler) *Route {
	return newRoute(g.ship, g.prefix, path, handler, g.mdwares...)
}

// R is short for Group#Route(path, handler).
func (g *Group) R(path string, handler Handler) *Route {
	return g.Route(path, handler)
}

// Path is equal to g.Route(path, nil), so you must set the handler later.
func (g *Group) Path(path string) *Route {
	return g.Route(path, nil)
}
