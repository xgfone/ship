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
	"net/http"
	"strings"
)

// IsInteger reports whether s is the integer or not.
func IsInteger(s string) bool {
	if s == "" {
		return false
	}

	for i, _len := 0, len(s); i < _len; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// InStrings reports whether s is in the string slice ss or not.
func InStrings(s string, ss []string) bool {
	for i, _len := 0, len(ss); i < _len; i++ {
		if s == ss[i] {
			return true
		}
	}
	return false
}

// DisalbeRedirect is used to disalbe the default redirect behavior
// of http.Client, that's, http.Client won't handle the redirect response
// and just return it to the caller.
func DisalbeRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// SplitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func SplitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
