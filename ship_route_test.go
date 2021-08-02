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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func TestRoute(t *testing.T) {
	router := New()
	handler := OkHandler()
	routes := []Route{
		{Name: "name1", Path: "/path1", Method: http.MethodGet, Handler: handler},
		{Name: "name2", Path: "/path2", Method: http.MethodPut, Handler: handler},
		{Name: "name3", Path: "/path3", Method: http.MethodPost, Handler: handler},
		{Name: "name4", Path: "/path4", Method: http.MethodHead, Handler: handler},
		{Name: "name5", Path: "/path5", Method: http.MethodPatch, Handler: handler},
		{Name: "name6", Path: "/path6", Method: http.MethodTrace, Handler: handler},
		{Name: "name7", Path: "/path7", Method: http.MethodDelete, Handler: handler},
		{Name: "name8", Path: "/path8", Method: http.MethodOptions, Handler: handler},
		{Name: "name9", Path: "/path9", Method: http.MethodDelete, Handler: handler},
	}

	for i, r := range routes {
		router.Route(r.Path).Name(r.Name).Data(i).Method(r.Handler, r.Method)
	}

	if rs := router.Routes(); len(rs) != 9 {
		t.Errorf("expect the number of the routes is %d, not 9\n", len(rs))
	} else {
		sort.Slice(rs, func(i, j int) bool {
			return rs[i].Data.(int) < rs[j].Data.(int)
		})

		for i, r := range rs {
			if index := r.Data.(int); i != index {
				t.Errorf("expect data index '%d', but got '%d'", i, index)
			}
		}
	}
}

func TestRoute_RemoveAny(t *testing.T) {
	h := OkHandler()
	router := New()
	router.Route("/path1").GET(h).POST(h).DELETE(h)
	if routes := router.Routes(); len(routes) != 3 {
		t.Error(routes)
	}

	router.Route("/path1").RemoveAny()
	if routes := router.Routes(); len(routes) != 0 {
		t.Error(routes)
	}
}

func TestRouteMap(t *testing.T) {
	router := New()
	router.Route("/path/to").Map(map[string]Handler{
		"GET":  OkHandler(),
		"POST": OkHandler(),
	})

	req := httptest.NewRequest(http.MethodGet, "/path/to", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/path/to", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAllRouteMethods(t *testing.T) {
	p := New()
	p.Use(func(next Handler) Handler {
		return func(c *Context) error {
			return next(c)
		}
	})

	handler := func(c *Context) error { return c.Text(200, c.Method()) }
	tests := []struct {
		method  string
		path    string
		url     string
		handler Handler
		code    int
		body    string
	}{
		{
			method:  http.MethodGet,
			path:    "/get",
			url:     "/get",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodGet,
		},
		{
			method:  http.MethodPost,
			path:    "/post",
			url:     "/post",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodPost,
		},
		{
			method:  http.MethodHead,
			path:    "/head",
			url:     "/head",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodHead,
		},
		{
			method:  http.MethodPut,
			path:    "/put",
			url:     "/put",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodPut,
		},
		{
			method:  http.MethodDelete,
			path:    "/delete",
			url:     "/delete",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodDelete,
		},
		{
			method:  http.MethodConnect,
			path:    "/connect",
			url:     "/connect",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodConnect,
		},
		{
			method:  http.MethodOptions,
			path:    "/options",
			url:     "/options",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodOptions,
		},
		{
			method:  http.MethodPatch,
			path:    "/patch",
			url:     "/patch",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodPatch,
		},
		{
			method:  http.MethodTrace,
			path:    "/trace",
			url:     "/trace",
			handler: handler,
			code:    http.StatusOK,
			body:    http.MethodTrace,
		},
		{
			method:  "PROPFIND",
			path:    "/propfind",
			url:     "/propfind",
			handler: handler,
			code:    http.StatusOK,
			body:    "PROPFIND",
		},
		{
			method: http.MethodGet,
			path:   "/users/:id",
			url:    "/users/13",
			handler: func(c *Context) (err error) {
				return c.Text(200, c.Param("id"))
			},
			code: http.StatusOK,
			body: "13",
		},
		{
			method: http.MethodGet,
			path:   "/2params/:p1",
			url:    "/2params/10",
			handler: func(c *Context) (err error) {
				return c.Text(200, c.Param("p1")+"|"+c.Param("p2"))
			},
			code: http.StatusOK,
			body: "10|",
		},
		{
			method: http.MethodGet,
			path:   "/2params/:p1/params/:p2",
			url:    "/2params/13/params/12",
			handler: func(c *Context) (err error) {
				return c.Text(200, c.Param("p1")+"|"+c.Param("p2"))
			},
			code: http.StatusOK,
			body: "13|12",
		},
	}

	for _, tt := range tests {
		switch tt.method {
		case http.MethodGet:
			p.Route(tt.path).GET(tt.handler)
		case http.MethodPost:
			p.Route(tt.path).POST(tt.handler)
		case http.MethodHead:
			p.Route(tt.path).HEAD(tt.handler)
		case http.MethodPut:
			p.Route(tt.path).PUT(tt.handler)
		case http.MethodDelete:
			p.Route(tt.path).DELETE(tt.handler)
		case http.MethodConnect:
			p.Route(tt.path).CONNECT(tt.handler)
		case http.MethodOptions:
			p.Route(tt.path).OPTIONS(tt.handler)
		case http.MethodPatch:
			p.Route(tt.path).PATCH(tt.handler)
		case http.MethodTrace:
			p.Route(tt.path).TRACE(tt.handler)
		default:
			p.Route(tt.path).Method(tt.handler, tt.method)
		}
	}

	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, tt.url, nil)
		if err != nil {
			t.Error(err)
		}

		res := httptest.NewRecorder()
		p.ServeHTTP(res, req)

		if tt.code != res.Code {
			t.Errorf("expect status code '%d', got '%d'", tt.code, res.Code)
		}

		if len(tt.body) > 0 {
			if body, err := ioutil.ReadAll(res.Body); err != nil {
				t.Error(err)
			} else if s := string(body); s != tt.body {
				t.Errorf("expect body '%s', got '%s'", tt.body, s)
			}
		}
	}

	// test any

	p2 := New()
	p2.Route("/test").Any(handler)

	test2 := []struct{ method string }{
		{method: http.MethodConnect},
		{method: http.MethodDelete},
		{method: http.MethodGet},
		{method: http.MethodHead},
		{method: http.MethodOptions},
		{method: http.MethodPatch},
		{method: http.MethodPost},
		{method: http.MethodPut},
		{method: http.MethodTrace},
	}

	for _, tt := range test2 {
		req, err := http.NewRequest(tt.method, "/test", nil)
		if err != nil {
			t.Error(err)
		}

		res := httptest.NewRecorder()
		p2.ServeHTTP(res, req)

		if 200 != res.Code {
			t.Errorf("expect status code '%d', got '%d'", 200, res.Code)
		}

		if body, err := ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		} else if s := string(body); s != tt.method {
			t.Errorf("expect body '%s', got '%s'", tt.method, s)
		}
	}
}
