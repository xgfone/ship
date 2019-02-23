// Copyright 2019 xgfone
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

//SetCtxHandler sets the context handler to h.
func SetCtxHandler(h func(*ship.Context, ...interface{}) error) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			ctx.SetHandler(h)
			return next(ctx)
		}
	}
}
