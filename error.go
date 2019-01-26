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
	"errors"

	"github.com/xgfone/ship/core"
)

// Re-import some HTTP errors from the sub-package core.
var (
	ErrUnsupportedMediaType        = core.ErrUnsupportedMediaType
	ErrNotFound                    = core.ErrNotFound
	ErrUnauthorized                = core.ErrUnauthorized
	ErrForbidden                   = core.ErrForbidden
	ErrMethodNotAllowed            = core.ErrMethodNotAllowed
	ErrStatusRequestEntityTooLarge = core.ErrStatusRequestEntityTooLarge
	ErrTooManyRequests             = core.ErrTooManyRequests
	ErrBadRequest                  = core.ErrBadRequest
	ErrBadGateway                  = core.ErrBadGateway
	ErrInternalServerError         = core.ErrInternalServerError
	ErrRequestTimeout              = core.ErrRequestTimeout
	ErrServiceUnavailable          = core.ErrServiceUnavailable
)

// Some non-HTTP Errors
var (
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")
	ErrCookieNotFound        = errors.New("cookie not found")
	ErrNoHandler             = errors.New("no handler")
	ErrNoSession             = errors.New("no session support")
)

// ErrSkip is the alias of core.ErrSkip, which is not an error and used to
// suggest that the middeware should skip and return it back to the outer
// middleware to handle.
//
// Notice: it is only a suggestion.
var ErrSkip = core.ErrSkip

// HTTPError is the alias of core.HTTPError, which stands for an HTTP error.
//
// Methods:
//    Code() int
//    Message() string
//    Error() string
//    ContentType() string
//    SetContentType(string) HTTPError
//    InnerError() error
//    SetInnerError(error) HTTPError
type HTTPError = core.HTTPError

// NewHTTPError returns a new HTTPError.
func NewHTTPError(code int, msg ...string) HTTPError {
	return core.NewHTTPError(code, msg...)
}
