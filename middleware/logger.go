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
	"time"

	"github.com/xgfone/ship"
)

// Logger returns a new logger middleware that will log the request.
//
// By default getTime is time.Now().
func Logger(now ...func() time.Time) Middleware {
	_now := time.Now
	if len(now) > 0 && now[0] != nil {
		_now = now[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			start := _now()
			err = next(ctx)
			end := _now()

			logger := ctx.Logger()
			if logger != nil {
				req := ctx.Request()
				logger.Info("method=%s, url=%s, starttime=%d, cost=%s, err=%v",
					req.Method, req.URL.RequestURI(), start.Unix(),
					end.Sub(start).String(), err)
			}
			return
		}
	}
}
