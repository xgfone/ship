// Copyright 2018 xgfone
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
	"fmt"

	"github.com/xgfone/ship"
)

func handlePanic(ctx *ship.Context, err interface{}) {
	if logger := ctx.Logger(); logger != nil {
		logger.Error("%v", err)
	}
}

// Recover returns a middleware to wrap the panic.
//
// Change:
//    1. Ignore the argument handle. In order to keep the backward compatibility,
//       we don't remove it until the next major version.
//    2. This middleware only recovers the panic and returns it as an error.
func Recover(handle ...func(*ship.Context, interface{})) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			defer func() {
				switch e := recover().(type) {
				case nil:
				case error:
					err = e
				default:
					err = fmt.Errorf("%v", e)
				}
			}()
			return next(ctx)
		}
	}
}
