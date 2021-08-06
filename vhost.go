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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// IsDomainName is used to check whether the domain name is valid or not.
// And you can reset it to a customized one.
//
// The default implementation has these limits as follow:
//   - The maximum length of the full qualified domain name is equal to 253.
//   - The maximum length of the sub-domain name is equal to 63.
//   - The valid characters only contain "a-zA-z0-9_-.".
//
var IsDomainName func(domainName string) bool = isDomainName

// HostManager is used to manage the domain hosts.
type HostManager interface {
	// Len returns the number of the hosts.
	Len() int

	// Range is used to traverse all the hosts.
	Range(func(host string, handler http.Handler))

	// AddHost adds the host and the handler, then returns the added handler.
	//
	// If the host has been added, return the added handler.
	AddHost(host string, handler http.Handler) (http.Handler, error)

	// DelHost deletes the host accurately and returns the handler.
	//
	// If the host does not exist, return nil.
	DelHost(host string) http.Handler

	// GetHost matches the host accurately and returns the handler.
	//
	// If the host does not exist, return nil.
	GetHost(host string) http.Handler

	// MatchHost matches the host by the implemented rules, such as the regular
	// expression, and returns the corresponding handler.
	//
	// If there is no host to match it, return ("", nil).
	MatchHost(host string) (matchedHost string, matchedHandler http.Handler)
}

type lockHostManager struct {
	lock  sync.RWMutex
	hosts HostManager
}

// NewLockHostManager returns a new HostManager based on lock
// to manage the hosts safely.
func NewLockHostManager(hm HostManager) HostManager {
	return &lockHostManager{hosts: hm}
}

func (h *lockHostManager) Len() int {
	h.lock.RLock()
	_len := h.hosts.Len()
	h.lock.RUnlock()
	return _len
}

func (h *lockHostManager) Range(f func(host string, handler http.Handler)) {
	h.lock.RLock()
	h.hosts.Range(f)
	h.lock.RUnlock()
}

func (h *lockHostManager) AddHost(host string, handler http.Handler) (
	http.Handler, error) {
	h.lock.Lock()
	handler, err := h.hosts.AddHost(host, handler)
	h.lock.Unlock()
	return handler, err
}

func (h *lockHostManager) DelHost(host string) http.Handler {
	h.lock.Lock()
	handler := h.hosts.DelHost(host)
	h.lock.Unlock()
	return handler
}

func (h *lockHostManager) GetHost(host string) http.Handler {
	h.lock.RLock()
	handler := h.hosts.GetHost(host)
	h.lock.RUnlock()
	return handler
}

func (h *lockHostManager) MatchHost(host string) (string, http.Handler) {
	h.lock.RLock()
	host, handler := h.hosts.MatchHost(host)
	h.lock.RUnlock()
	return host, handler
}

type hostManager struct {
	total   int
	exacts  map[string]http.Handler
	prefixs map[string]http.Handler
	suffixs map[string]http.Handler
	regexps HostManager
}

// NewHostManager returns a new HostManager implementation,
// which uses IsDomainName to check whether a host name is the valid domain
// and supports three kinds of hosts:
//
//   - Exact: a valid domain, such as "www.example.com".
//   - Prefix: a valid domain with the suffix ".*", such as "www.example.*".
//   - Suffix: a valid domain with the prefix "*.", such as "*.example.com".
//   - Regexp: a valid regular expression defined by regexpHostManager.
//
// Notice: if the host name is not any of the exact, prefix and suffix formats,
// it will be regarded as the regexp host name.
//
// If regexpHostManager is nil, it is NewRegexpHostManager() by default.
func NewHostManager(regexpHostManager HostManager) HostManager {
	if regexpHostManager == nil {
		regexpHostManager = NewRegexpHostManager()
	}

	return &hostManager{
		exacts:  make(map[string]http.Handler, 4),
		prefixs: make(map[string]http.Handler, 4),
		suffixs: make(map[string]http.Handler, 4),
		regexps: regexpHostManager,
	}
}

func (h *hostManager) updateLen() {
	h.total = h.regexps.Len() +
		len(h.prefixs) +
		len(h.suffixs) +
		len(h.exacts)
}

func (h *hostManager) Len() int { return h.total }

func (h *hostManager) Range(f func(host string, handler http.Handler)) {
	for host, router := range h.exacts {
		f(host, router)
	}

	for host, router := range h.prefixs {
		f(host, router)
	}

	for host, router := range h.suffixs {
		f(host, router)
	}

	h.regexps.Range(f)
}

func (h *hostManager) AddHost(host string, handler http.Handler) (http.Handler, error) {
	var err error
	if host == "" {
		return nil, errors.New("host must not be empty")
	} else if strings.HasPrefix(host, "*.") { // Suffix Matching
		if !IsDomainName(host[2:]) {
			return nil, fmt.Errorf("invalid domain '%s'", host)
		} else if addedHandler, ok := h.suffixs[host]; ok {
			handler = addedHandler
		} else {
			h.suffixs[host] = handler
		}
	} else if strings.HasSuffix(host, ".*") { // Prefix Matching
		if !IsDomainName(host[:len(host)-2]) {
			return nil, fmt.Errorf("invalid domain '%s'", host)
		} else if addedHandler, ok := h.prefixs[host]; ok {
			handler = addedHandler
		} else {
			h.prefixs[host] = handler
		}
	} else if IsDomainName(host) { // Exact Matching
		if addedHandler, ok := h.exacts[host]; ok {
			handler = addedHandler
		} else {
			h.exacts[host] = handler
		}
	} else if handler, err = h.regexps.AddHost(host, handler); err != nil {
		return nil, err
	}

	h.updateLen()
	return handler, nil
}

func (h *hostManager) DelHost(host string) (handler http.Handler) {
	var ok bool
	if host == "" {
		return nil
	} else if strings.HasPrefix(host, "*.") {
		if handler, ok = h.suffixs[host]; ok {
			delete(h.suffixs, host)
		}
	} else if strings.HasSuffix(host, ".*") {
		if handler, ok = h.prefixs[host]; ok {
			delete(h.prefixs, host)
		}
	} else if IsDomainName(host) {
		if handler, ok = h.exacts[host]; ok {
			delete(h.exacts, host)
		}
	} else {
		handler = h.regexps.DelHost(host)
		ok = handler != nil
	}

	if ok {
		h.updateLen()
	}

	return
}

func (h *hostManager) GetHost(host string) http.Handler {
	if host == "" {
		return nil
	} else if handler, ok := h.exacts[host]; ok {
		return handler
	} else if handler, ok := h.suffixs[host]; ok {
		return handler
	} else if handler, ok := h.prefixs[host]; ok {
		return handler
	}

	return h.regexps.GetHost(host)
}

func (h *hostManager) MatchHost(host string) (
	matchedHost string, matchedHandler http.Handler) {
	host = splitHost(host)

	// Exact Matching
	if len(h.exacts) != 0 {
		if router, ok := h.exacts[host]; ok {
			return host, router
		}
	}

	// Suffix Matching
	if len(h.suffixs) != 0 {
		for matchedHost, matchedHandler = range h.suffixs {
			if strings.HasSuffix(host, matchedHost[1:]) {
				return
			}
		}
	}

	// Prefix Matching
	if len(h.prefixs) != 0 {
		for matchedHost, matchedHandler = range h.prefixs {
			if strings.HasPrefix(host, matchedHost[:len(matchedHost)-1]) {
				return
			}
		}
	}

	// Regexp Matching
	return h.regexps.MatchHost(host)
}

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
