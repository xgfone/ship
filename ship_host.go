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
	"strings"
)

// RegexpHostRouter is the manager to match the host by the regular expression.
type RegexpHostRouter interface {
	// Len returns the number of the regexp host routers.
	Len() int

	// Range is used to traverse all the regexp host routers.
	Range(func(regexpHost string, router Router))

	// Add adds and returns the regexp host router.
	//
	// If the regexp host has been added, do nothing and return the added router.
	Add(regexpHost string, router Router) (Router, error)

	// Del deletes and returns the regexp host router.
	//
	// If the regexp host router does not exist, return nil.
	Del(regexpHost string) Router

	// Router accurately returns the router by the regexp host.
	//
	// If the regexp host router does not exist, return nil.
	Router(regexpHost string) Router

	// Match matches the host and returns the corresponding router information.
	//
	// If there is no regexp host router to match it, return ("", nil).
	Match(host string) (matchedRegexpHost string, matchedRouter Router)
}

type hostManager struct {
	Sum    int
	ehosts map[string]Router // Host Matching: Exact
	fhosts map[string]Router // Host Matching: Prefix/Suffix
	rhosts RegexpHostRouter  // Host Matching: Regexp
}

func newHostManager(regexpHostRouter RegexpHostRouter) *hostManager {
	return &hostManager{
		ehosts: make(map[string]Router, 4),
		fhosts: make(map[string]Router, 4),
		rhosts: regexpHostRouter,
	}
}

func (hm *hostManager) Len() int {
	return len(hm.ehosts) + len(hm.fhosts) + hm.rhosts.Len()
}

func (hm *hostManager) Range(f func(string, Router)) {
	for host, router := range hm.ehosts {
		f(host, router)
	}
	for host, router := range hm.fhosts {
		f(host, router)
	}
	hm.rhosts.Range(f)
}

func (hm *hostManager) Router(host string) Router {
	if host == "" {
		return nil
	} else if router, ok := hm.ehosts[host]; ok {
		return router
	} else if router, ok := hm.fhosts[host]; ok {
		return router
	} else if router := hm.rhosts.Router(host); router != nil {
		return router
	}
	return nil
}

func (hm *hostManager) Match(host string) (string, Router) {
	host = splitHost(host)

	// Exact Matching
	if len(hm.ehosts) != 0 {
		if router, ok := hm.ehosts[host]; ok {
			return host, router
		}
	}

	// Prefix/Suffix Matching
	if len(hm.fhosts) != 0 {
		for h, r := range hm.fhosts {
			if h[0] == '*' { // Suffix
				if strings.HasSuffix(host, h[1:]) {
					return h, r
				}
			} else if strings.HasPrefix(host, h[:len(h)-1]) { // Prefix
				return h, r
			}
		}
	}

	// Regexp Matching
	return hm.rhosts.Match(host)
}

func (hm *hostManager) Add(h string, r Router) (n Router, err error) {
	if strings.HasPrefix(h, "*.") { // Suffix Matching
		if !isDomainName(h[2:]) {
			return nil, fmt.Errorf("invalid domain '%s'", h)
		}

		if router, ok := hm.fhosts[h]; ok {
			r = router
		} else {
			hm.fhosts[h] = r
		}
	} else if strings.HasSuffix(h, ".*") { // Prefix Matching
		if !isDomainName(h[:len(h)-2]) {
			return nil, fmt.Errorf("invalid domain '%s'", h)
		}

		if router, ok := hm.fhosts[h]; ok {
			r = router
		} else {
			hm.fhosts[h] = r
		}
	} else if isDomainName(h) { // Exact Matching
		if router, ok := hm.ehosts[h]; ok {
			r = router
		} else {
			hm.ehosts[h] = r
		}
	} else if r, err = hm.rhosts.Add(h, r); err != nil { // Regexp Matching
		return nil, err
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
		hm.rhosts.Del(host)
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
