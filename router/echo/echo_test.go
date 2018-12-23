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

package echo

import (
	"fmt"
	"testing"

	"github.com/xgfone/ship/core"
)

func TestRouter(t *testing.T) {
	router := NewRouter(nil, nil)
	router.Add("static", "/static", "GET", func(ctx core.Context) error { return ctx.String(200, "STATIC") })
	router.Add("param", "/test/:name", "POST", func(ctx core.Context) error {
		return ctx.String(200, fmt.Sprintf("hello %s", ctx.Param("name")))
	})

	router.Each(func(name, method, path string) {
		switch name {
		case "static":
			if method != "GET" || path != "/static" {
				t.Fail()
			}
		case "param":
			if method != "POST" || path != "/test/:name" {
				t.Fail()
			}
		}
	})

	if router.URL("param", "Aaron") != "/test/Aaron" {
		t.Fail()
	}

	if router.Find("GET", "/static", nil, nil) == nil {
		t.Fail()
	}

	pnames := make([]string, 1)
	pvalues := make([]string, 1)
	if router.Find("POST", "/test/Aaron", pnames, pvalues) == nil {
		t.Fail()
	}
	if pnames[0] != "name" || pvalues[0] != "Aaron" {
		t.Fail()
	}

	pnames[0] = ""
	pvalues[0] = ""
	router.Add("", "/static/*path", "GET", func(ctx core.Context) error { return nil })
	if router.Find("GET", "/static/path/to/file", pnames, pvalues) == nil {
		t.Fail()
	}
	if len(pnames) != 1 || pnames[0] != "path" || pvalues[0] != "path/to/file" {
		t.Fail()
	}
}
