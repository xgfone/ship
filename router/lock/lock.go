// Copyright 2019 xgfone <xgfone@126.com>
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

package lock

import (
	"errors"
	"sync"

	"github.com/xgfone/ship/core"
)

// LockedRouter returns a Router with the sync.RWMutex.
func LockedRouter(router core.Router) core.Router {
	if router == nil {
		panic(errors.New("the router is nil"))
	}
	return &lockedRouter{router: router}
}

type lockedRouter struct {
	sync.RWMutex
	router core.Router
}

func (lr *lockedRouter) URL(name string, params ...interface{}) (url string) {
	lr.RLock()
	url = lr.router.URL(name, params...)
	lr.RUnlock()
	return
}

func (lr *lockedRouter) Add(name string, path string, method string,
	handler core.Handler) (paramNum int) {
	lr.Lock()
	paramNum = lr.router.Add(name, path, method, handler)
	lr.Unlock()
	return
}

func (lr *lockedRouter) Find(method string, path string,
	pnames []string, pvalues []string) (handler core.Handler) {
	lr.RLock()
	handler = lr.router.Find(method, path, pnames, pvalues)
	lr.RUnlock()
	return
}

func (lr *lockedRouter) Each(f func(string, string, string)) {
	lr.RLock()
	lr.router.Each(f)
	lr.RUnlock()
}
