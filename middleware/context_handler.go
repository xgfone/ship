package middleware

import (
	"github.com/xgfone/ship"
)

//SetCtxHandler sets the context handler to h.
func SetCtxHandler(h func(ship.Context, ...interface{}) error) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			ctx.SetHandler(h)
			return next(ctx)
		}
	}
}
