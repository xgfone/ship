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

package utils

import (
	"bytes"
	"io"
)

// ReadN reads the data from io.Reader until n bytes or no incoming data
// if n is equal to or less than 0.
func ReadN(r io.Reader, n int64) (v []byte, err error) {
	buf := bytes.NewBuffer(nil)
	err = ReadNWriter(buf, r, n)
	return buf.Bytes(), err
}

// ReadNWriter reads n bytes to the writer w from the reader r.
//
// It will return io.EOF if the length of the data from r is less than n.
// But the data has been read into w.
func ReadNWriter(w io.Writer, r io.Reader, n int64) (err error) {
	if n < 1 {
		_, err := io.Copy(w, r)
		return err
	}

	if buf, ok := w.(*bytes.Buffer); ok {
		if n < 32768 { // 32KB
			buf.Grow(int(n))
		} else {
			buf.Grow(32768)
		}
	}

	if m, err := io.Copy(w, io.LimitReader(r, n)); err != nil {
		return err
	} else if m < n {
		return io.EOF
	}
	return nil
}
