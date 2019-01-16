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

package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xgfone/ship"
)

func TestFlat(t *testing.T) {
	beforeLog := func(ctx ship.Context) error {
		ctx.Logger().Info("before handling the request")
		return nil
	}
	afterLog := func(ctx ship.Context) error {
		ctx.Logger().Info("after handling the request")
		return nil
	}

	buf := bytes.NewBuffer(nil)
	router := ship.New(ship.Config{Logger: ship.NewNoLevelLogger(buf, 0)})
	router.Use(Flat([]ship.Handler{beforeLog}, []ship.Handler{afterLog}))
	router.R("/").GET(func(ctx ship.Context) error {
		ctx.Logger().Info("handling the request")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[I] before handling the request\n[I] handling the request\n[I] after handling the request\n",
		buf.String())
}

func TestFlatFail(t *testing.T) {
	beforeLog := func(ctx ship.Context) error {
		ctx.Logger().Info("before handling the request")
		return fmt.Errorf("before error")
	}
	afterLog := func(ctx ship.Context) error {
		ctx.Logger().Info("after handling the request")
		return nil
	}

	buf := bytes.NewBuffer(nil)
	router := ship.New(ship.Config{Logger: ship.NewNoLevelLogger(buf, 0)})
	router.Use(Flat([]ship.Handler{beforeLog}, []ship.Handler{afterLog}))
	router.R("/").GET(func(ctx ship.Context) error {
		ctx.Logger().Info("handling the request")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 500, rec.Code)
	assert.Equal(t, "[I] before handling the request\n[E] before error\n",
		buf.String())
}

func TestFlatError(t *testing.T) {
	beforeLog := func(ctx ship.Context) error {
		ctx.Logger().Info("before handling the request")
		return nil
	}
	afterLog := func(ctx ship.Context) error {
		ctx.Logger().Info("after handling the request")
		return nil
	}

	buf := bytes.NewBuffer(nil)
	router := ship.New(ship.Config{Logger: ship.NewNoLevelLogger(buf, 0)})
	router.Use(Flat([]ship.Handler{beforeLog}, []ship.Handler{afterLog}))
	router.R("/").GET(func(ctx ship.Context) error {
		ctx.Logger().Info("handling the request")
		ctx.SetError(fmt.Errorf("handler error"))
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 500, rec.Code)
	assert.Equal(t, "[I] before handling the request\n[I] handling the request\n[E] handler error\n",
		buf.String())
}
