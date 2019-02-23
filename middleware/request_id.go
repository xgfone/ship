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
	"github.com/xgfone/ship"
)

// RequestID returns a X-Request-ID middleware.
//
// If the request header does not contain X-Request-ID, it will set a new one.
//
// generateRequestID is GenerateToken(32).
func RequestID(generateRequestID ...func() string) Middleware {
	getRequestID := GenerateToken(32)
	if len(generateRequestID) > 0 {
		getRequestID = generateRequestID[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {

			req := ctx.Request()
			xid := req.Header.Get(ship.HeaderXRequestID)
			if xid == "" {
				xid = getRequestID()
				req.Header.Set(ship.HeaderXRequestID, xid)
			}
			ctx.Response().Header().Set(ship.HeaderXRequestID, xid)

			return next(ctx)
		}
	}
}
