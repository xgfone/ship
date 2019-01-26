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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

type sessionT struct {
	stores map[string]interface{}
}

func (s sessionT) GetSession(id string) (interface{}, error) {
	return s.stores[id], nil
}

func (s sessionT) SetSession(id string, value interface{}) error {
	s.stores[id] = value
	return nil
}

func TestContextSession(t *testing.T) {
	session := sessionT{stores: make(map[string]interface{})}
	buf := bytes.NewBuffer(nil)

	s := New(Config{Session: session})
	s.R("/").GET(func(ctx Context) error {
		v, _ := ctx.GetSession("id")
		fmt.Fprintf(buf, "%v", v)
		return nil
	}).POST(func(ctx Context) error {
		return ctx.SetSession("id", "abc")
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, "abc", buf.String())
}
