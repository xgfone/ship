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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetContext(t *testing.T) {
	s := New(EnableCtxHTTPContext(true))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)

	if GetContext(ctx.Request()) != ctx {
		t.Fail()
	}
}

func TestContext_ContentType_Charset(t *testing.T) {
	app := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderContentType, "application/json; version=1.0; charset = utf-8 ")
	rec := httptest.NewRecorder()
	ctx := app.NewContext(req, rec)

	assert.Equal(t, "application/json", ctx.ContentType())
	assert.Equal(t, "utf-8", ctx.Charset())

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderContentType, "application/json; charset = utf-8 ")
	rec = httptest.NewRecorder()
	ctx = app.NewContext(req, rec)

	assert.Equal(t, "application/json", ctx.ContentType())
	assert.Equal(t, "utf-8", ctx.Charset())
}

func TestToContentTypes(t *testing.T) {
	ct := ToContentTypes(MIMEApplicationJSON)
	if len(ct) != 1 && ct[0] != MIMEApplicationJSONs[0] {
		t.Error(ct)
	}
}

func ExampleContext_SetHandler() {
	responder := func(ctx *Context, args ...interface{}) error {
		return ctx.String(http.StatusOK, fmt.Sprintf("%s, %s", args...))
	}

	sethandler := func(next Handler) Handler {
		return func(ctx *Context) error {
			ctx.SetHandler(responder)
			return next(ctx)
		}
	}

	router := New()
	router.Use(sethandler)
	router.Route("/path/to").GET(func(c *Context) error { return c.Handle("Hello", "World") })

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
	responder := func(ctx *Context, args ...interface{}) error {
		return ctx.String(http.StatusOK, fmt.Sprintf("%s, %s", args...))
	}

	sethandler := func(next Handler) Handler {
		return func(ctx *Context) error {
			ctx.SetHandler(responder)
			return next(ctx)
		}
	}

	router := New()
	router.Use(sethandler)
	router.Route("/path/to").GET(func(c *Context) error { return c.Handle("Hello", "World") })

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

func (s sessionT) DelSession(id string) error {
	delete(s.stores, id)
	return nil
}

func TestContextSession(t *testing.T) {
	session := sessionT{stores: make(map[string]interface{})}
	buf := bytes.NewBuffer(nil)

	s := New(SetSession(session))
	s.R("/").GET(func(ctx *Context) error {
		v, _ := ctx.GetSession("id")
		fmt.Fprintf(buf, "%v\n", v)
		return nil
	}).POST(func(ctx *Context) error {
		return ctx.SetSession("id", "abc")
	}).PUT(func(ctx *Context) error {
		ctx.SetSession("id", "xyz")
		v, _ := ctx.GetSession("id")
		fmt.Fprintf(buf, "%v\n", v)
		ctx.SetSession("id", nil)
		v, _ = ctx.GetSession("id")
		fmt.Fprintf(buf, "%v\n", v)
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodPut, "/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, "abc\nxyz\n<nil>\nxyz\n", buf.String())
}

func TestContextURLParam(t *testing.T) {
	var maxParam int
	var name string
	var age string
	var names = make([]string, 5)
	var values = make([]string, 5)
	var maps map[string]string

	s := New()
	s.R("/hello/:name/:age").GET(func(c *Context) error {
		copy(names, c.ParamNames())
		copy(values, c.ParamValues())
		name = c.Param("name")
		age = c.Param("age")
		maps = c.Params()
		return nil
	})
	maxParam = s.maxNum

	req := httptest.NewRequest(http.MethodGet, "/hello/aaron/123", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 2, maxParam)
	assert.Equal(t, "aaron", name)
	assert.Equal(t, "123", age)
	assert.Equal(t, []string{"name", "age"}, names[:maxParam])
	assert.Equal(t, []string{"aaron", "123"}, values[:maxParam])
	assert.Equal(t, 2, len(maps))
	assert.Equal(t, "aaron", maps["name"])
	assert.Equal(t, "123", maps["age"])
}

func TestContext_ParamToStruct(t *testing.T) {
	type S struct {
		Name    string `url:"name"`
		Age     int    `url:"-"`
		Address string
	}
	var v S

	s := New()
	s.R("/hello/:name/:Age").GET(func(c *Context) error {
		return c.ParamToStruct(&v)
	})
	req := httptest.NewRequest(http.MethodGet, "/hello/aaron/123", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Body.String())
	assert.Equal(t, "aaron", v.Name)
	assert.Equal(t, 0, v.Age)
	assert.Equal(t, "", v.Address)
}
