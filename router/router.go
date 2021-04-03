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

// Router is a router manager based on the path with the optional method.
type Router interface {
	// Routes uses the filter to filter and return the routes if it returns true.
	//
	// Return all the routes if filter is nil.
	Routes(filter func(name, path, method string) bool) []Route

	// Path generates a url path by the path name and parameters.
	//
	// Return "" if there is not the route path named name.
	Path(name string, params ...interface{}) string

	// Add adds the route and returns the number of the parameters
	// if there are the parameters in the route path.
	//
	// name is the name of the path, which is optional and must be unique
	// if not empty.
	//
	// If method is empty, handler is the handler of all the methods supported
	// by the implementation. Or, it is only that of the given method.
	//
	// For the parameter in the path, the format is determined by the implementation.
	Add(name, path, method string, handler interface{}) (paramNum int, err error)

	// Del deletes the given route.
	//
	// If method is empty, deletes all the routes associated with the path.
	// Or, only delete the given method for the path.
	Del(path, method string) (err error)

	// Match matches the route by path and method, puts the path parameters
	// into pnames and pvalues, then returns the handler and the number
	// of the path paramethers.
	//
	// If pnames or pvalues is empty, it will ignore the path paramethers
	// when finding the route handler.
	//
	// Return (nil, 0) if not found the route handler.
	Match(path, method string, pnames, pvalues []string) (handler interface{}, pn int)
}
