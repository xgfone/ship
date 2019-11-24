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
	"net/url"
	"strings"
	"testing"

	"github.com/xgfone/ship/v2"
)

func TestCSRF(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.AcquireContext(req, rec)
	csrf := CSRF(CSRFConfig{GenerateToken: GenerateToken(16)})

	handler := csrf(func(ctx *ship.Context) error {
		return ctx.Text(http.StatusOK, "test")
	})

	// Generate CSRF token
	handler(ctx)
	if v := rec.Header().Get(ship.HeaderSetCookie); !strings.Contains(v, "_csrf") {
		t.Fail()
	}

	// Without CSRF cookie
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	if handler(ctx) == nil {
		t.Fail()
	}

	// Empty/invalid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	req.Header.Set(ship.HeaderXCSRFToken, "")
	if handler(ctx) == nil {
		t.Fail()
	}

	// Valid CSRF token
	token := GenerateToken(16)()
	req.Header.Set(ship.HeaderCookie, "_csrf="+token)
	req.Header.Set(ship.HeaderXCSRFToken, token)
	if err := handler(ctx); err != nil {
		t.Error(err)
	} else if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestCSRFTokenFromForm(t *testing.T) {
	form := make(url.Values)
	form.Set("csrf", "token")

	s := ship.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Add(ship.HeaderContentType, ship.MIMEApplicationForm)
	ctx := s.AcquireContext(req, nil)

	if token, err := GetTokenFromForm("csrf")(ctx); err != nil {
		t.Error(err)
	} else if token != "token" {
		t.Errorf("token != '%s'", token)
	} else if _, err = GetTokenFromForm("invalid")(ctx); err == nil {
		t.Fail()
	}
}

func TestCSRFTokenFromQuery(t *testing.T) {
	form := make(url.Values)
	form.Set("csrf", "token")

	s := ship.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Add(ship.HeaderContentType, ship.MIMEApplicationForm)
	req.URL.RawQuery = form.Encode()
	ctx := s.AcquireContext(req, nil)

	if token, err := GetTokenFromQuery("csrf")(ctx); err != nil {
		t.Error(err)
	} else if token != "token" {
		t.Errorf("token != '%s'", token)
	} else if _, err = GetTokenFromQuery("invalid")(ctx); err == nil {
		t.Fail()
	}
}
