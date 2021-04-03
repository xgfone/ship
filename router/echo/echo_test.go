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

package echo

import (
	"net/http"
	"testing"
)

func TestRemoveTrailingSlash(t *testing.T) {
	if path := removeTrailingSlash(""); path != "" {
		t.Error(path)
	}
	if path := removeTrailingSlash("/"); path != "" {
		t.Error(path)
	}
	if path := removeTrailingSlash("a"); path != "a" {
		t.Error(path)
	}
	if path := removeTrailingSlash("abc//"); path != "abc" {
		t.Error(path)
	}
	if path := removeTrailingSlash("/a/b/"); path != "/a/b" {
		t.Error(path)
	}
	if path := removeTrailingSlash("/a/b"); path != "/a/b" {
		t.Error(path)
	}

	r := NewRouter(&Config{RemoveTrailingSlash: true})
	if _, err := r.Add("", "/v1/path/", http.MethodGet, true); err != nil {
		t.Error(err)
	} else if rs := r.Routes(nil); len(rs) != 1 {
		t.Error(rs)
	}

	if h, _ := r.Match("/v1/path/", http.MethodGet, nil, nil); h == nil {
		t.Error("no route")
	}
	if h, _ := r.Match("/v1/path", http.MethodGet, nil, nil); h == nil {
		t.Error("no route")
	}

	if err := r.Del("/v1/path/", http.MethodGet); err != nil {
		t.Error(err)
	} else if rs := r.Routes(nil); len(rs) != 0 {
		t.Error(rs)
	}

	r.Add("", "/v1/path/", http.MethodGet, true)
	if err := r.Del("/v1/path", http.MethodGet); err != nil {
		t.Error(err)
	} else if rs := r.Routes(nil); len(rs) != 0 {
		t.Error(rs)
	}
}

func TestRouter(t *testing.T) {
	var handler bool
	var n int
	var h interface{}
	var pnames, pvalues []string

	router := NewRouter(nil)
	router.Add("static", "/static", "GET", handler)
	router.Add("param", "/test/:name", "POST", handler)

	if v := router.Path("param", "Aaron"); v != "/test/Aaron" {
		t.Errorf("expected '/test/Aaron', got '%s'", v)
	}

	if h, _ = router.Match("/static", "GET", nil, nil); h == nil {
		t.Error("no route handler for 'GET /static'")
	}

	pnames = make([]string, 1)
	pvalues = make([]string, 1)
	if h, n = router.Match("/test/Aaron", "POST", pnames, pvalues); h == nil {
		t.Error("no route handler for 'POST /test/Aaron'")
	} else if n == 0 {
		t.Errorf("no paramether number")
	} else if pnames[0] != "name" {
		t.Errorf("expected url param name 'name', but got '%s'", pnames[0])
	} else if pvalues[0] != "Aaron" {
		t.Errorf("expected url param value 'Aaron', but got '%s'", pvalues[0])
	}

	pnames = make([]string, 1)
	pvalues = make([]string, 1)
	router.Add("", "/static1/*", "GET", handler)
	if h, n = router.Match("/static1/path/to/file", "GET", pnames, pvalues); h == nil {
		t.Error("no route handler for 'GET /static1/path/to/file'")
	} else if n != 1 || pnames[0] != "*" || pvalues[0] != "path/to/file" {
		t.Errorf("expected dir 'path/to/file', but got '%s'", pvalues[0])
	}

	pnames = make([]string, 1)
	pvalues = make([]string, 1)
	router.Add("", "/static2/*filepath", "GET", handler)
	if h, n = router.Match("/static2/path/to/file", "GET", pnames, pvalues); h == nil {
		t.Error("no route handler for 'GET /static2/path/to/file'")
	} else if n != 1 || pnames[0] != "filepath" {
		t.Errorf("ParamName: expect '%s', got '%s'", "filepath", pnames[0])
	} else if pvalues[0] != "path/to/file" {
		t.Errorf("ParamValue: expected dir 'path/to/file', but got '%s'", pvalues[0])
	}

	if h, _ := router.Match("/test/param", "POST", nil, nil); h == nil {
		t.Error("not found the handler")
	}

	routes := router.Routes(func(n, _, _ string) bool { return n == "param" })
	if len(routes) != 1 {
		t.Error(routes)
	} else if r := routes[0]; r.Method != "POST" || r.Path != "/test/:name" {
		t.Errorf("expect method='%s', path='%s', but got method='%s', path='%s'",
			"POST", "/test/:name", r.Method, r.Path)
	}

}

func TestRouterAnyMethod(t *testing.T) {
	handler1 := 1
	handler2 := 2
	handler3 := 3
	handler4 := 4

	router := NewRouter(nil)
	router.Add("", "/path1", "GET", handler1)
	router.Add("", "/path2", "PUT", handler2)
	router.Add("", "/path2", "POST", handler3)
	router.Add("", "/path2", "", handler4)

	if rs := router.Routes(nil); len(rs) != 12 {
		t.Error(rs)
	}

	router.Del("/path2", "POST")
	if rs := router.Routes(nil); len(rs) != 11 {
		t.Error(rs)
	} else {
		for _, r := range rs {
			switch r.Method {
			case "POST":
				t.Error(r)
			}
		}
	}

	router.Del("/path2", "")
	if rs := router.Routes(nil); len(rs) != 1 || rs[0].Path != "/path1" {
		t.Error(rs)
	}
}
