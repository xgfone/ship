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
	"bytes"
	"io"
	"net/http"
	"sync/atomic"
)

// OnceRunner is used to run the task only once, which is different from
// sync.Once, the second calling does not wait until the first calling finishes.
type OnceRunner struct {
	done uint32
	task func()
}

// NewOnceRunner returns a new OnceRunner.
func NewOnceRunner(task func()) *OnceRunner { return &OnceRunner{task: task} }

// Run runs the task.
func (r *OnceRunner) Run() {
	if atomic.CompareAndSwapUint32(&r.done, 0, 1) {
		r.task()
	}
}

// ReadNWriter reads n bytes to the writer w from the reader r.
//
// It will return io.EOF if the length of the data from r is less than n.
// But the data has been read into w.
func ReadNWriter(w io.Writer, r io.Reader, n int64) (err error) {
	buf := make([]byte, 4096)
	if n < 1 {
		_, err := io.CopyBuffer(w, r, buf)
		return err
	}

	if buf, ok := w.(*bytes.Buffer); ok {
		if n < 32768 { // 32KB
			buf.Grow(int(n))
		} else {
			buf.Grow(32768)
		}
	}

	if m, err := io.CopyBuffer(w, io.LimitReader(r, n), buf); err != nil {
		return err
	} else if m < n {
		return io.EOF
	}
	return nil
}

// DisalbeRedirect is used to disalbe the default redirect behavior
// of http.Client, that's, http.Client won't handle the redirect response
// and just return it to the caller.
func DisalbeRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
