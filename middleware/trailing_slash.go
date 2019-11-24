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
	"strings"

	"github.com/xgfone/ship/v2"
)

// RemoveTrailingSlash returns a new middleware to remove the trailing slash
// in the request path if it exists.
//
// Notice: it should be used as the pre-middleware by ship#Pre().
func RemoveTrailingSlash() Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			req := ctx.Request()
			path := req.URL.Path
			if path != "" && path != "/" && path[len(path)-1] == '/' {
				path = strings.TrimRight(path, "/")
				if path == "" {
					req.URL.Path = "/"
				} else {
					req.URL.Path = path
				}
			}
			return next(ctx)
		}
	}
}
