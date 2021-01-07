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

package general

import "testing"

func TestRouter(t *testing.T) {
	var handler bool
	var n int
	var h interface{}
	var pnames, pvalues []string

	router := NewRouter(nil, nil)
	router.Add("static", "GET", "/static", handler)
	router.Add("param", "POST", "/test/:name", handler)

	if v := router.URL("param", "Aaron"); v != "/test/Aaron" {
		t.Errorf("expected '/test/Aaron', got '%s'", v)
	}

	if h, n = router.Find("GET", "/static", nil, nil); h == nil {
		t.Error("no route handler for 'GET /static'")
	}

	pnames = make([]string, 1)
	pvalues = make([]string, 1)
	if h, n = router.Find("POST", "/test/Aaron", pnames, pvalues); h == nil {
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
	router.Add("", "GET", "/static1/*", handler)
	if h, n = router.Find("GET", "/static1/path/to/file", pnames, pvalues); h == nil {
		t.Error("no route handler for 'GET /static1/path/to/file'")
	} else if n != 1 || pnames[0] != "*" || pvalues[0] != "path/to/file" {
		t.Errorf("expected dir 'path/to/file', but got '%s'", pvalues[0])
	}

	pnames = make([]string, 1)
	pvalues = make([]string, 1)
	router.Add("", "GET", "/static2/*filepath", handler)
	if h, n = router.Find("GET", "/static2/path/to/file", pnames, pvalues); h == nil {
		t.Error("no route handler for 'GET /static2/path/to/file'")
	} else if n != 1 || pnames[0] != "filepath" {
		t.Errorf("ParamName: expect '%s', got '%s'", "filepath", pnames[0])
	} else if pvalues[0] != "path/to/file" {
		t.Errorf("ParamValue: expected dir 'path/to/file', but got '%s'", pvalues[0])
	}

	if h, _ := router.Find("POST", "/test/param", nil, nil); h == nil {
		t.Error("not found the handler")
	}
}
