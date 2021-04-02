// Copyright 2021 xgfone
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

	"github.com/xgfone/ship/v4"
)

// HandleError returns a middleware to wrap and respond the error to the client
// if the handler has no response.
func HandleError() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			if err = next(ctx); err != nil && !ctx.IsResponded() {
				switch e := err.(type) {
				case ship.HTTPServerError:
					if e.CT == "" {
						return ctx.BlobText(e.Code, ship.MIMETextPlain, e.Error())
					} else {
						return ctx.BlobText(e.Code, e.CT, e.Error())
					}
				default:
					return ctx.NoContent(http.StatusInternalServerError)
				}
			}
			return nil
		}
	}
}
