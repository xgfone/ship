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

// Package router supplies a router interface ane some implementations.
package router

import "fmt"

// Route represents the information of the registered route.
type Route struct {
	Name    string
	Path    string
	Method  string
	Handler interface{}
}

func (r Route) String() string {
	if r.Name == "" {
		return fmt.Sprintf("Route(method=%s, path=%s)", r.Method, r.Path)
	}
	return fmt.Sprintf("Route(name=%s, method=%s, path=%s)", r.Name, r.Method, r.Path)
}

// Router stands for a router management.
type Router interface {
	// Routes returns the list of all the routes.
	Routes() []Route

	// URL generates a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add adds a route and returns the number of the parameters
	// if there are the parameters in the route.
	//
	// For keeping consistent, the parameter should start with ":" or "*".
	// ":" stands for the single parameter, and "*" stands for the wildcard.
	Add(name, method, path string, handler interface{}) (paramNum int, err error)

	// Del deletes the given route.
	//
	// If name is not empty, lookup the path by it instead.
	//
	// If method is empty, deletes all the routes associated with the path.
	// Or only delete the given method for the path.
	Del(name, method, path string) (err error)

	// Find searchs and returns the handler and the number of the url path
	// paramethers. For the paramethers, they are put into pnames and pvalues.
	//
	// Return (nil, 0) if not found the route handler.
	Find(method, path string, pnames, pvalues []string) (handler interface{}, pn int)
}
