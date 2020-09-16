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
	ErrBadRequest                    = NewHTTPError(http.StatusBadRequest)
	ErrUnauthorized                  = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                     = NewHTTPError(http.StatusForbidden)
	ErrNotFound                      = NewHTTPError(http.StatusNotFound)
	ErrMethodNotAllowed              = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusNotAcceptable           = NewHTTPError(http.StatusNotAcceptable)
	ErrRequestTimeout                = NewHTTPError(http.StatusRequestTimeout)
	ErrStatusConflict                = NewHTTPError(http.StatusConflict)
	ErrStatusGone                    = NewHTTPError(http.StatusGone)
	ErrStatusRequestEntityTooLarge   = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrUnsupportedMediaType          = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrTooManyRequests               = NewHTTPError(http.StatusTooManyRequests)
	ErrInternalServerError           = NewHTTPError(http.StatusInternalServerError)
	ErrStatusNotImplemented          = NewHTTPError(http.StatusNotImplemented)
	ErrBadGateway                    = NewHTTPError(http.StatusBadGateway)
	ErrServiceUnavailable            = NewHTTPError(http.StatusServiceUnavailable)
	ErrStatusGatewayTimeout          = NewHTTPError(http.StatusGatewayTimeout)
	ErrStatusHTTPVersionNotSupported = NewHTTPError(http.StatusHTTPVersionNotSupported)
)

// ErrSkip is not an error, which is used to suggest that the middeware should
// skip and return it back to the outer middleware to handle.
//
// Notice: it is only a suggestion.
var ErrSkip = errors.New("skip")

// HTTPError represents an error with HTTP Status Code.
type HTTPError struct {
	Code int
	Msg  string // DEPRECATED!!!
	Err  error
	CT   string // For Content-Type
}

// NewHTTPError returns a new HTTPError.
func NewHTTPError(code int, msg ...string) HTTPError {
	if len(msg) > 0 {
		return HTTPError{Code: code, Err: errors.New(msg[0])}
	}
	return HTTPError{Code: code}
}

func (e HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

// Unwrap unwraps the inner error.
func (e HTTPError) Unwrap() error { return e.Err }

// GetError returns the inner error.
//
// If Err is nil but Msg is not "", return `errors.New(e.Msg)` instead;
// Or return nil.
//
//     HTTPError{Err: errors.New("")}.GetError() != nil
//     HTTPError{Msg: "xxx"}.GetError() != nil
//     HTTPError{Code: 200}.GetError() == nil
//
// DEPRECATED!!!
func (e HTTPError) GetError() error {
	if e.Err != nil {
		return e.Err
	} else if e.Msg != "" {
		return errors.New(e.Msg)
	}
	return nil
}

// GetMsg returns a message.
//
// DEPRECATED!!!
func (e HTTPError) GetMsg() string {
	if e.Msg != "" {
		return e.Msg
	} else if e.Code < 500 && e.Err != nil {
		return e.Err.Error()
	}
	return ""
}

// NewCT returns a new HTTPError with the new ContentType ct.
func (e HTTPError) NewCT(ct string) HTTPError { e.CT = ct; return e }

// NewError returns a new HTTPError with the new error.
func (e HTTPError) NewError(err error) HTTPError { e.Err = err; return e }

// NewErrorf is equal to NewError(fmt.Errorf(msg, args...)).
func (e HTTPError) NewErrorf(msg string, args ...interface{}) HTTPError {
	if len(args) == 0 {
		return e.NewError(errors.New(msg))
	}
	return e.NewError(fmt.Errorf(msg, args...))
}

// NewMsg returns a new HTTPError with the new msg.
//
// DEPRECATED!!! Please use NewErrorf instead.
func (e HTTPError) NewMsg(msg string, args ...interface{}) HTTPError {
	if len(args) == 0 {
		e.Msg = msg
	} else {
		e.Msg = fmt.Sprintf(msg, args...)
	}
	return e
}
