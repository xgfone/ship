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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRenderer(t *testing.T) {
	r := SimpleRenderer("plain", "text/plain", func(v interface{}) ([]byte, error) {
		return []byte(fmt.Sprintf("%v", v)), nil
	})

	s := New()
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
	r := RendererFunc(func(ctx *Context, name string, code int, v interface{}) error {
		return ctx.String(code, fmt.Sprintf("%v", v))
	})

	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := r.Render(s.AcquireContext(req, rec), "plain", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `data` {
		t.Error(rec.Body.String())
	}
}

func TestShipRenderer(t *testing.T) {
	s := New()
	s.Route("/json").GET(func(ctx *Context) error { return ctx.Render("json", 200, "json") })
	s.Route("/jsonpretty").GET(func(ctx *Context) error { return ctx.Render("jsonpretty", 200, "jsonpretty") })
	s.Route("/xml").GET(func(ctx *Context) error { return ctx.Render("xml", 200, "xml") })
	s.Route("/xmlpretty").GET(func(ctx *Context) error { return ctx.Render("xmlpretty", 200, "xmlpretty") })

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

//////////////////////////////////////////////////////////////////////////////

func TestJSON(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := JSONRenderer().Render(s.AcquireContext(req, rec), "json", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := JSONRenderer(json.Marshal).Render(s.AcquireContext(req, rec), "json", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}
}

func TestJSONPretty(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := JSONPrettyRenderer("").Render(s.AcquireContext(req, rec), "jsonpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := JSONPrettyRenderer("", func(v interface{}) ([]byte, error) {
		return json.MarshalIndent(v, "", "    ")
	}).Render(s.AcquireContext(req, rec), "jsonpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != `"data"` {
		t.Error(rec.Body.String())
	}
}

//////////////////////////////////////////////////////////////////////////////

func TestXML(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := XMLRenderer().Render(s.AcquireContext(req, rec), "xml", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := XMLRenderer(xml.Marshal).Render(s.AcquireContext(req, rec), "xml", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}
}

func TestXMLPretty(t *testing.T) {
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	if err := XMLPrettyRenderer("   ").Render(s.AcquireContext(req, rec), "xmlpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	if err := XMLPrettyRenderer("", func(v interface{}) ([]byte, error) {
		return xml.MarshalIndent(v, "", "    ")
	}).Render(s.AcquireContext(req, rec), "xmlpretty", 200, "data"); err != nil {
		t.Error(err)
	}

	if rec.Body.String() != "<string>data</string>" {
		t.Error(rec.Body.String())
	}
}
