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
