package core

import (
	"errors"
	"fmt"
	"net/http"
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

// HTTPError stands for a HTTP error.
type HTTPError interface {
	Code() int
	Message() string
	Error() string
	ContentType() string
	SetContentType(string) HTTPError
	InnerError() error
	SetInnerError(error) HTTPError
}

type httpError struct {
	code int
	msg  string
	ct   string
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

func (he httpError) ContentType() string {
	return he.ct
}

func (he httpError) SetContentType(ct string) HTTPError {
	return httpError{code: he.code, msg: he.msg, err: he.err, ct: ct}
}

func (he httpError) InnerError() error {
	return he.err
}

func (he httpError) SetInnerError(err error) HTTPError {
	return httpError{code: he.code, msg: he.msg, ct: he.ct, err: err}
}
