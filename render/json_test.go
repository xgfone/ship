package render_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/render"
)

func TestJSON(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := render.JSON().Render(s.AcquireContext(req, rec), "json", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := render.JSON(json.Marshal).Render(s.AcquireContext(req, rec), "json", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}
}

func TestJSONPretty(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := render.JSONPretty("").Render(s.AcquireContext(req, rec), "jsonpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := render.JSONPretty("", func(v interface{}) ([]byte, error) {
		return json.MarshalIndent(v, "", "    ")
	}).Render(s.AcquireContext(req, rec), "jsonpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}
}
