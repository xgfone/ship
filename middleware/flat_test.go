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
	assert.Equal(t, "[INFO] before handling the request\n[INFO] handling the request\n[INFO] after handling the request\n",
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
	assert.Equal(t, "[INFO] before handling the request\n[EROR] before error\n",
		buf.String())
}
