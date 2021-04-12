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

package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xgfone/ship/v4"
)

func TestLogger(t *testing.T) {
	bs := bytes.NewBuffer(nil)
	logger := ship.NewLoggerFromWriter(bs, "", 0)

	router := ship.New()
	router.Logger = logger
	router.Use(Logger(&LoggerConfig{LogReqBody: true}))

	router.Route("/test").GET(func(ctx *ship.Context) error {
		ctx.Logger().Infof("handler")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewBufferString("body"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Remove the starttime and the cost, which is uncontrollable.
	ss := strings.Split(strings.TrimSpace(bs.String()), "\n")
	if len(ss) != 2 {
		t.Errorf("expected two lines, but got '%d'", len(ss))
	} else if ss[0] != "[I] handler" {
		t.Errorf("expected '[I] handler', but got '%s'", ss[0])
	} else if s := strings.Join(strings.Split(ss[1], ", ")[1:5], ", "); s !=
		`method=GET, path=/test, reqbody=body, code=200` {
		t.Error(s)
	}
}

func TestLoggerLog(t *testing.T) {
	router := ship.New()

	var reqlog string
	router.Use(Logger(&LoggerConfig{
		LogReqBody: true,
		Log: func(r *http.Request, _ bool, reqBody string, code int,
			_ time.Time, _ time.Duration, err error) {
			reqlog = fmt.Sprintf("method=%s, uri=%s, reqbody=%s, code=%d, err=%v",
				r.Method, r.RequestURI, reqBody, code, err)
		},
	}))

	router.Route("/test").GET(func(ctx *ship.Context) error {
		ctx.Logger().Infof("handler")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", bytes.NewBufferString("body"))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	expect := "method=GET, uri=/test, reqbody=body, code=200, err=<nil>"
	if expect != reqlog {
		t.Error(reqlog)
	}
}
