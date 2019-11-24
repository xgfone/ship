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

package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/ship/v2"
)

func TestRecover(t *testing.T) {
	bs := bytes.NewBuffer(nil)
	router := ship.New().Use(Recover())
	router.HandleError = func(ctx *ship.Context, err error) {
		bs.WriteString(err.Error())
	}

	router.Route("/panic").GET(func(ctx *ship.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if bs.String() != "test panic" {
		t.Fail()
	}
}
