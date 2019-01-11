package middleware

import (
	"github.com/xgfone/ship"
)

// Flat returns a flat middleware, which will execute the handlers
// in turn, and terminate the rest handlers and return the error if there is
// an error returned by a certain handler.
//
// befores will be executed before handling the request, and afters is after
// handling the request.
//
// Example
//
//     beforeLog := func(ctx ship.Context) error {
//         ctx.Logger().Info("before handling the request")
//         return nil
//     }
//     afterLog := func(ctx ship.Context) error {
//         ctx.Logger().Info("after handling the request")
//         return nil
//     }
//     handler := func(ctx ship.Context) error {
//         ctx.Logger().Info("handling the request")
//         return nil
//     })
//
//     router := ship.New()
//     router.Use(Flat([]ship.Handler{beforeLog}, []ship.Handler{afterLog}))
//     router.R("/").GET(handler)
//
// You can pass the error by the ctx.SetError(err). For example,
//
//     handler := func(ctx ship.Context) error {
//         // ...
//         ctx.SetError(err)
//         return nil
//     })
//
//     afterLog := func(ctx ship.Context) (err error) {
//         if err = ctx.Error(); err != nil {
//             ctx.Logger().Info("after handling the request: %s", err.Error())
//             ctx.SetError(nil)  // Avoid to handle the error repeatedly by other middlewares.
//         } else {
//             ctx.Logger().Info("after handling the request")
//         }
//
//         return
//     }
//
func Flat(befores, afters []ship.Handler) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) (err error) {
			for _, h := range befores {
				if err = h(ctx); err != nil {
					return
				} else if err = ctx.Error(); err != nil {
					ctx.SetError(nil)
					return err
				}
			}

			if err = next(ctx); err != nil {
				return
			} else if err = ctx.Error(); err != nil {
				ctx.SetError(nil)
				return err
			}

			for _, h := range afters {
				if err = h(ctx); err != nil {
					return
				} else if err = ctx.Error(); err != nil {
					ctx.SetError(nil)
					return err
				}
			}

			return nil
		}
	}
}
