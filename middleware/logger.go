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

	// Log is used to log the request.
	//
	// If LogReqBody is false, hasReqBody is false and reqBody is empty.
	// Or, hasReqBody is true and reqBody is the body of the request.
	//
	// Default: use Context.Logger() to log the request.
	Log func(req *http.Request, hasReqBody bool, reqBody string,
		statusCode int, startTime time.Time, cost time.Duration, err error)
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
				defer ctx.ReleaseBuffer(buf)
				body := bufferBody{Closer: ctx.Body(), Buffer: buf}
				_, err = ship.CopyNBuffer(body.Buffer, ctx.Body(), ctx.ContentLength(), nil)
				if err != nil {
					return ship.ErrBadRequest.New(err)
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

			var lerr error
			switch e := err.(type) {
			case nil:
			case ship.HTTPServerError:
				if !ctx.IsResponded() {
					code = e.Code
				}
				if e.Code >= 400 {
					lerr = err
				}
			default:
				lerr = err
				if !ctx.IsResponded() {
					code = http.StatusInternalServerError
				}
			}

			if conf.Log != nil {
				conf.Log(req, conf.LogReqBody, bodyCnt, code, start, cost, lerr)
			} else if code < 400 {
				ctx.Logger().Infof("addr=%s, method=%s, path=%s%s%s, code=%d, starttime=%d, cost=%s, err=%v",
					req.RemoteAddr, req.Method, req.URL.RequestURI(), bodyFmt, bodyCnt, code, start.Unix(), cost, lerr)
			} else if code < 500 {
				ctx.Logger().Warnf("addr=%s, method=%s, path=%s%s%s, code=%d, starttime=%d, cost=%s, err=%v",
					req.RemoteAddr, req.Method, req.URL.RequestURI(), bodyFmt, bodyCnt, code, start.Unix(), cost, lerr)
			} else {
				ctx.Logger().Errorf("addr=%s, method=%s, path=%s%s%s, code=%d, starttime=%d, cost=%s, err=%v",
					req.RemoteAddr, req.Method, req.URL.RequestURI(), bodyFmt, bodyCnt, code, start.Unix(), cost, lerr)
			}

			return
		}
	}
}

type bufferBody struct {
	io.Closer
	*bytes.Buffer
}
