// Copyright 2021 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WArANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ship

import (
	"regexp"

	"github.com/xgfone/ship/v3/router"
)

type regexpRouter struct {
	router router.Router
	regexp *regexp.Regexp
}

type reHostRouter struct {
	hosts map[string]regexpRouter
}

// NewRegexpHostRouter returns a new RegexpHostRouter.
func NewRegexpHostRouter() RegexpHostRouter {
	return &reHostRouter{hosts: make(map[string]regexpRouter, 8)}
}

func (rr *reHostRouter) Len() int {
	return len(rr.hosts)
}

func (rr *reHostRouter) Each(f func(string, router.Router)) {
	for rehost, router := range rr.hosts {
		f(rehost, router.router)
	}
}

func (rr *reHostRouter) Add(h string, r router.Router) (router.Router, error) {
	re, err := regexp.Compile(h)
	if err != nil {
		return nil, err
	} else if router, ok := rr.hosts[h]; ok {
		return router.router, nil
	} else {
		rr.hosts[h] = regexpRouter{router: r, regexp: re}
		return r, nil
	}
}

func (rr *reHostRouter) Del(regexpHost string) router.Router {
	if r, ok := rr.hosts[regexpHost]; ok {
		delete(rr.hosts, regexpHost)
		return r.router
	}
	return nil
}

func (rr *reHostRouter) Router(regexpHost string) router.Router {
	if r, ok := rr.hosts[regexpHost]; ok {
		return r.router
	}
	return nil
}

func (rr *reHostRouter) Match(host string) (string, router.Router) {
	if len(rr.hosts) != 0 {
		for rehost, r := range rr.hosts {
			if r.regexp.MatchString(host) {
				return rehost, r.router
			}
		}
	}
	return "", nil
}
