// The MIT License (MIT)
//
// Copyright (c) 2018 xgfone
// Copyright (c) 2017 LabStack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ship

import (
	"bufio"
	"net"
	"net/http"
)

// Response implements http.ResponseWriter.
type Response struct {
	Writer    http.ResponseWriter
	Status    int
	Size      int64
	Committed bool
	Filter    func([]byte) []byte
}

// NewResponse returns a new instance of Response.
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{Writer: w}
}

// SetWriter sets the writer and return itself.
func (r *Response) SetWriter(w http.ResponseWriter) *Response {
	r.Writer = w
	return r
}

// SetFilter sets the filter to filter the output.
func (r *Response) SetFilter(filter func([]byte) []byte) *Response {
	r.Filter = filter
	return r
}

// Header implements http.ResponseWriter#Header().
func (r *Response) Header() http.Header {
	return r.Writer.Header()
}

// WriteHeader implements http.ResponseWriter#WriteHeader().
func (r *Response) WriteHeader(code int) {
	if r.Committed {
		return
	}
	r.Status = code
	r.Committed = true
	r.Writer.WriteHeader(code)
}

// Write implements http.ResponseWriter#Writer().
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Committed {
		r.WriteHeader(http.StatusOK)
	}

	if r.Filter != nil {
		b = r.Filter(b)
	}

	if len(b) == 0 {
		return
	}

	n, err = r.Writer.Write(b)
	r.Size += int64(n)
	return
}

// Flush implements the http.Flusher interface to allow an HTTP handler
// to flush buffered data to the client.
//
// See [http.Flusher](https://golang.org/pkg/net/http/#Flusher)
func (r *Response) Flush() {
	r.Writer.(http.Flusher).Flush()
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler
// to take over the connection.
//
// See [http.Hijacker](https://golang.org/pkg/net/http/#Hijacker)
func (r *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.Writer.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
//
// This mechanism can be used to cancel long operations on the server
// if the client has disconnected before the response is ready.
//
// See [http.CloseNotifier](https://golang.org/pkg/net/http/#CloseNotifier)
func (r *Response) CloseNotify() <-chan bool {
	return r.Writer.(http.CloseNotifier).CloseNotify()
}

// Reset resets the response to the initialized and returns itself.
func (r *Response) Reset(w http.ResponseWriter) *Response {
	r.Writer = w
	r.Size = 0
	r.Status = http.StatusOK
	r.Committed = false
	r.Filter = nil
	return r
}
