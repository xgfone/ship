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

package ship

import (
	"errors"
	"fmt"
	"net/http"
)

// Some non-HTTP Errors
var (
	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")
	ErrCookieNotFound        = errors.New("cookie not found")
	ErrNoHandler             = errors.New("no handler")
	ErrNoSessionSupport      = errors.New("no session support")
	ErrInvalidSession        = errors.New("invalid session")
	ErrSessionNotExist       = errors.New("session does not exist")
	ErrMissingContentType    = errors.New("missing the header 'Content-Type'")
)

// Some HTTP error.
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
)

// ErrSkip is not an error, which is used to suggest that the middeware should
// skip and return it back to the outer middleware to handle.
//
// Notice: it is only a suggestion.
var ErrSkip = errors.New("skip")

// HTTPError represents an error with HTTP Status Code.
type HTTPError struct {
	Code int
	Msg  string
	Err  error
	CT   string // For Content-Type
}

// NewHTTPError returns a new HTTPError.
func NewHTTPError(code int, msg ...string) HTTPError {
	if len(msg) > 0 {
		return HTTPError{Code: code, Msg: msg[0]}
	}
	return HTTPError{Code: code}
}

func (e HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

// GetError returns the inner error.
//
// If Err is nil but Msg is not "", return `errors.New(e.Msg)` instead;
// Or return nil.
//
//     HTTPError{Err: errors.New("")}.GetError() != nil
//     HTTPError{Msg: "xxx"}.GetError() != nil
//     HTTPError{Code: 200}.GetError() == nil
func (e HTTPError) GetError() error {
	if e.Err != nil {
		return e.Err
	} else if e.Msg != "" {
		return errors.New(e.Msg)
	}
	return nil
}

// GetMsg returns a message.
func (e HTTPError) GetMsg() string {
	if e.Msg != "" {
		return e.Msg
	} else if e.Code < 500 && e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// NewError returns a new HTTPError with the new error.
func (e HTTPError) NewError(err error) HTTPError {
	nerr := e
	nerr.Err = err
	return nerr
}

// NewMsg returns a new HTTPError with the new msg.
func (e HTTPError) NewMsg(msg string, args ...interface{}) HTTPError {
	nerr := e
	if len(args) == 0 {
		nerr.Msg = msg
	} else {
		nerr.Msg = fmt.Sprintf(msg, args...)
	}
	return nerr
}

// NewCT returns a new HTTPError with the new ContentType ct.
func (e HTTPError) NewCT(ct string) HTTPError {
	nerr := e
	nerr.CT = ct
	return nerr
}
