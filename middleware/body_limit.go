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
	"io"
	"sync"

	"github.com/xgfone/ship/v4"
)

// BodyLimit is used to limit the maximum body of the request.
func BodyLimit(maxBodySize int64) Middleware {
	if maxBodySize < 1 {
		panic("BodyLimit: maxBodySize must be greater than 0")
	}

	pool := newLimitedReaderPool(maxBodySize)
	putIntoPool := func(r *limitedReader) { r.ReadCloser = nil; pool.Put(r) }
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			req := ctx.Request()

			if ctx.ContentLength() > maxBodySize {
				return ship.ErrStatusRequestEntityTooLarge
			}

			reader := pool.Get().(*limitedReader)
			reader.Reset(req.Body)
			req.Body = reader
			defer putIntoPool(reader)
			return next(ctx)
		}
	}
}

type limitedReader struct {
	io.ReadCloser
	read  int64
	limit int64
}

func newLimitedReaderPool(limit int64) sync.Pool {
	return sync.Pool{New: func() interface{} { return &limitedReader{limit: limit} }}
}

func (lr *limitedReader) Read(b []byte) (n int, err error) {
	n, err = lr.ReadCloser.Read(b)
	lr.read += int64(n)
	if lr.read > lr.limit {
		return n, ship.ErrStatusRequestEntityTooLarge
	}
	return
}

func (lr *limitedReader) Reset(reader io.ReadCloser) {
	lr.ReadCloser = reader
	lr.read = 0
}
