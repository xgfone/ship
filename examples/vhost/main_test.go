package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/router", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "default", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host1.example.com"
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "vhost1", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host2.example.com"
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "vhost2", rec.Body.String())
}
