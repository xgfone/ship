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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ship

import (
	"net/http"
	"regexp"
)

var _ HostManager = &reHostManager{}

type regexpHost struct {
	handler http.Handler
	regexp  *regexp.Regexp
}

type reHostManager struct {
	hosts map[string]regexpHost
}

// NewRegexpHostManager returns a new HostManager based on regular expression,
// which uses the stdlib "regexp" to implement the regular expression syntax.
//
// For the golang regexp syntax, see https://pkg.go.dev/regexp/syntax.
func NewRegexpHostManager() HostManager {
	return &reHostManager{hosts: make(map[string]regexpHost, 4)}
}

func (rhm *reHostManager) Len() int { return len(rhm.hosts) }

func (rhm *reHostManager) Range(f func(string, http.Handler)) {
	for host, handler := range rhm.hosts {
		f(host, handler.handler)
	}
}

func (rhm *reHostManager) AddHost(host string, handler http.Handler) (http.Handler, error) {
	re, err := regexp.Compile(host)
	if err != nil {
		return nil, err
	} else if _handler, ok := rhm.hosts[host]; ok {
		return _handler.handler, nil
	}

	rhm.hosts[host] = regexpHost{handler: handler, regexp: re}
	return handler, nil
}

func (rhm *reHostManager) DelHost(host string) http.Handler {
	if handler, ok := rhm.hosts[host]; ok {
		delete(rhm.hosts, host)
		return handler.handler
	}
	return nil
}

func (rhm *reHostManager) GetHost(host string) http.Handler {
	if handler, ok := rhm.hosts[host]; ok {
		return handler.handler
	}
	return nil
}

func (rhm *reHostManager) MatchHost(host string) (string, http.Handler) {
	if len(rhm.hosts) != 0 {
		for matchedHost, matchedHandler := range rhm.hosts {
			if matchedHandler.regexp.MatchString(host) {
				return matchedHost, matchedHandler.handler
			}
		}
	}
	return "", nil
}
