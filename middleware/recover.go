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

func handlePanic(ctx ship.Context, err interface{}) {
	if logger := ctx.Logger(); logger != nil {
		logger.Error("%v", err)
	}
}

// Recover returns a middleware to wrap the panic.
//
// If missing handle, it will use the default, which logs the panic.
func Recover(handle ...func(ship.Context, interface{})) ship.Middleware {
	handlePanic := handlePanic
	if len(handle) > 0 && handle[0] != nil {
		handlePanic = handle[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) (err error) {
			defer func() {
				if err := recover(); err != nil {
					handlePanic(ctx, err)
				}
			}()
			return next(ctx)
		}
	}
}
