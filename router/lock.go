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

package router

import "sync"

// NewLockRouter returns a new lock Router based on the original router r.
// So you can access and modify the routes concurrently and safely.
//
// Notice: the wrapped router must not panic.
func NewLockRouter(r Router) Router { return &lockRouter{router: r} }

var _ Router = &lockRouter{}

type lockRouter struct {
	lock   sync.RWMutex
	router Router
}

func (r *lockRouter) Routes(filter func(string, string, string) bool) []Route {
	r.lock.RLock()
	routes := r.router.Routes(filter)
	r.lock.RUnlock()
	return routes
}

func (r *lockRouter) Path(name string, params ...interface{}) string {
	r.lock.RLock()
	url := r.router.Path(name, params...)
	r.lock.RUnlock()
	return url
}

func (r *lockRouter) Add(name, path, method string, handler interface{}) (int, error) {
	r.lock.Lock()
	num, err := r.router.Add(name, path, method, handler)
	r.lock.Unlock()
	return num, err
}

func (r *lockRouter) Del(path, method string) (err error) {
	r.lock.Lock()
	err = r.router.Del(path, method)
	r.lock.Unlock()
	return
}

func (r *lockRouter) Match(path, method string, ns, vs []string) (interface{}, int) {
	r.lock.RLock()
	h, n := r.router.Match(path, method, ns, vs)
	r.lock.RUnlock()
	return h, n
}
