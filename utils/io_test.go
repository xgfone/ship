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
	"testing"
)

func TestReadNWriter(t *testing.T) {
	writer := bytes.NewBuffer(nil)
	reader := bytes.NewBufferString("test")
	if ReadNWriter(writer, reader, 4) != nil || writer.String() != "test" {
		t.Errorf("writer: %s", writer.String())
	}

	writer = bytes.NewBuffer(nil)
	reader = bytes.NewBufferString("test")
	if ReadNWriter(writer, reader, 2) != nil || writer.String() != "te" {
		t.Errorf("writer: %s", writer.String())
	} else if ReadNWriter(writer, reader, 2) != nil || writer.String() != "test" {
		t.Errorf("writer: %s", writer.String())
	}

	writer = bytes.NewBuffer(nil)
	reader = bytes.NewBufferString("test")
	if err := ReadNWriter(writer, reader, 5); err == nil {
		t.Error("non-nil")
	} else if err != io.EOF {
		t.Error(err)
	}
}
