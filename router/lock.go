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
func NewLockRouter(r Router) Router { return &lockRouter{router: r} }

type lockRouter struct {
	lock   sync.RWMutex
	router Router
}

func (r *lockRouter) URL(name string, params ...interface{}) string {
	r.lock.RLock()
	url := r.router.URL(name, params...)
	r.lock.RUnlock()
	return url
}

func (r *lockRouter) Add(name, method, path string, handler interface{}) int {
	r.lock.Lock()
	num := r.router.Add(name, method, path, handler)
	r.lock.Unlock()
	return num
}

func (r *lockRouter) Find(m, p string, ns, vs []string, h interface{}) interface{} {
	r.lock.RLock()
	handler := r.router.Find(m, p, ns, vs, h)
	r.lock.RUnlock()
	return handler
}
