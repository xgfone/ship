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

// HTTPClientError represents an error about the http client response.
type HTTPClientError struct {
	Code   int    `json:"code" xml:"code"`
	Method string `json:"method" xml:"method"`
	URL    string `json:"url" xml:"url"`
	Data   string `json:"data" xml:"data"`
	Err    error  `json:"err" xml:"err"`
}

// NewHTTPClientError returns a new HTTPClientError.
func NewHTTPClientError(method, url string, code int, err error,
	data ...string) HTTPClientError {
	var d string
	if len(data) > 0 {
		d = data[0]
	}

	return HTTPClientError{Method: method, URL: url, Code: code, Data: d, Err: err}
}

func (e HTTPClientError) Unwrap() error  { return e.Err }
func (e HTTPClientError) String() string { return e.Error() }
func (e HTTPClientError) Error() string {
	if e.Data == "" {
		return fmt.Sprintf("method=%s, url=%s, code=%d, err=%s",
			e.Method, e.URL, e.Code, e.Err)
	}

	return fmt.Sprintf("method=%s, url=%s, code=%d, data=%s, err=%s",
		e.Method, e.URL, e.Code, e.Data, e.Err)
}
