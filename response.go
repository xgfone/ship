// Copyright 2019 xgfone
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

package ship

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync"
)

// ResponsePool is used to cache the Response.
var responsePool = sync.Pool{New: func() interface{} { return NewResponse(nil) }}

// GetResponseFromPool returns a Response from the pool.
func GetResponseFromPool(w http.ResponseWriter) *Response {
	res := responsePool.Get().(*Response)
	res.SetWriter(w)
	return res
}

// PutResponseIntoPool puts a Response into the pool.
func PutResponseIntoPool(r *Response) { r.Reset(nil); responsePool.Put(r) }

// Response implements http.ResponseWriter.
type Response struct {
	http.ResponseWriter

	Size   int64
	Wrote  bool
	Status int
}

// NewResponse returns a new instance of Response.
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{ResponseWriter: w, Status: http.StatusOK}
}

// WriteHeader implements http.ResponseWriter#WriteHeader().
func (r *Response) WriteHeader(code int) {
	if !r.Wrote {
		r.Wrote = true
		r.Status = code
		r.ResponseWriter.WriteHeader(code)
	}
}

// Write implements http.ResponseWriter#Writer().
func (r *Response) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	r.WriteHeader(http.StatusOK)
	n, err = r.ResponseWriter.Write(b)
	r.Size += int64(n)
	return
}

// WriteString implements io.StringWriter.
func (r *Response) WriteString(s string) (n int, err error) {
	if len(s) == 0 {
		return
	}

	r.WriteHeader(http.StatusOK)
	n, err = io.WriteString(r.ResponseWriter, s)
	r.Size += int64(n)
	return
}

// Reset resets the response to the initialized and returns itself.
func (r *Response) Reset(w http.ResponseWriter) {
	*r = Response{ResponseWriter: w, Status: http.StatusOK}
}

// SetWriter resets the writer to w and return itself.
func (r *Response) SetWriter(w http.ResponseWriter) { r.ResponseWriter = w }

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
//
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// Push implements the http.Pusher interface to support HTTP/2 server push.
//
// See [http.Pusher](https://golang.org/pkg/net/http/#Pusher)
func (r *Response) Push(target string, opts *http.PushOptions) error {
	return r.ResponseWriter.(http.Pusher).Push(target, opts)
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
//
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
