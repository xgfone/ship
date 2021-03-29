// Copyright 2019 xgfone
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
	"testing"

	"github.com/xgfone/ship/v4"
)

// Refer to github.com/labstack/echo/middleware#TestCORS
func TestCORS(t *testing.T) {
	r := ship.New()

	// Wildcard origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := r.AcquireContext(req, rec)
	h := CORS(nil)(ship.NotFoundHandler())
	h(ctx)
	if rec.Header().Get(ship.HeaderAccessControlAllowOrigin) != "*" {
		t.Errorf("%s is not *", ship.HeaderAccessControlAllowOrigin)
	}

	// Allow origins
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	ctx = r.AcquireContext(req, rec)
	h = CORS(&CORSConfig{AllowOrigins: []string{"localhost"}})(ship.NotFoundHandler())
	req.Header.Set(ship.HeaderOrigin, "localhost")
	h(ctx)
	if rec.Header().Get(ship.HeaderAccessControlAllowOrigin) != "localhost" {
		t.Errorf("%s is not 'localhost'", ship.HeaderAccessControlAllowOrigin)
	}

	// Preflight request
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	ctx = r.AcquireContext(req, rec)
	req.Header.Set(ship.HeaderOrigin, "localhost")
	req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationJSON)
	cors := CORS(&CORSConfig{AllowOrigins: []string{"localhost"}, AllowCredentials: true, MaxAge: 3600})
	h = cors(ship.NotFoundHandler())
	h(ctx)
	if rec.Header().Get(ship.HeaderAccessControlAllowOrigin) != "localhost" {
		t.Errorf("%s is not 'localhost'", ship.HeaderAccessControlAllowOrigin)
	} else if rec.Header().Get(ship.HeaderAccessControlAllowMethods) == "" {
		t.Fail()
	} else if rec.Header().Get(ship.HeaderAccessControlAllowCredentials) != "true" {
		t.Errorf("%s is not true", ship.HeaderAccessControlAllowCredentials)
	} else if rec.Header().Get(ship.HeaderAccessControlMaxAge) != "3600" {
		t.Errorf("%s is not 3600", ship.HeaderAccessControlMaxAge)
	}

	// Preflight request with `AllowOrigins` *
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	ctx = r.AcquireContext(req, rec)
	req.Header.Set(ship.HeaderOrigin, "localhost")
	req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationJSON)
	cors = CORS(&CORSConfig{AllowOrigins: []string{"*"}, AllowCredentials: true, MaxAge: 3600})
	h = cors(ship.NotFoundHandler())
	h(ctx)
	if rec.Header().Get(ship.HeaderAccessControlAllowOrigin) != "localhost" {
		t.Errorf("%s is not 'localhost'", ship.HeaderAccessControlAllowOrigin)
	} else if rec.Header().Get(ship.HeaderAccessControlAllowMethods) == "" {
		t.Fail()
	} else if rec.Header().Get(ship.HeaderAccessControlAllowCredentials) != "true" {
		t.Errorf("%s is not true", ship.HeaderAccessControlAllowCredentials)
	} else if rec.Header().Get(ship.HeaderAccessControlMaxAge) != "3600" {
		t.Errorf("%s is not 3600", ship.HeaderAccessControlMaxAge)
	}

	// Preflight request with `AllowOrigins` which allow all subdomains with *
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	rec = httptest.NewRecorder()
	ctx = r.AcquireContext(req, rec)
	req.Header.Set(ship.HeaderOrigin, "http://aaa.example.com")
	cors = CORS(&CORSConfig{AllowOrigins: []string{"http://*.example.com"}})
	h = cors(ship.NotFoundHandler())
	h(ctx)
	if s := rec.Header().Get(ship.HeaderAccessControlAllowOrigin); s != "http://aaa.example.com" {
		t.Errorf("%s: expect '%s', got '%s'", ship.HeaderAccessControlAllowOrigin,
			"http://aaa.example.com", s)
	}

	req.Header.Set(ship.HeaderOrigin, "http://bbb.example.com")
	h(ctx)
	if s := rec.Header().Get(ship.HeaderAccessControlAllowOrigin); s != "http://bbb.example.com" {
		t.Errorf("%s: expect '%s', got '%s'", ship.HeaderAccessControlAllowOrigin,
			"http://bbb.example.com", s)
	}
}
