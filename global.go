// Copyright 2019 xgfone
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
	"context"
	"net"
	"net/http"
)

// DefaultShip is the default global ship.
var DefaultShip = New()

// MaxNumOfURLParams is equal to DefaultShip.MaxNumOfURLParams().
func MaxNumOfURLParams() int {
	return DefaultShip.MaxNumOfURLParams()
}

// Configure is equal to DefaultShip.Configure(options...).
func Configure(options ...Option) *Ship {
	return DefaultShip.Configure(options...)
}

// Clone is equal to DefaultShip.Clone(name...).
func Clone(name ...string) *Ship {
	return DefaultShip.Clone(name...)
}

// Link is equal to DefaultShip.Link(other).
func Link(other *Ship) *Ship {
	return DefaultShip.Link(other)
}

// VHost is equal to DefaultShip.VHost(host).
func VHost(host string) *Ship {
	return DefaultShip.VHost(host)
}

// RegisterOnShutdown is equal to DefaultShip.RegisterOnShutdown(functions...).
func RegisterOnShutdown(functions ...func()) *Ship {
	return DefaultShip.RegisterOnShutdown(functions...)
}

// SetConnStateHandler is equal to DefaultShip.SetConnStateHandler(h).
func SetConnStateHandler(h func(net.Conn, http.ConnState)) *Ship {
	return DefaultShip.SetConnStateHandler(h)
}

// SetRouteFilter is equal to DefaultShip.SetRouteFilter(filter).
func SetRouteFilter(filter func(name, path, method string) bool) *Ship {
	return DefaultShip.SetRouteFilter(filter)
}

// SetRouteModifier is equal to DefaultShip.SetRouteModifier(filter).
func SetRouteModifier(modifier func(name, path, method string) (string, string, string)) *Ship {
	return DefaultShip.SetRouteModifier(modifier)
}

// Pre is equal to DefaultShip.Pre(middlewares...).
func Pre(middlewares ...Middleware) *Ship {
	return DefaultShip.Pre(middlewares...)
}

// Use is equal to DefaultShip.Use(middlewares...).
func Use(middlewares ...Middleware) *Ship {
	return DefaultShip.Use(middlewares...)
}

// Start is equal to DefaultShip.Start(addr, tlsFiles...).
func Start(addr string, tlsFiles ...string) *Ship {
	return DefaultShip.Start(addr, tlsFiles...)
}

// StartServer is equal to DefaultShip.StartServer(server).
func StartServer(server *http.Server) *Ship {
	return DefaultShip.StartServer(server)
}

// Wait is equal to DefaultShip.Wait().
func Wait() {
	DefaultShip.Wait()
}

// Shutdown is equal to DefaultShip.Shutdown(ctx).
func Shutdown(ctx context.Context) error {
	return DefaultShip.Shutdown(ctx)
}

// G is equal to DefaultShip.Group(prefix, middlewares...).
func G(prefix string, middlewares ...Middleware) *Group {
	return DefaultShip.Group(prefix, middlewares...)
}

// GroupWithoutMiddleware is equal to DefaultShip.GroupWithoutMiddleware(prefix, middlewares...).
func GroupWithoutMiddleware(prefix string, middlewares ...Middleware) *Group {
	return DefaultShip.GroupWithoutMiddleware(prefix, middlewares...)
}

// R is equal to DefaultShip.Route(path).
func R(path string) *Route {
	return DefaultShip.Route(path)
}

// RouteWithoutMiddleware is equal to DefaultShip.RouteWithoutMiddleware(path).
func RouteWithoutMiddleware(path string) *Route {
	return DefaultShip.RouteWithoutMiddleware(path)
}

// URL is equal to DefaultShip.URL(name, params...).
func URL(name string, params ...interface{}) string {
	return DefaultShip.URL(name, params...)
}

// Traverse is equal to DefaultShip.Traverse(f).
func Traverse(f func(name string, method string, path string)) {
	DefaultShip.Traverse(f)
}
