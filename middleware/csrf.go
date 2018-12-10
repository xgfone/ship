package middleware

import (
	"crypto/subtle"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/xgfone/ship"
)

const (
	uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercase    = "abcdefghijklmnopqrstuvwxyz"
	alphabetic   = uppercase + lowercase
	numeric      = "0123456789"
	alphanumeric = alphabetic + numeric
)

var defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))

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
	GetTokenFromRequest func(ship.Context) (string, error)
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
//       GetTokenFromRequest: GetCSRFTokenFromHeader(ship.HeaderXCSRFToken),
//   }
//
func CSRF(config ...CSRFConfig) ship.Middleware {
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
		conf.GetTokenFromRequest = GetCSRFTokenFromHeader(ship.HeaderXCSRFToken)
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
			buf[i] = charset[rand.Int63()%_len]
		}
		return string(buf)
	}
}

// GetCSRFTokenFromHeader is used to get the CSRF token from the request header.
func GetCSRFTokenFromHeader(header string) func(ship.Context) (string, error) {
	return func(ctx ship.Context) (string, error) {
		return ctx.Request().Header.Get(header), nil
	}
}

// GetCSRFTokenFromForm is used to get the CSRF token from the request body FORM.
func GetCSRFTokenFromForm(param string) func(ship.Context) (string, error) {
	return func(ctx ship.Context) (string, error) {
		token := ctx.FormValue(param)
		if token != "" {
			return token, nil
		}
		return "", errors.New("missing CSRF token in the form parameter")
	}
}

// GetCSRFTokenFromQuery is used to get the CSRF token from the request URL query.
func GetCSRFTokenFromQuery(param string) func(ship.Context) (string, error) {
	return func(ctx ship.Context) (string, error) {
		token := ctx.QueryParam(param)
		if token != "" {
			return token, nil
		}
		return "", errors.New("missing CSRF token in the url query")
	}
}
