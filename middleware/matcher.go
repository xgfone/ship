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
	"errors"

	"github.com/xgfone/ship"
)

// Matchers returns a middleware to execute the matchers, which will execute
// those matchers in turn. If a certain matcher returns an error, it will
// return a HTTPError with 404 and the error by default. But you can appoint
// a error handler.
func Matchers(matchers []ship.Matcher, handleError ...func(*ship.Context, error) error) Middleware {
	if len(matchers) == 0 {
		panic(errors.New("the matchers must not be empty"))
	}

	var handleErr func(*ship.Context, error) error
	if len(handleError) > 0 {
		if handleError[0] == nil {
			panic(errors.New("the error handler is nil"))
		}
		handleErr = handleError[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			req := ctx.Request()
			for _, matcher := range matchers {
				if err = matcher(req); err != nil {
					if handleErr != nil {
						return handleErr(ctx, err)
					}
					return ship.ErrNotFound.NewError(err)
				}
			}
			return nil
		}
	}
}
