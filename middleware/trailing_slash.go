package middleware

import (
	"strings"

	"github.com/xgfone/ship"
)

// RemoveTrailingSlash returns a new middleware to remove the trailing slash
// in the request path if it exists.
//
// Notice: it should be used as the pre-middleware by ship#Pre().
func RemoveTrailingSlash() ship.Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) (err error) {
			req := ctx.Request()
			req.URL.Path = strings.TrimRight(req.URL.Path, "/")
			return next(ctx)
		}
	}
}
