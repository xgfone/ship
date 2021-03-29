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
	"errors"
	"fmt"
	"net/http"
)

// Some non-HTTP Errors
var (
	ErrMissingContentType    = errors.New("missing the header 'Content-Type'")
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")
	ErrInvalidSession        = errors.New("invalid session")
	ErrSessionNotExist       = errors.New("session does not exist")
	ErrNoSessionSupport      = errors.New("no session support")
	ErrNoResponder           = errors.New("no responder")
)

// Some HTTP error.
var (
	ErrBadRequest                    = NewHTTPServerError(http.StatusBadRequest)
	ErrUnauthorized                  = NewHTTPServerError(http.StatusUnauthorized)
	ErrForbidden                     = NewHTTPServerError(http.StatusForbidden)
	ErrNotFound                      = NewHTTPServerError(http.StatusNotFound)
	ErrMethodNotAllowed              = NewHTTPServerError(http.StatusMethodNotAllowed)
	ErrStatusNotAcceptable           = NewHTTPServerError(http.StatusNotAcceptable)
	ErrRequestTimeout                = NewHTTPServerError(http.StatusRequestTimeout)
	ErrStatusConflict                = NewHTTPServerError(http.StatusConflict)
	ErrStatusGone                    = NewHTTPServerError(http.StatusGone)
	ErrStatusRequestEntityTooLarge   = NewHTTPServerError(http.StatusRequestEntityTooLarge)
	ErrUnsupportedMediaType          = NewHTTPServerError(http.StatusUnsupportedMediaType)
	ErrTooManyRequests               = NewHTTPServerError(http.StatusTooManyRequests)
	ErrInternalServerError           = NewHTTPServerError(http.StatusInternalServerError)
	ErrStatusNotImplemented          = NewHTTPServerError(http.StatusNotImplemented)
	ErrBadGateway                    = NewHTTPServerError(http.StatusBadGateway)
	ErrServiceUnavailable            = NewHTTPServerError(http.StatusServiceUnavailable)
	ErrStatusGatewayTimeout          = NewHTTPServerError(http.StatusGatewayTimeout)
	ErrStatusHTTPVersionNotSupported = NewHTTPServerError(http.StatusHTTPVersionNotSupported)
)

// ErrSkip is not an error, which is used to suggest that the middeware should
// skip and return it back to the outer middleware to handle.
//
// Notice: it is only a suggestion.
var ErrSkip = errors.New("skip")

// RouteError represents a route error when adding a route.
type RouteError struct {
	RouteInfo
	Err error
}

func (re RouteError) Error() string {
	return fmt.Sprintf("%s: name=%s, path=%s, method=%s, host=%s",
		re.Err, re.Name, re.Path, re.Method, re.Host)
}

// HTTPServerError represents a server error with HTTP Status Code.
type HTTPServerError struct {
	Code int
	Err  error
	CT   string // Content-Type
}

// NewHTTPServerError returns a new HTTPServerError.
func NewHTTPServerError(code int, msg ...string) HTTPServerError {
	if len(msg) == 0 {
		return HTTPServerError{Code: code, Err: errors.New(http.StatusText(code))}
	}
	return HTTPServerError{Code: code, Err: errors.New(msg[0])}
}

func (e HTTPServerError) Error() string { return e.Err.Error() }

// Unwrap unwraps the inner error.
func (e HTTPServerError) Unwrap() error { return e.Err }

// NewCT returns a new HTTPError with the new ContentType ct.
func (e HTTPServerError) NewCT(ct string) HTTPServerError { e.CT = ct; return e }

// New returns a new HTTPError with the new error.
func (e HTTPServerError) New(err error) HTTPServerError { e.Err = err; return e }

// Newf is equal to New(fmt.Errorf(msg, args...)).
func (e HTTPServerError) Newf(msg string, args ...interface{}) HTTPServerError {
	if len(args) == 0 {
		return e.New(errors.New(msg))
	}
	return e.New(fmt.Errorf(msg, args...))
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
	var err string
	if e.Err != nil {
		err = fmt.Sprintf(", err=%s", e.Err.Error())
	}

	var data string
	if e.Data != "" {
		data = fmt.Sprintf(", data=%s", e.Data)
	}

	return fmt.Sprintf("method=%s, url=%s, code=%d%s%s",
		e.Method, e.URL, e.Code, data, err)
}
