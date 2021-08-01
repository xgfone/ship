// Copyright 2020 xgfone
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
	"net/http"
	"time"

	"github.com/xgfone/ship/v5"
)

const logfmt = "addr=%s, method=%s, path=%s, code=%d, starttime=%d, cost=%s, err=%v"

// Logger returns a new logger middleware that will log the request.
func Logger() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			start := time.Now()
			err = next(ctx)
			cost := time.Since(start)

			code := ctx.StatusCode()
			if err != nil && !ctx.IsResponded() {
				if hse, ok := err.(ship.HTTPServerError); ok {
					code = hse.Code
				} else {
					code = http.StatusInternalServerError
				}
			}

			var logf func(string, ...interface{})
			if code < 400 {
				logf = ctx.Logger().Infof
			} else if code < 500 {
				logf = ctx.Logger().Warnf
			} else {
				logf = ctx.Logger().Errorf
			}

			req := ctx.Request()
			logf(logfmt, req.RemoteAddr, req.Method, req.URL.RequestURI(),
				code, start.Unix(), cost, err)

			return
		}
	}
}
