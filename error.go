// Copyright 2020 xgfone
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

	"github.com/xgfone/ship/v3/herror"
)

// Re-export some errors.
var (
	// Some non-HTTP errors
	ErrMissingContentType    = herror.ErrMissingContentType
	ErrRendererNotRegistered = herror.ErrRendererNotRegistered
	ErrInvalidRedirectCode   = herror.ErrInvalidRedirectCode
	ErrInvalidSession        = herror.ErrInvalidSession
	ErrSessionNotExist       = herror.ErrSessionNotExist
	ErrNoSessionSupport      = herror.ErrNoSessionSupport
	ErrNoResponder           = herror.ErrNoResponder

	// Some HTTP error.
	ErrBadRequest                    = herror.ErrBadRequest
	ErrUnauthorized                  = herror.ErrUnauthorized
	ErrForbidden                     = herror.ErrForbidden
	ErrNotFound                      = herror.ErrNotFound
	ErrMethodNotAllowed              = herror.ErrMethodNotAllowed
	ErrStatusNotAcceptable           = herror.ErrStatusNotAcceptable
	ErrRequestTimeout                = herror.ErrRequestTimeout
	ErrStatusConflict                = herror.ErrStatusConflict
	ErrStatusGone                    = herror.ErrStatusGone
	ErrStatusRequestEntityTooLarge   = herror.ErrStatusRequestEntityTooLarge
	ErrUnsupportedMediaType          = herror.ErrUnsupportedMediaType
	ErrTooManyRequests               = herror.ErrTooManyRequests
	ErrInternalServerError           = herror.ErrInternalServerError
	ErrStatusNotImplemented          = herror.ErrStatusNotImplemented
	ErrBadGateway                    = herror.ErrBadGateway
	ErrServiceUnavailable            = herror.ErrServiceUnavailable
	ErrStatusGatewayTimeout          = herror.ErrStatusGatewayTimeout
	ErrStatusHTTPVersionNotSupported = herror.ErrStatusHTTPVersionNotSupported

	// ErrSkip is not an error, which is used to suggest that the middeware
	// should skip and return it back to the outer middleware to handle.
	ErrSkip = herror.ErrSkip
)

// HTTPError is the alias of herror.HTTPError.
type HTTPError = herror.HTTPError

// HTTPServerError is the alias of HTTPError.
type HTTPServerError = HTTPError

// NewHTTPError is the alias of herror.NewHTTPError.
var NewHTTPError = herror.NewHTTPError

// RouteError represents a route error when adding a route.
type RouteError struct {
	RouteInfo
	Err error
}

func (re RouteError) Error() string {
	return fmt.Sprintf("%s: name=%s, path=%s, method=%s, host=%s",
		re.Err, re.Name, re.Path, re.Method, re.Host)
}
