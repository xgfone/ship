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

package router

// Router stands for a router management.
type Router interface {
	// Generate a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add a route with name, path, method and handler, and return the number
	// of the parameters if there are the parameters in the route. Or return 0.
	//
	// If there is any error, it should panic.
	//
	// Notice: for keeping consistent, the parameter should start with ":"
	// or "*". ":" stands for a single parameter, and "*" stands for
	// a wildcard parameter.
	Add(name, method, path string, handler interface{}) (paramNum int)

	// Find a route handler by the method and path of the request.
	//
	// Return defaultHandler instead if the route does not exist.
	//
	// If the route has parameters, the name and value of the parameters
	// should be stored `pnames` and `pvalues` respectively, which has
	// the enough capacity to store the paramether names and values.
	Find(method, path string, pnames, pvalues []string,
		defaultHandler interface{}) (handler interface{})
}
