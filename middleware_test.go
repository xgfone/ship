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

package ship

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func ExampleMiddleware() {
	// We fix the timestamp.
	startTime := time.Date(2018, time.December, 3, 14, 10, 0, 0, time.Local)
	addTime := time.Duration(60)
	getNow := func() time.Time {
		startTime = startTime.Add(addTime)
		return startTime
	}

	bs := bytes.NewBuffer(nil)
	logger := NewNoLevelLogger(bs, 0)

	router := NewRouter(Config{Logger: logger})
	router.Use(NewLoggerMiddleware(getNow), NewRecoverMiddleware())

	router.Get("/test", func(ctx Context) error {
		ctx.Logger().Info("handler")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// We removes the cost string, which is uncontrollable.
	ss := strings.Split(strings.TrimSpace(bs.String()), "\n")
	fmt.Println(ss[0])
	fmt.Println(strings.Join(strings.Split(ss[1], ",")[:3], ","))

	// Output:
	// [INFO] handler
	// [INFO] method=GET, url=/test, starttime=1543817400
}
