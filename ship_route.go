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
	"strings"
)

var (
	errInvalidHandler   = errors.New("handler must not be nil")
	errTooManyURLParams = errors.New("too many url params")
)

// Route is used to represent the information of the registered route.
type Route struct {
	// Path and Method represent the unique route in a certain host.
	//
	// Path maybe contain the parameters, which is determined by the underlying
	// router. And if Method is empty, it stands for all the methods.
	Path   string `json:"path,omitempty" xml:"path,omitempty"`
	Method string `json:"method,omitempty" xml:"method,omitempty"`

	// Name is the name of the path, which may be empty to represent no name.
	Name string `json:"name,omitempty" xml:"name,omitempty"`

	// Handler is the handler of the route to handle the request.
	Handler Handler `json:"-" xml:"-"`

	// Data is any additional data associated with the route.
	Data interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

func (r Route) String() string {
	if r.Name == "" {
		return fmt.Sprintf("RouteInfo(method=%s, path=%s)", r.Method, r.Path)
	}

	return fmt.Sprintf("RouteInfo(name=%s, method=%s, path=%s)",
		r.Name, r.Method, r.Path)
}

func (r Route) checkPath() error {
	if len(r.Path) == 0 || r.Path[0] != '/' {
		return fmt.Errorf("path '%s' must start with '/'", r.Path)
	}

	if i := strings.Index(r.Path, "//"); i != -1 {
		return fmt.Errorf("bad path '%s' contains duplicate // at index:%d",
			r.Path, i)
	}

	return nil
}

// Routes returns the information of all the routes.
func (s *Ship) Routes() (routes []Route) {
	routes = make([]Route, 0, 16)
	s.Router.Range(func(name, path, method string, handler interface{}) {
		routes = append(routes, handler.(Route))
	})
	return
}

// AddRoutes registers a set of the routes.
//
// It will panic with it if there is an error when adding the routes.
func (s *Ship) AddRoutes(routes ...Route) {
	for _, r := range routes {
		if err := s.AddRoute(r); err != nil {
			panic(err)
		}
	}
}

// DelRoutes deletes a set of the registered routes.
//
// It will panic with it if there is an error when deleting the routes.
func (s *Ship) DelRoutes(routes ...Route) {
	for _, r := range routes {
		if err := s.DelRoute(r); err != nil {
			panic(err)
		}
	}
}

// AddRoute registers the route.
func (s *Ship) AddRoute(r Route) (err error) {
	ok, err := s.checkRouteInfo(&r)
	if err != nil || !ok {
		return
	} else if r.Handler == nil {
		return RouteError{Route: r, Err: errInvalidHandler}
	}

	if n, _err := s.Router.Add(r.Name, r.Path, r.Method, r); _err != nil {
		err = RouteError{Route: r, Err: _err}
	} else if n > s.URLParamMaxNum {
		s.Router.Del(r.Path, r.Method)
		err = RouteError{Route: r, Err: errTooManyURLParams}
	}

	return
}

// DelRoute deletes the registered route, which only uses "Path" and "Method",
// and others are ignored.
//
// If Method is empty, deletes all the routes associated with the path.
//
// If the route does not exist, do nothing and return nil.
func (s *Ship) DelRoute(r Route) (err error) {
	ok, err := s.checkRouteInfo(&r)
	if !ok || err != nil {
		return
	}

	if err = s.Router.Del(r.Path, r.Method); err != nil {
		err = RouteError{Route: r, Err: err}
	}

	return
}

func (s *Ship) checkRouteInfo(r *Route) (ok bool, err error) {
	r.Method = strings.ToUpper(r.Method)
	if s.RouteModifier != nil {
		*r = s.RouteModifier(*r)
	}

	if s.RouteFilter != nil && s.RouteFilter(*r) {
		return
	}

	if err = r.checkPath(); err == nil {
		ok = true
	} else {
		err = RouteError{Route: *r, Err: err}
	}

	return
}
