// Copyright 2020 xgfone
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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/ship/v5/router/echo"
)

func TestNotFound(t *testing.T) {
	notFound := func(c *Context) error {
		return c.Text(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	router := New()
	router.NotFound = notFound
	router.Route("/home/").GET(OkHandler())
	router.Route("/home/").POST(OkHandler())
	router.Route("/users/:id").GET(OkHandler())
	router.Route("/users/:id/:id2/:id3").GET(OkHandler())

	req, _ := http.NewRequest(http.MethodOptions, "/home/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, rec.Code)
	}

	req, _ = http.NewRequest(http.MethodGet, "/users/14/more", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	router := New()
	router.Router = echo.NewRouter(&echo.Config{
		MethodNotAllowedHandler: func(allowedMethods []string) interface{} {
			return MethodNotAllowedHandler(allowedMethods)
		}},
	)
	router.Route("/path").GET(OkHandler()).POST(OkHandler())

	req, _ := http.NewRequest(http.MethodPut, "/path", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expect status code '%d', but got '%d'", 405, rec.Code)
	} else if methods := rec.Header().Get(HeaderAllow); methods != "GET, POST" {
		t.Errorf("expect Allow header '%s', but got '%v'", "GET, POST", methods)
	}
}

func TestRouteFilter(t *testing.T) {
	router := New()
	router.RouteFilter = func(ri Route) bool {
		if ri.Name == "" {
			return true
		} else if !strings.HasPrefix(ri.Path, "/group/") {
			return true
		}
		return false
	}

	handler := func(ctx *Context) error { return nil }
	router.Group("/group").Route("/name").Name("test").GET(handler)
	router.Route("/noname").GET(handler)

	switch routes := router.Routes(); len(routes) {
	case 0:
		t.Error("no routes")
	case 1:
		if name := routes[0].Name; name != "test" {
			t.Errorf("expect route name '%s', but got '%s'", "test", name)
		}
	default:
		t.Errorf("too many routes: %v", routes)
	}
}

func TestRouteModifier(t *testing.T) {
	router := New()
	router.RouteModifier = func(ri Route) Route {
		if !strings.HasPrefix(ri.Path, "/prefix/") {
			ri.Path = "/prefix" + ri.Path
		}
		return ri
	}

	handler := func(ctx *Context) error { return nil }
	router.Route("/path").GET(handler)

	switch routes := router.Routes(); len(routes) {
	case 0:
		t.Error("no routes")
	case 1:
		if path := routes[0].Path; path != "/prefix/path" {
			t.Errorf("expect path '%s', but got '%s'", "/prefix/path", path)
		}
	default:
		t.Errorf("too many routes: %v", routes)
	}
}

const middlewareoutput = `
pre m1 start
pre m2 start
use m1 start
use m2 start
group m1 start
group m2 start
route m1 start
route m2 start
route m2 end
route m1 end
group m2 end
group m1 end
use m2 end
use m1 end
pre m2 end
pre m1 end
`

func TestMiddleware(t *testing.T) {
	bs := bytes.NewBufferString("\n")
	router := New()
	router.Pre(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("pre m1 start\n")
			err := next(ctx)
			bs.WriteString("pre m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("pre m2 start\n")
			err := next(ctx)
			bs.WriteString("pre m2 end\n")
			return err
		}
	})

	router.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("use m1 start\n")
			err := next(ctx)
			bs.WriteString("use m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("use m2 start\n")
			err := next(ctx)
			bs.WriteString("use m2 end\n")
			return err
		}
	})

	group := router.Group("/v1")
	group.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("group m1 start\n")
			err := next(ctx)
			bs.WriteString("group m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("group m2 start\n")
			err := next(ctx)
			bs.WriteString("group m2 end\n")
			return err
		}
	})

	group.Route("/route").Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("route m1 start\n")
			err := next(ctx)
			bs.WriteString("route m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("route m2 start\n")
			err := next(ctx)
			bs.WriteString("route m2 end\n")
			return err
		}
	}).GET(OkHandler())

	req := httptest.NewRequest(http.MethodGet, "/v1/route", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if bs.String() != middlewareoutput {
		t.Error(bs.String())
		t.Fail()
	}
}
