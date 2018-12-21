package ship

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShipRenderer(t *testing.T) {
	s := New()
	s.Route("/json").GET(func(ctx Context) error { return ctx.Render("json", 200, "json") })
	s.Route("/jsonpretty").GET(func(ctx Context) error { return ctx.Render("jsonpretty", 200, "jsonpretty") })
	s.Route("/xml").GET(func(ctx Context) error { return ctx.Render("xml", 200, "xml") })
	s.Route("/xmlpretty").GET(func(ctx Context) error { return ctx.Render("xmlpretty", 200, "xmlpretty") })

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `"json"`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/jsonpretty", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `"jsonpretty"`, rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/xml", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<string>xml</string>", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/xmlpretty", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "<string>xmlpretty</string>", rec.Body.String())
}

func TestShipMuxRender(t *testing.T) {
	s := New()
	if mr := s.MuxRender(); mr == nil {
		t.Fail()
	}
}
