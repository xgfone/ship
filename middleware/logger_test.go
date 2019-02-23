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
	"strings"
	"testing"
	"time"

	"github.com/xgfone/ship"
)

func TestLogger(t *testing.T) {
	// We fix the timestamp.
	startTime := time.Date(2018, time.December, 3, 14, 10, 0, 0, time.UTC)
	addTime := time.Duration(60)
	getNow := func() time.Time {
		startTime = startTime.Add(addTime)
		return startTime
	}

	bs := bytes.NewBuffer(nil)
	logger := ship.NewNoLevelLogger(bs, 0)

	router := ship.New(ship.SetLogger(logger))
	router.Use(Logger(getNow))

	router.Route("/test").GET(func(ctx *ship.Context) error {
		ctx.Logger().Info("handler")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// We removes the cost string, which is uncontrollable.
	ss := strings.Split(strings.TrimSpace(bs.String()), "\n")
	if ss[0] != "[I] handler" {
		t.Fail()
	}
	if strings.Join(strings.Split(ss[1], ",")[:3], ",") !=
		"[I] method=GET, url=/test, starttime=1543846200" {
		t.Fail()
	}
}
