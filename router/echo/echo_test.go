// Copyright 2019 xgfone
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

import "testing"

func TestRouter(t *testing.T) {
	var handler bool
	router := NewRouter(nil)
	router.Add("static", "GET", "/static", handler)
	router.Add("param", "POST", "/test/:name", handler)

	if v := router.URL("param", "Aaron"); v != "/test/Aaron" {
		t.Errorf("expected '/test/Aaron', got '%s'", v)
	}

	if router.Find("GET", "/static", nil, nil, nil) == nil {
		t.Error("no route handler for 'GET /static'")
	}

	pnames := make([]string, 1)
	pvalues := make([]string, 1)
	if router.Find("POST", "/test/Aaron", pnames, pvalues, nil) == nil {
		t.Error("no route handler for 'POST /test/Aaron'")
	}
	if pnames[0] != "name" {
		t.Errorf("expected url param name 'name', but got '%s'", pnames[0])
	} else if pvalues[0] != "Aaron" {
		t.Errorf("expected url param value 'Aaron', but got '%s'", pvalues[0])
	}

	pnames[0] = ""
	pvalues[0] = ""
	router.Add("", "GET", "/static1/*", handler)
	if router.Find("GET", "/static1/path/to/file", pnames, pvalues, nil) == nil {
		t.Error("no route handler for 'GET /static1/path/to/file'")
	} else if len(pnames) != 1 || pnames[0] != "*" || pvalues[0] != "path/to/file" {
		t.Errorf("expected dir 'path/to/file', but got '%s'", pvalues[0])
	}

	pnames[0] = ""
	pvalues[0] = ""
	router.Add("", "GET", "/static2/*filepath", handler)
	if router.Find("GET", "/static2/path/to/file", pnames, pvalues, nil) == nil {
		t.Error("no route handler for 'GET /static2/path/to/file'")
	} else if len(pnames) != 1 || pnames[0] != "filepath" {
		t.Errorf("ParamName: expect '%s', got '%s'", "filepath", pnames[0])
	} else if pvalues[0] != "path/to/file" {
		t.Errorf("ParamValue: expected dir 'path/to/file', but got '%s'", pvalues[0])
	}
}
