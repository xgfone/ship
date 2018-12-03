package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	if http.StatusOK != w.Code {
		t.Fail()
	}
	if "{\"message\":\"pong\"}" != w.Body.String() {
		t.Fail()
	}
}
