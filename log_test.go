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

package ship

import (
	"bytes"
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
	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	ss := strings.Split(strings.TrimSpace(w.buf.String()), "\n")
	if len(ss) != 5 {
		t.Fail()
	}

	if !strings.Contains(ss[0], ": [T] trace") {
		t.Errorf("000: %s", ss[0])
	}

	if !strings.Contains(ss[1], ": [D] debug") {
		t.Errorf("111: %s", ss[1])
	}

	if !strings.Contains(ss[2], ": [I] info") {
		t.Errorf("222: %s", ss[2])
	}

	if !strings.Contains(ss[3], ": [W] warn") {
		t.Errorf("333: %s", ss[3])
	}

	if !strings.Contains(ss[4], ": [E] error") {
		t.Errorf("444: %s", ss[4])
	}
}
