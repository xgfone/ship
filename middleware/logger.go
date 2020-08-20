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

	"github.com/xgfone/ship/v3"
)

// Logger returns a new logger middleware that will log the request.
func Logger() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			start := time.Now()
			err = next(ctx)
			cost := time.Now().Sub(start).String()

			req := ctx.Request()
			code := ctx.StatusCode()
			errmsg := ""

			switch e := err.(type) {
			case nil:
			case ship.HTTPError:
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
				ctx.Logger().Infof("addr=%s, code=%d, method=%s, path=%s, starttime=%d, cost=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), start.Unix(), cost)
			} else {
				ctx.Logger().Errorf("addr=%s, code=%d, method=%s, path=%s, starttime=%d, cost=%s, err=%s",
					req.RemoteAddr, code, req.Method, req.URL.RequestURI(), start.Unix(), cost, errmsg)
			}

			return
		}
	}
}

type bufferBody struct {
	io.Closer
	*bytes.Buffer
}

// ReqBodyLogger returns a middleware to log the request body.
func ReqBodyLogger() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			if r := ctx.Request(); r.ContentLength > 0 {
				body := bufferBody{Closer: r.Body, Buffer: new(bytes.Buffer)}
				_, err = ship.CopyNBuffer(body.Buffer, r.Body, r.ContentLength, nil)
				if err != nil {
					return
				}
				ctx.Request().Body = body
				ctx.Logger().Debugf("addr=%s, method=%s, path=%s, reqbody=%s",
					r.RemoteAddr, r.Method, r.URL.RequestURI(), body.Buffer.String())
			}

			return next(ctx)
		}
	}
}
