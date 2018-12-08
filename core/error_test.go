package core

import (
	"errors"
	"testing"
)

func TestHTTPError(t *testing.T) {
	err := NewHTTPError(200, "OK")
	if err.Code() != 200 || err.Message() != "OK" {
		t.Fail()
	}

	if err.Error() != "code=200, msg=OK" {
		t.Fail()
	}

	err = err.SetInnerError(errors.New("inner error"))
	if err.InnerError().Error() != "inner error" {
		t.Fail()
	}
}
