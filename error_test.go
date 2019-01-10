package ship

import (
	"net/http"
	"testing"
)

func TestNewHTTPError(t *testing.T) {
	if NewHTTPError(http.StatusBadRequest) != ErrBadRequest {
		t.Fail()
	}
}
