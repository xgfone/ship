// Copyright 2021 xgfone
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
	"net/http"
	"testing"
)

func TestContentType(t *testing.T) {
	header := make(http.Header)
	SetContentType(header, MIMEApplicationJSON)
	if ct := header.Get(HeaderContentType); ct != MIMEApplicationJSON {
		t.Errorf("expect Content-Type '%s', but got '%s'", MIMEApplicationJSON, ct)
	}

	SetContentType(header, "text/test")
	if ct := header.Get(HeaderContentType); ct != "text/test" {
		t.Errorf("expect Content-Type '%s', but got '%s'", "text/test", ct)
	}

	AddContentTypeMapping("text/test", []string{"text/test_ct"})
	SetContentType(header, "text/test")
	if ct := header.Get(HeaderContentType); ct != "text/test_ct" {
		t.Errorf("expect Content-Type '%s', but got '%s'", "text/test_ct", ct)
	}
}
