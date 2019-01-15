package ship

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetContext(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)

	if GetContext(ctx.Request()) != ctx {
		t.Fail()
	}
}
