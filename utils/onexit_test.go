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
	"sync"
	"testing"
)

func TestCallOnExit(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	OnExit(func() { buf.WriteString("call1\n") }, func() { buf.WriteString("call2\n") })
	OnExit(CallOnExit)
	OnExit()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			CallOnExit()
			wg.Done()
		}()
	}
	wg.Wait()

	if buf.String() != "call2\ncall1\n" {
		t.Error(buf.String())
	}
}
