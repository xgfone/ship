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

package ship

import (
	"fmt"
	"net/http"
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrTooManyRequests             = NewHTTPError(http.StatusTooManyRequests)
	ErrBadRequest                  = NewHTTPError(http.StatusBadRequest)
	ErrBadGateway                  = NewHTTPError(http.StatusBadGateway)
	ErrInternalServerError         = NewHTTPError(http.StatusInternalServerError)
	ErrRequestTimeout              = NewHTTPError(http.StatusRequestTimeout)
	ErrServiceUnavailable          = NewHTTPError(http.StatusServiceUnavailable)

	ErrValidatorNotRegistered = fmt.Errorf("validator not registered")
	ErrRendererNotRegistered  = fmt.Errorf("renderer not registered")
	ErrInvalidRedirectCode    = fmt.Errorf("invalid redirect status code")
	ErrCookieNotFound         = fmt.Errorf("cookie not found")
)

type httpError struct {
	code int
	msg  string
	err  error
}

// NewHTTPError returns a new HTTPError.
func NewHTTPError(code int, msg ...string) HTTPError {
	m := http.StatusText(code)
	if len(msg) > 0 && msg[0] != "" {
		m = msg[0]
	}
	return httpError{code: code, msg: m}
}

func (he httpError) Code() int {
	return he.code
}

func (he httpError) Message() string {
	return he.msg
}

func (he httpError) Error() string {
	return fmt.Sprintf("code=%d, msg=%s", he.code, he.msg)
}

func (he httpError) InnerError() error {
	return he.err
}

func (he httpError) SetInnerError(err error) HTTPError {
	return httpError{code: he.code, msg: he.msg, err: err}
}
