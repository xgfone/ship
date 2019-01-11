package middleware

import (
	"net/http"
	"sync/atomic"

	"github.com/xgfone/ship"
)

// MaxRequests returns a Middleware to allow the maximum number of the requests
// to max at a time.
//
// If the number of the requests exceeds the maximum, it will call the handler,
// which return the status code 429. But you can appoint yourself handler.
func MaxRequests(max uint32, handler ...ship.Handler) Middleware {
	h := func(c ship.Context) error { return c.NoContent(http.StatusTooManyRequests) }
	if len(handler) > 0 && handler[0] != nil {
		h = handler[0]
	}

	var maxNum = int32(max)
	var current int32

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			atomic.AddInt32(&current, 1)
			defer atomic.AddInt32(&current, -1)

			if atomic.LoadInt32(&current) > maxNum {
				return h(ctx)
			}
			return next(ctx)
		}
	}
}
