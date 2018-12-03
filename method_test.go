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

package ship_test

import (
	"testing"

	"github.com/xgfone/ship"
)

type TestStruct struct{}

func (t TestStruct) Create(ctx ship.Context) error { return nil }
func (t TestStruct) Delete(ctx ship.Context) error { return nil }
func (t TestStruct) Update(ctx ship.Context) error { return nil }
func (t TestStruct) Get(ctx ship.Context) error    { return nil }
func (t TestStruct) Has(ctx ship.Context) error    { return nil }
func (t TestStruct) NotHandler()                   {}

func strIsInSlice(s string, ss []string) bool {
	for _, _s := range ss {
		if _s == s {
			return true
		}
	}
	return false
}

func TestMapMethodIntoRouter(t *testing.T) {
	router := ship.NewRouter()
	ts := TestStruct{}
	paths := ship.MapMethodIntoRouter(router, ts, "/v1")
	if len(paths) != 4 {
		t.Fail()
	} else {
		if !strIsInSlice("/v1/teststruct/get", paths) {
			t.Fail()
		}
		if !strIsInSlice("/v1/teststruct/create", paths) {
			t.Fail()
		}
		if !strIsInSlice("/v1/teststruct/update", paths) {
			t.Fail()
		}
		if !strIsInSlice("/v1/teststruct/delete", paths) {
			t.Fail()
		}
	}

	router.Each(func(name, method, path string, handler ship.Handler) {
		switch method {
		case "GET":
			if name != "teststruct_get" || path != "/v1/teststruct/get" {
				t.Fail()
			}
		case "POST":
			if name != "teststruct_create" || path != "/v1/teststruct/create" {
				t.Fail()
			}
		case "PUT":
			if name != "teststruct_update" || path != "/v1/teststruct/update" {
				t.Fail()
			}
		case "DELETE":
			if name != "teststruct_delete" || path != "/v1/teststruct/delete" {
				t.Fail()
			}
		default:
			t.Fail()
		}
	})
}
