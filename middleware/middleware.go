// Copyright 2018 xgfone <xgfone@126.com>
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

package middleware

import (
	"errors"
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

// Predefine some errors.
var (
	ErrTokenFromHeader = errors.New("missing token in the url header")
	ErrTokenFromQuery  = errors.New("missing token in the url query")
	ErrTokenFromForm   = errors.New("missing token in the form parameter")
)

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

// IsNoTokenError reports whether the error is that there is no token.
func IsNoTokenError(err error) bool {
	if err == ErrTokenFromForm || err == ErrTokenFromHeader || err == ErrTokenFromQuery {
		return true
	}
	return false
}

// GetTokenFromHeaderWithType is the same as GetTokenFromHeader, but it also
// supports the type of the token.
func getTokenFromHeaderWithType(header, _type string) TokenFunc {
	if header == "" {
		panic(errors.New("the header is empty"))
	}
	if _type == "" {
		panic(errors.New("the type is emtpy"))
	}
	typelen := len(_type)

	return func(ctx ship.Context) (string, error) {
		token := ctx.Request().Header.Get(header)
		if token == "" {
			return "", ErrTokenFromHeader
		} else if len(token) > typelen+1 && token[:typelen] == _type {
			return token[typelen+1:], nil
		}
		return "", ErrTokenFromHeader
	}
}

// GetTokenFromHeader is used to get the token from the request header.
//
// You can appoint the type of the token, which is separated by a whitespace,
// such as the header "Authorization".
func GetTokenFromHeader(header string, _type ...string) TokenFunc {
	if len(_type) > 0 {
		return getTokenFromHeaderWithType(header, _type[0])
	}

	return func(ctx ship.Context) (string, error) {
		if token := ctx.Request().Header.Get(header); token != "" {
			return token, nil
		}
		return "", ErrTokenFromHeader
	}
}

// GetTokenFromQuery is used to get the token from the request URL query.
func GetTokenFromQuery(param string) TokenFunc {
	return func(ctx ship.Context) (string, error) {
		if token := ctx.QueryParam(param); token != "" {
			return token, nil
		}
		return "", ErrTokenFromQuery
	}
}

// GetTokenFromForm is used to get the token from the request FORM body.
func GetTokenFromForm(param string) TokenFunc {
	return func(ctx ship.Context) (string, error) {
		if token := ctx.FormValue(param); token != "" {
			return token, nil
		}
		return "", ErrTokenFromForm
	}
}
