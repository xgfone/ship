// Copyright 2018 xgfone <xgfone@126.com>
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
	"io"
	"sync"

	"github.com/xgfone/ship"
)

// BodyLimit is used to limit the maximum body of the request.
func BodyLimit(maxBodySize int64) Middleware {
	var pool sync.Pool
	if maxBodySize > 0 {
		pool = newLimitedReaderPool(maxBodySize)
	}

	putLimitedReaderIntoPool := func(r *limitedReader) {
		r.reader = nil
		pool.Put(r)
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx ship.Context) error {
			if maxBodySize < 1 {
				return next(ctx)
			}

			req := ctx.Request()
			if req.ContentLength > maxBodySize {
				return ship.ErrStatusRequestEntityTooLarge
			}

			reader := pool.Get().(*limitedReader)
			reader.Reset(req.Body)
			defer putLimitedReaderIntoPool(reader)
			req.Body = reader

			return next(ctx)
		}
	}
}

type limitedReader struct {
	reader io.ReadCloser
	read   int64
	limit  int64
}

func newLimitedReaderPool(limit int64) sync.Pool {
	return sync.Pool{New: func() interface{} { return &limitedReader{limit: limit} }}
}

func (lr *limitedReader) Read(b []byte) (n int, err error) {
	n, err = lr.reader.Read(b)
	lr.read += int64(n)
	if lr.read > lr.limit {
		return n, ship.ErrStatusRequestEntityTooLarge
	}
	return
}

func (lr *limitedReader) Close() error {
	return lr.reader.Close()
}

func (lr *limitedReader) Reset(reader io.ReadCloser) {
	lr.reader = reader
	lr.read = 0
}
