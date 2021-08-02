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
	"net/http/httptest"
	"testing"
)

func TestHTTPHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	rec := httptest.NewRecorder()

	s := New()
	httphandler := NotFoundHandler().HTTPHandler(s)
	httphandler.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("expect status code '%d', but got '%d'", 404, rec.Code)
	}

	handler := FromHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	req, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	rec = httptest.NewRecorder()
	ctx := s.AcquireContext(req, rec)
	handler(ctx)
	if rec.Code != 404 {
		t.Errorf("expect status code '%d', but got '%d'", 404, rec.Code)
	}
}
