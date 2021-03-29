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
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/xgfone/ship/v4"
)

// LoggerConfig is used to configure the logger middleware.
type LoggerConfig struct {
	// If true, log the request body.
	//
	// Default: false
	LogReqBody bool
}

// Logger returns a new logger middleware that will log the request.
func Logger(config *LoggerConfig) Middleware {
	var conf LoggerConfig
	if config != nil {
		conf = *config
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			var bodyFmt string
			var bodyCnt string
			if conf.LogReqBody {
				buf := ctx.AcquireBuffer()
				body := bufferBody{Closer: ctx.Body(), Buffer: buf}
				_, err = ship.CopyNBuffer(body.Buffer, ctx.Body(), ctx.ContentLength(), nil)
				if err != nil {
					return
				}

				bodyFmt = ", reqbody="
				bodyCnt = body.Buffer.String()
				ctx.Request().Body = body
			}

			start := time.Now()
			err = next(ctx)
			cost := time.Since(start)

			req := ctx.Request()
			code := ctx.StatusCode()
			errmsg := ""

			switch e := err.(type) {
			case nil:
			case ship.HTTPServerError:
				if !ctx.IsResponded() {
					code = e.Code
				}
				if e.Code >= 400 {
					errmsg = e.Error()
				}
			default:
				errmsg = e.Error()
				if !ctx.IsResponded() {
					code = http.StatusInternalServerError
				}
			}

			if errmsg == "" {
				ctx.Logger().Infof("addr=%s, code=%d, method=%s, path=%s%s%s, starttime=%d, cost=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), bodyFmt, bodyCnt, start.Unix(), cost)
			} else {
				ctx.Logger().Errorf("addr=%s, code=%d, method=%s, path=%s%s%s, starttime=%d, cost=%s, err=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), bodyFmt, bodyCnt, start.Unix(), cost, errmsg)
			}

			return
		}
	}
}

type bufferBody struct {
	io.Closer
	*bytes.Buffer
}
