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
	"sync/atomic"

	"github.com/xgfone/ship/v5"
)

// MaxRequests returns a Middleware to allow the maximum number of the requests
// to max at a time.
//
// If the number of the requests exceeds the maximum, it will return the error
// ship.ErrTooManyRequests.
func MaxRequests(max uint32) Middleware {
	var maxNum = int32(max)
	var current int32

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			num := atomic.AddInt32(&current, 1)
			defer atomic.AddInt32(&current, -1)

			if num > maxNum {
				return ship.ErrTooManyRequests
			}
			return next(ctx)
		}
	}
}
