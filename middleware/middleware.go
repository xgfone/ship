package middleware

import (
	"math/rand"
	"strings"
	"time"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/core"
)

const (
	uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase    = "abcdefghijklmnopqrstuvwxyz"
	alphabetic   = uppercase + lowercase
	numeric      = "0123456789"
	alphanumeric = alphabetic + numeric
)

var defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Middleware is the alias of core.Middleware.
//
// We add it in order to show the middlewares in together by the godoc.
type Middleware = core.Middleware

// TokenFunc stands for a function to get a token from the request context.
type TokenFunc func(ctx ship.Context) (token string, err error)

// TokenValidator stands for a validator to validate whether a token is valid.
type TokenValidator func(token string) (ok bool, err error)

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
