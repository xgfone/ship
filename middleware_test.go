// Copyright 2022 xgfone
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

func TestNewMiddleware(t *testing.T) {
	r := New()
	r.Use(NewMiddleware(func(h http.Handler, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("test", "abc")
		h.ServeHTTP(w, r)
		return nil
	}))
	r.Route("/").GET(func(c *Context) error { return c.NoContent(201) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect the status code %d, but got %d", 201, rec.Code)
	}
	if test := rec.Header().Get("test"); test != "abc" {
		t.Errorf("expect 'test' header '%s', but got '%s'", "abc", test)
	}
}

func BenchmarkNewMiddleware(b *testing.B) {
	r := New()
	r.Use(NewMiddleware(func(h http.Handler, w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}))
	r.Route("/").GET(func(c *Context) error { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.ServeHTTP(rec, req)
		}
	})
}
