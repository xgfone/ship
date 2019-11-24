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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/xgfone/ship/v2"
)

func TestMaxRequests(t *testing.T) {
	sleep := time.Millisecond * 300
	wg := sync.WaitGroup{}
	wg.Add(3)

	s := ship.New()
	s.Use(MaxRequests(2, func(c *ship.Context) error {
		c.NoContent(http.StatusTooManyRequests)
		wg.Done()
		return nil
	}))
	s.R("/").GET(func(ctx *ship.Context) error {
		time.Sleep(sleep)
		wg.Done()
		return nil
	})

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec3 := httptest.NewRecorder()

	go s.ServeHTTP(rec1, req1)
	go s.ServeHTTP(rec2, req2)
	go s.ServeHTTP(rec3, req3)

	wg.Wait()
	time.Sleep(sleep)
	if rec1.Code+rec2.Code+rec3.Code != 200+200+429 {
		t.Errorf("req1=%d, req2=%d, req3=%d", rec1.Code, rec2.Code, rec3.Code)
	}
}
