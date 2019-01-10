package render_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/core"
	"github.com/xgfone/ship/render"
)

func TestSimpleRenderer(t *testing.T) {
	r := render.SimpleRenderer("plain", "text/plain", func(v interface{}) ([]byte, error) {
		return []byte(fmt.Sprintf("%v", v)), nil
	})

	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := r.Render(s.AcquireContext(req, rec), "plain", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `data` {
		t.Error(rec.Body.String())
	}
}

func TestRendererFunc(t *testing.T) {
	r := render.RendererFunc(func(ctx core.Context, name string, code int, v interface{}) error {
		return ctx.String(code, fmt.Sprintf("%v", v))
	})

	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := r.Render(s.AcquireContext(req, rec), "plain", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `data` {
		t.Error(rec.Body.String())
	}
}
