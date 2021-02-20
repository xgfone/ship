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
	"fmt"
	"regexp"
	"strings"

	"github.com/xgfone/ship/v3/router"
)

type regexpRouter struct {
	Router router.Router
	Regexp *regexp.Regexp
}

type hostManager struct {
	Sum    int
	ship   *Ship
	ehosts map[string]router.Router // Host Matching: Exact
	fhosts map[string]router.Router // Host Matching: Prefix/Suffix
	rhosts map[string]regexpRouter  // Host Matching: Regexp
}

func newHostManager(ship *Ship) *hostManager {
	return &hostManager{
		ship:   ship,
		ehosts: make(map[string]router.Router, 4),
		fhosts: make(map[string]router.Router, 4),
		rhosts: make(map[string]regexpRouter, 4),
	}
}

func (hm *hostManager) Len() int {
	return len(hm.ehosts) + len(hm.fhosts) + len(hm.rhosts)
}

func (hm *hostManager) Each(f func(string, router.Router)) {
	for host, router := range hm.ehosts {
		f(host, router)
	}
	for host, router := range hm.fhosts {
		f(host, router)
	}
	for host, router := range hm.rhosts {
		f(host, router.Router)
	}
}

func (hm *hostManager) Router(host string) router.Router {
	if router, ok := hm.ehosts[host]; ok {
		return router
	} else if router, ok := hm.fhosts[host]; ok {
		return router
	} else if router, ok := hm.rhosts[host]; ok {
		return router.Router
	}
	return nil
}

func (hm *hostManager) Match(host string) router.Router {
	host = splitHost(host)

	// Exact Matching
	if len(hm.ehosts) != 0 {
		if router, ok := hm.ehosts[host]; ok {
			return router
		}
	}

	// Prefix/Suffix Matching
	if len(hm.fhosts) != 0 {
		for h, r := range hm.fhosts {
			if h[0] == '*' { // Suffix
				if strings.HasSuffix(host, h[1:]) {
					return r
				}
			} else if strings.HasPrefix(host, h[:len(h)-1]) { // Prefix
				return r
			}
		}
	}

	// Regexp Matching
	if len(hm.rhosts) != 0 {
		for _, r := range hm.rhosts {
			if r.Regexp.MatchString(host) {
				return r.Router
			}
		}
	}

	return nil
}

func (hm *hostManager) Add(h string, r router.Router) (router.Router, error) {
	if strings.HasPrefix(h, "*.") { // Prefix Matching
		if !isDomainName(h[2:]) {
			return nil, fmt.Errorf("invalid domain '%s'", h)
		}

		if router, ok := hm.fhosts[h]; ok {
			r = router
		} else if r == nil {
			r = hm.ship.newRouter()
			hm.fhosts[h] = r
		} else {
			hm.fhosts[h] = r
		}
	} else if strings.HasSuffix(h, ".*") { // Suffix Matching
		if !isDomainName(h[:len(h)-2]) {
			return nil, fmt.Errorf("invalid domain '%s'", h)
		}

		if router, ok := hm.fhosts[h]; ok {
			r = router
		} else if r == nil {
			r = hm.ship.newRouter()
			hm.fhosts[h] = r
		} else {
			hm.fhosts[h] = r
		}
	} else if isDomainName(h) { // Exact Matching
		if router, ok := hm.ehosts[h]; ok {
			r = router
		} else if r == nil {
			r = hm.ship.newRouter()
			hm.ehosts[h] = r
		} else {
			hm.ehosts[h] = r
		}
	} else { // Regexp Matching
		re, err := regexp.Compile(h)
		if err != nil {
			return nil, err
		}

		if router, ok := hm.rhosts[h]; ok {
			r = router.Router
		} else if r == nil {
			r = hm.ship.newRouter()
			hm.rhosts[h] = regexpRouter{Router: r, Regexp: re}
		} else {
			hm.rhosts[h] = regexpRouter{Router: r, Regexp: re}
		}
	}

	hm.Sum = hm.Len()
	return r, nil
}

func (hm *hostManager) Del(host string) {
	if host[0] == '*' || host[len(host)-1] == '*' {
		delete(hm.fhosts, host)
	} else if isDomainName(host) {
		delete(hm.ehosts, host)
	} else {
		delete(hm.rhosts, host)
	}
	hm.Sum = hm.Len()
}

/// --------------------------------------------------------------------------

// isDomainName checks if a string is a presentation-format domain name
// (currently restricted to hostname-compatible "preferred name" LDH labels and
// SRV-like "underscore labels"; see golang.org/issue/12421).
func isDomainName(s string) bool {
	// See RFC 1035, RFC 3696.
	// Presentation format has dots before every label except the first, and the
	// terminal empty label is optional here because we assume fully-qualified
	// (absolute) input. We must therefore reserve space for the first and last
	// labels' length octets in wire format, where they are necessary and the
	// maximum total length is 255.
	// So our _effective_ maximum is 253, but 254 is not rejected if the last
	// character is a dot.
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			nonNumeric = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
			nonNumeric = true
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return nonNumeric
}

func splitHost(host string) string {
	if host[0] == '[' { // For IPv6
		if index := strings.IndexByte(host, ']'); index != -1 {
			return host[1:index]
		}
		return ""
	} else if index := strings.IndexByte(host, ':'); index != -1 {
		return host[:index]
	}
	return host
}
