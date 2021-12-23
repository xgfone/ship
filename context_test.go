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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkContext(b *testing.B) {
	c := NewContext(0, 0)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			GetContext(SetContext(context.Background(), c))
		}
	})
}

func TestContext(t *testing.T) {
	c := NewContext(0, 0)
	nc := GetContext(SetContext(context.Background(), c))
	if nc != c {
		t.Errorf("unexpect the context")
	}
}

func TestContextBindQuery(t *testing.T) {
	type V struct {
		A string `query:"a" default:"xyz"`
		B int    `query:"b"`
	}
	v := V{}

	router := New()
	router.Route("/path").GET(func(c *Context) error { return c.BindQuery(&v) })

	req := httptest.NewRequest(http.MethodGet, "/path?b=2", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if v.A != "xyz" {
		t.Errorf("expect '%s', got '%s'", "xyz", v.A)
	} else if v.B != 2 {
		t.Errorf("expect %d, got %d", 2, v.B)
	}
}

func TestContextAccept(t *testing.T) {
	expected := []string{
		"text/html",
		"image/webp",
		"application/",
		"",
	}
	var accepts []string
	router := New()
	router.Route("/path").GET(func(ctx *Context) error {
		accepts = ctx.Accept()
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set(HeaderAccept, "text/html, application/*;q=0.9, image/webp, */*;q=0.8")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}

	for i := range expected {
		if expected[i] != accepts[i] {
			t.Errorf("expect '%s', got '%s'", expected[i], accepts[i])
		}
	}
}
