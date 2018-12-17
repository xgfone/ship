package middleware

import (
	"github.com/xgfone/ship"
)

// RequestID returns a X-Request-ID middleware.
//
// If the request header does not contain X-Request-ID, it will set a new one.
//
// generateRequestID is GenerateToken(32).
func RequestID(generateRequestID ...func() string) Middleware {
	getRequestID := GenerateToken(32)
	if len(generateRequestID) > 0 {
		getRequestID = generateRequestID[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {

			req := ctx.Request()
			xid := req.Header.Get(ship.HeaderXRequestID)
			if xid == "" {
				xid = getRequestID()
				req.Header.Set(ship.HeaderXRealIP, xid)
			}
			ctx.Response().Header().Set(ship.HeaderXRequestID, xid)

			return next(ctx)
		}
	}
}
