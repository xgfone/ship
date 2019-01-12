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
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xgfone/ship"
)

func TestCSRF(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)
	csrf := CSRF(CSRFConfig{GenerateToken: GenerateToken(16)})

	handler := csrf(func(ctx ship.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Generate CSRF token
	handler(ctx)
	assert.Contains(t, rec.Header().Get(ship.HeaderSetCookie), "_csrf")

	// Without CSRF cookie
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	ctx = s.NewContext(req, rec)
	assert.Error(t, handler(ctx))

	// Empty/invalid CSRF token
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	ctx = s.NewContext(req, rec)
	req.Header.Set(ship.HeaderXCSRFToken, "")
	assert.Error(t, handler(ctx))

	// Valid CSRF token
	token := GenerateToken(16)()
	req.Header.Set(ship.HeaderCookie, "_csrf="+token)
	req.Header.Set(ship.HeaderXCSRFToken, token)
	if assert.NoError(t, handler(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCSRFTokenFromForm(t *testing.T) {
	form := make(url.Values)
	form.Set("csrf", "token")

	s := ship.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Add(ship.HeaderContentType, ship.MIMEApplicationForm)
	ctx := s.NewContext(req, nil)

	token, err := GetTokenFromForm("csrf")(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = GetTokenFromForm("invalid")(ctx)
	assert.Error(t, err)
}

func TestCSRFTokenFromQuery(t *testing.T) {
	form := make(url.Values)
	form.Set("csrf", "token")

	s := ship.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Add(ship.HeaderContentType, ship.MIMEApplicationForm)
	req.URL.RawQuery = form.Encode()
	ctx := s.NewContext(req, nil)

	token, err := GetTokenFromQuery("csrf")(ctx)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = GetTokenFromQuery("invalid")(ctx)
	assert.Error(t, err)
}
