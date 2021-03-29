// Copy from github.com/labstack/echo/middleware:util.go

package middleware

import (
	"math/rand"
	"strings"
	"time"
)

const (
	uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase    = "abcdefghijklmnopqrstuvwxyz"
	alphabetic   = uppercase + lowercase
	numeric      = "0123456789"
	alphanumeric = alphabetic + numeric
)

var defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateToken returns a token generator which will generate
// a n-length token string.
func GenerateToken(n int, charsets ...string) func() string {
	charset := strings.Join(charsets, "")
	if charset == "" {
		charset = alphanumeric
	}
	_len := int64(len(charset))

	return func() string {
		buf := make([]byte, n)
		for i := range buf {
			buf[i] = charset[defaultRand.Int63()%_len]
		}
		return string(buf)
	}
}

func matchScheme(domain, pattern string) bool {
	didx := strings.Index(domain, ":")
	pidx := strings.Index(pattern, ":")
	return didx != -1 && pidx != -1 && domain[:didx] == pattern[:pidx]
}

// matchSubdomain compares authority with wildcard
func matchSubdomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}
	didx := strings.Index(domain, "://")
	pidx := strings.Index(pattern, "://")
	if didx == -1 || pidx == -1 {
		return false
	}
	domAuth := domain[didx+3:]
	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}
	patAuth := pattern[pidx+3:]

	domComp := strings.Split(domAuth, ".")
	patComp := strings.Split(patAuth, ".")
	for i := len(domComp)/2 - 1; i >= 0; i-- {
		opp := len(domComp) - 1 - i
		domComp[i], domComp[opp] = domComp[opp], domComp[i]
	}
	for i := len(patComp)/2 - 1; i >= 0; i-- {
		opp := len(patComp) - 1 - i
		patComp[i], patComp[opp] = patComp[opp], patComp[i]
	}

	for i, v := range domComp {
		if len(patComp) <= i {
			return false
		}
		p := patComp[i]
		if p == "*" {
			return true
		}
		if p != v {
			return false
		}
	}
	return false
}
