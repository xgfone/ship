package middleware

import (
	"strings"

	"github.com/xgfone/ship"
)

// RemoveTrailingSlash returns a new middleware to remove the trailing slash
// in the request path if it exists.
//
// Notice: it should be used as the pre-middleware by ship#Pre().
func RemoveTrailingSlash() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) (err error) {
			req := ctx.Request()
			path := req.URL.Path
			if path != "" && path != "/" && path[len(path)-1] == '/' {
				path = strings.TrimRight(path, "/")
				if path == "" {
					req.URL.Path = "/"
				} else {
					req.URL.Path = path
				}
			}
			return next(ctx)
		}
	}
}
