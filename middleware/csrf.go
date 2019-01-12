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
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/xgfone/ship"
)

// CSRFConfig is used to configure the CSRF middleware.
type CSRFConfig struct {
	CookieCtxKey   string
	CookieName     string
	CookiePath     string
	CookieDomain   string
	CookieMaxAge   int
	CookieSecure   bool
	CookieHTTPOnly bool

	GenerateToken       func() string
	GetTokenFromRequest TokenFunc
}

// CSRF returns a CSRF middleware.
//
// If the config is missing, it will use:
//
//   conf := CSRFConfig{
//       CookieName:   "_csrf",
//       CookieCtxKey: "csrf",
//       CookieMaxAge: 86400,
//
//       GenerateToken:       GenerateToken(32),
//       GetTokenFromRequest: GetTokenFromHeader(ship.HeaderXCSRFToken),
//   }
//
func CSRF(config ...CSRFConfig) Middleware {
	var conf CSRFConfig
	if len(config) > 0 {
		conf = config[0]
	}

	if conf.CookieCtxKey == "" {
		conf.CookieCtxKey = "csrf"
	}
	if conf.CookieName == "" {
		conf.CookieName = "_csrf"
	}
	if conf.CookieMaxAge == 0 {
		conf.CookieMaxAge = 86400
	}
	if conf.GenerateToken == nil {
		conf.GenerateToken = GenerateToken(32)
	}
	if conf.GetTokenFromRequest == nil {
		conf.GetTokenFromRequest = GetTokenFromHeader(ship.HeaderXCSRFToken)
	}

	maxAge := time.Duration(conf.CookieMaxAge) * time.Second

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			var token string
			if cookie, err := ctx.Cookie(conf.CookieName); err != nil {
				token = conf.GenerateToken() // Generate the new token
			} else {
				token = cookie.Value // Reuse the token
			}

			req := ctx.Request()
			switch req.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			default:
				// Validate token only for requests which are not defined as 'safe' by RFC7231
				clientToken, err := conf.GetTokenFromRequest(ctx)
				if err != nil {
					if _, ok := err.(ship.HTTPError); ok {
						return err
					}
					return ship.NewHTTPError(http.StatusBadRequest, err.Error())
				}
				if !validateCSRFToken(token, clientToken) {
					return ship.NewHTTPError(http.StatusForbidden, "invalid csrf token")
				}
			}

			// Set CSRF cookie
			cookie := new(http.Cookie)
			cookie.Name = conf.CookieName
			cookie.Value = token
			if conf.CookiePath != "" {
				cookie.Path = conf.CookiePath
			}
			if conf.CookieDomain != "" {
				cookie.Domain = conf.CookieDomain
			}
			cookie.Expires = time.Now().Add(maxAge)
			cookie.Secure = conf.CookieSecure
			cookie.HttpOnly = conf.CookieHTTPOnly
			ctx.SetCookie(cookie)

			// Store token in the context
			ctx.Set(conf.CookieCtxKey, token)

			// Protect clients from caching the response
			ctx.Response().Header().Set(ship.HeaderVary, ship.HeaderCookie)

			return next(ctx)
		}
	}
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
