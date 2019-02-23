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

	"github.com/stretchr/testify/assert"
	"github.com/xgfone/ship"
)

func TestAuthToken(t *testing.T) {
	assert := assert.New(t)
	s := ship.New()

	validateToken := func(token string) (bool, error) {
		return token == "valid_token", nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)
	authMiddleware := TokenAuth(validateToken, GetTokenFromHeader(ship.HeaderAuthorization, "abc"))
	handler := authMiddleware(func(ctx *ship.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Valid AuthToken
	auth := "abc valid_token"
	req.Header.Set(ship.HeaderAuthorization, auth)
	assert.NoError(handler(ctx))

	// Invalid AuthToken
	auth = "abc invalid_token"
	req.Header.Set(ship.HeaderAuthorization, auth)
	he := handler(ctx).(ship.HTTPError)
	assert.Equal(http.StatusUnauthorized, he.Code)

	// Missing Authorization header
	req.Header.Del(ship.HeaderAuthorization)
	he = handler(ctx).(ship.HTTPError)
	assert.Equal(http.StatusBadRequest, he.Code)

	// Token from custom header
	handler = TokenAuth(validateToken, GetTokenFromHeader("API-Token"))(
		func(ctx *ship.Context) error {
			return ctx.String(http.StatusOK, "test")
		})
	req.Header.Set("API-Token", "valid_token")
	assert.NoError(handler(ctx))

	// Token from URL query
	handler = TokenAuth(validateToken, GetTokenFromQuery("token"))(
		func(ctx *ship.Context) error {
			return ctx.String(http.StatusOK, "test")
		})
	query := req.URL.Query()
	query.Add("token", "valid_token")
	ctx.Request().URL.RawQuery = query.Encode()
	assert.NoError(handler(ctx))

	// Token from Form
	handler = TokenAuth(validateToken, GetTokenFromForm("token"))(
		func(ctx *ship.Context) error {
			return ctx.String(http.StatusOK, "test")
		})
	form := make(url.Values)
	form.Set("token", "valid_token")
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationForm)
	ctx = s.NewContext(req, rec)
	assert.NoError(handler(ctx))
}
