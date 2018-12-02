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

package ship

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type writerWrapper struct{ buf *bytes.Buffer }

func (w writerWrapper) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func newWriterWrapper() writerWrapper {
	return writerWrapper{buf: bytes.NewBuffer(nil)}
}

func TestLogger(t *testing.T) {
	w := newWriterWrapper()
	logger := NewNoLevelLogger(w)
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	ss := strings.Split(strings.TrimSpace(w.buf.String()), "\n")
	if len(ss) != 4 {
		t.Fail()
	}

	if !strings.Contains(ss[0], ": [DBUG] debug") {
		fmt.Println("000", ss[0])
		t.Fail()
	}

	if !strings.Contains(ss[1], ": [INFO] info") {
		fmt.Println("111", ss[1])
		t.Fail()
	}

	if !strings.Contains(ss[2], ": [WARN] warn") {
		fmt.Println("222", ss[2])
		t.Fail()
	}

	if !strings.Contains(ss[3], ": [EROR] error") {
		fmt.Println("333", ss[3])
		t.Fail()
	}
}
