package render_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/render"
)

func TestXML(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := render.XML().Render(s.AcquireContext(req, rec), "xml", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := render.XML(xml.Marshal).Render(s.AcquireContext(req, rec), "xml", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}
}

func TestXMLPretty(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := render.XMLPretty("   ").Render(s.AcquireContext(req, rec), "xmlpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := render.XMLPretty("", func(v interface{}) ([]byte, error) {
		return xml.MarshalIndent(v, "", "    ")
	}).Render(s.AcquireContext(req, rec), "xmlpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}
}
