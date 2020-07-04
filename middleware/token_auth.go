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

	"github.com/xgfone/ship/v3"
)

// TokenAuth returns a TokenAuth middleware.
//
// For valid key it will calls the next handler.
// For invalid key, it responds "401 Unauthorized".
// For missing key, it responds "400 Bad Request".
//
// If getToken is missing, the default is
// GetTokenFromHeader(ship.HeaderAuthorization, "Bearer").
func TokenAuth(validator TokenValidator, getToken ...TokenFunc) Middleware {
	getAuthToken := GetTokenFromHeader(ship.HeaderAuthorization, "Bearer")
	if len(getToken) > 0 && getToken[0] != nil {
		getAuthToken = getToken[0]
	}

	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) error {
			token, err := getAuthToken(ctx)
			if err != nil {
				if _, ok := err.(ship.HTTPError); ok {
					return err
				}
				return ship.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			if valid, err := validator(token); err != nil {
				return err
			} else if valid {
				return next(ctx)
			}
			return ship.ErrUnauthorized
		}
	}
}
