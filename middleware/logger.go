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
	"net/http"
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
			cost := _now().Sub(start).String()

			req := ctx.Request()
			code := ctx.StatusCode()
			errmsg := ""

			switch e := err.(type) {
			case nil:
			case ship.HTTPError:
				if !ctx.IsResponded() {
					code = e.Code
				}
				if e.Code >= 500 {
					errmsg = e.Error()
				}
			default:
				errmsg = e.Error()
				if !ctx.IsResponded() {
					code = http.StatusInternalServerError
				}
			}

			if errmsg == "" {
				ctx.Logger().Info("addr=%s, code=%d, method=%s, url=%s, starttime=%d, cost=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), start.Unix(), cost)
			} else {
				ctx.Logger().Error("addr=%s, code=%d, method=%s, url=%s, starttime=%d, cost=%s, err=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), start.Unix(), cost, errmsg)
			}

			return
		}
	}
}
