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

// Package herror is deprecated!
package herror

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

// HTTPError is the alias of HTTPServerError
type HTTPError = HTTPServerError

// NewHTTPError is the alias of NewHTTPServerError.
var NewHTTPError = NewHTTPServerError

// HTTPServerError represents a server error with HTTP Status Code.
type HTTPServerError struct {
	Code int
	Err  error
	Msg  string // DEPRECATED!!!
	CT   string // Content-Type
}

// NewHTTPServerError returns a new HTTPServerError.
func NewHTTPServerError(code int, msg ...string) HTTPServerError {
	if len(msg) > 0 {
		return HTTPServerError{Code: code, Err: errors.New(msg[0])}
	}
	return HTTPServerError{Code: code}
}

func (e HTTPServerError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

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

/////////////////////////////////////////////////////////////////////////////

// GetMsg is DEPRECATED!
func (e HTTPError) GetMsg() string {
	if e.Msg != "" {
		return e.Msg
	} else if e.Code < 500 && e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// GetError is DEPRECATED!
func (e HTTPServerError) GetError() error {
	if e.Err != nil {
		return e.Err
	} else if e.Msg != "" {
		return errors.New(e.Msg)
	}
	return nil
}

// NewError is DEPRECATED!
func (e HTTPError) NewError(err error) HTTPError { e.Err = err; return e }

// NewMsg is DEPRECATED!
func (e HTTPError) NewMsg(msg string, args ...interface{}) HTTPError {
	if len(args) == 0 {
		e.Msg = msg
	} else {
		e.Msg = fmt.Sprintf(msg, args...)
	}
	return e
}
