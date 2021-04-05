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

// RouteInfo is used to represent the information of the registered route.
type RouteInfo struct {
	// The host is used to isolate the routes. And the different host
	// maybe has the same routes.
	//
	// If Host is empty, it represents the default host.
	Host string `json:"host,omitempty" xml:"host,omitempty"`

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

	// CtxData is any additional context data, which is passed to the handler
	// and it can be accessed by Context.RouteInfo.CtxData in the handler.
	CtxData interface{} `json:"ctxdata,omitempty" xml:"ctxdata,omitempty"`
}

func (ri RouteInfo) String() string {
	if ri.Host == "" {
		if ri.Name == "" {
			return fmt.Sprintf("RouteInfo(method=%s, path=%s)", ri.Method, ri.Path)
		}
		return fmt.Sprintf("RouteInfo(name=%s, method=%s, path=%s)",
			ri.Name, ri.Method, ri.Path)
	} else if ri.Name == "" {
		return fmt.Sprintf("RouteInfo(host=%s, method=%s, path=%s)",
			ri.Host, ri.Method, ri.Path)
	}
	return fmt.Sprintf("RouteInfo(host=%s, name=%s, method=%s, path=%s)",
		ri.Host, ri.Name, ri.Method, ri.Path)
}

func (ri RouteInfo) checkPath() error {
	if len(ri.Path) == 0 || ri.Path[0] != '/' {
		return fmt.Errorf("path '%s' must start with '/'", ri.Path)
	}

	if i := strings.Index(ri.Path, "//"); i != -1 {
		return fmt.Errorf("bad path '%s' contains duplicate // at index:%d", ri.Path, i)
	}

	return nil
}

func (s *Ship) getRoutes(host string, r Router, rs []RouteInfo) []RouteInfo {
	for _, route := range r.Routes(nil) {
		ch := route.Handler.(RouteInfo)
		rs = append(rs, RouteInfo{
			Host:    host,
			Name:    route.Name,
			Path:    route.Path,
			Method:  route.Method,
			Handler: ch.Handler,
			CtxData: ch.CtxData,
		})
	}
	return rs
}

// Routes returns the information of all the routes.
func (s *Ship) Routes() (routes []RouteInfo) {
	s.Lock.RLock()
	nodefault := true
	routes = make([]RouteInfo, 0, s.hostManager.Sum+1)
	s.hostManager.Range(func(host string, router Router) {
		routes = s.getRoutes(host, router, routes)
		if nodefault && host == s.defaultHost {
			nodefault = false
		}
	})
	if nodefault {
		routes = s.getRoutes(s.defaultHost, s.defaultRouter, routes)
	}
	s.Lock.RUnlock()
	return
}

// AddRoutes registers a set of the routes.
//
// It will panic with it if there is an error when adding the routes.
func (s *Ship) AddRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.AddRoute(ri); err != nil {
			panic(err)
		}
	}
}

// DelRoutes deletes a set of the registered routes.
//
// It will panic with it if there is an error when deleting the routes.
func (s *Ship) DelRoutes(ris ...RouteInfo) {
	for _, ri := range ris {
		if err := s.DelRoute(ri); err != nil {
			panic(err)
		}
	}
}

// AddRoute registers the route.
func (s *Ship) AddRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if err != nil {
		return RouteError{RouteInfo: ri, Err: err}
	} else if !ok {
		return
	} else if ri.Handler == nil {
		return RouteError{RouteInfo: ri, Err: errors.New("handler must not be nil")}
	}

	var router Router
	s.Lock.Lock()
	if ri.Host == "" {
		router = s.defaultRouter
	} else if router = s.hostManager.Router(ri.Host); router == nil {
		if ri.Host == s.defaultHost {
			router = s.defaultRouter
		} else {
			router, err = s.hostManager.Add(ri.Host, s.newRouter())
		}
	}
	s.Lock.Unlock()

	if err != nil {
		return RouteError{RouteInfo: ri, Err: err}
	} else if n, e := router.Add(ri.Name, ri.Path, ri.Method, ri); e != nil {
		err = RouteError{RouteInfo: ri, Err: e}
	} else if n > s.URLParamMaxNum {
		router.Del(ri.Path, ri.Method)
		err = RouteError{RouteInfo: ri, Err: errors.New("too many url params")}
	}

	return
}

func (s *Ship) checkRouteInfo(ri *RouteInfo) (ok bool, err error) {
	ri.Method = strings.ToUpper(ri.Method)
	if s.RouteModifier != nil {
		*ri = s.RouteModifier(*ri)
	}

	if s.RouteFilter != nil && s.RouteFilter(*ri) {
		return
	}

	if err = ri.checkPath(); err == nil {
		ok = true
	}

	return
}

// DelRoute deletes the registered route, which only uses "Host", "Path"
// and "Method", and others are ignored.
//
// If Method is empty, deletes all the routes associated with the path.
//
// If the route does not exist, do nothing and return nil.
func (s *Ship) DelRoute(ri RouteInfo) (err error) {
	ok, err := s.checkRouteInfo(&ri)
	if !ok || err != nil {
		return
	}

	var r Router
	s.Lock.Lock()
	if ri.Host == "" {
		r = s.defaultRouter
	} else if r = s.hostManager.Router(ri.Host); r == nil && ri.Host == s.defaultHost {
		r = s.defaultRouter
	}
	s.Lock.Unlock()

	if r != nil {
		if err = r.Del(ri.Path, ri.Method); err != nil {
			err = RouteError{RouteInfo: ri, Err: err}
		}
	}

	return
}
