package middleware

import (
	"net/http"

	"github.com/xgfone/ship"
)

// ResetResponse wraps and reset the response.
func ResetResponse(reset func(http.ResponseWriter) http.ResponseWriter) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			if resp := reset(ctx.Response()); resp != nil {
				ctx.SetResponse(resp)
			}
			return next(ctx)
		}
	}
}
