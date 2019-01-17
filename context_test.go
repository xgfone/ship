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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetContext(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)

	if GetContext(ctx.Request()) != ctx {
		t.Fail()
	}
}

func ExampleContext_SetHandler() {
	responder := func(ctx Context, args ...interface{}) error {
		return ctx.String(http.StatusOK, fmt.Sprintf("%s, %s", args...))
	}

	sethandler := func(next Handler) Handler {
		return func(ctx Context) error {
			ctx.SetHandler(responder)
			return next(ctx)
		}
	}

	router := New()
	router.Use(sethandler)
	router.Route("/path/to").GET(func(c Context) error { return c.Handle("Hello", "World") })

	// For test
	req := httptest.NewRequest(http.MethodGet, "/path/to", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	fmt.Println(resp.Code)
	fmt.Println(resp.Body.String())

	// Output:
	// 200
	// Hello, World
}

func ExampleContext_Handle() {
	responder := func(ctx Context, args ...interface{}) error {
		return ctx.String(http.StatusOK, fmt.Sprintf("%s, %s", args...))
	}

	sethandler := func(next Handler) Handler {
		return func(ctx Context) error {
			ctx.SetHandler(responder)
			return next(ctx)
		}
	}

	router := New()
	router.Use(sethandler)
	router.Route("/path/to").GET(func(c Context) error { return c.Handle("Hello", "World") })

	// For test
	req := httptest.NewRequest(http.MethodGet, "/path/to", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	fmt.Println(resp.Code)
	fmt.Println(resp.Body.String())

	// Output:
	// 200
	// Hello, World
}
