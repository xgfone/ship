// Copyright 2018 xgfone
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

package middleware

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xgfone/ship/v4"
)

func TestGzip(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := s.AcquireContext(req, rec)

	// Skip if no Accept-Encoding header
	handler := Gzip(nil)(func(ctx *ship.Context) error {
		return ctx.Text(200, "test")
	})

	handler(ctx)
	if s := rec.Body.String(); s != "test" {
		t.Errorf("Body: expect '%s', got '%s'", "test", s)
	}

	// Gzip
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	handler(ctx)
	if v := rec.Header().Get(ship.HeaderContentEncoding); v != "gzip" {
		t.Errorf("%s: expect '%s', got '%s'", ship.HeaderContentEncoding, "gzip", v)
	} else if v = rec.Header().Get(ship.HeaderContentType); !strings.Contains(v, ship.MIMETextPlain) {
		t.Errorf("%s is '%s', not contain '%s'", ship.HeaderContentType, v, ship.MIMETextPlain)
	} else if reader, err := gzip.NewReader(rec.Body); err != nil {
		t.Error(err)
	} else {
		buf := new(bytes.Buffer)
		defer reader.Close()
		buf.ReadFrom(reader)
		if buf.String() != "test" {
			t.Errorf("GZIP: expect '%s', got '%s'", "test", buf.String())
		}
	}
}

func TestGzipNoContent(t *testing.T) {
	s := ship.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	ctx := s.AcquireContext(req, rec)
	handler := Gzip(nil)(func(ctx *ship.Context) error {
		return ctx.NoContent(http.StatusNoContent)
	})

	if err := handler(ctx); err != nil {
		t.Error(err)
	} else if ce := rec.Header().Get(ship.HeaderContentEncoding); ce != "gzip" {
		t.Errorf("expect the header Content-Encoding '%s', but got '%s'", "gzip", ce)
	} else if ct := rec.Header().Get(ship.HeaderContentType); ct != "" {
		t.Errorf("unexpect the header Content-Type, but got '%s'", ct)
	} else if r, err := gzip.NewReader(rec.Body); err != nil {
		t.Errorf("got an unexpected error when newing gzip reader: %s", err)
	} else if data, err := ioutil.ReadAll(r); err != nil {
		t.Errorf("got an unexpected error when reading gzip data: %s", err)
	} else if s := string(data); s != "" {
		t.Errorf("unexpect response data, but got '%s'", s)
	}
}

func TestGzipErrorReturned(t *testing.T) {
	s := ship.New().Use(Gzip(nil), HandleError())
	s.Route("/").GET(func(ctx *ship.Context) error { return ship.ErrNotFound })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expect statuscode '%d', but got '%d'", http.StatusNotFound, rec.Code)
	} else if ce := rec.Header().Get(ship.HeaderContentEncoding); ce != "gzip" {
		t.Errorf("expect the header Conent-Encoding '%s', but got '%s'", "gzip", ce)
	} else if r, err := gzip.NewReader(rec.Body); err != nil {
		t.Errorf("got an unexpected error when newing gzip reader: %s", err)
	} else if data, err := ioutil.ReadAll(r); err != nil {
		t.Errorf("got an unexpected error when reading gzip data: %s", err)
	} else if s := string(data); s != "Not Found" {
		t.Errorf("expect response data '%s', but got '%s'", "Not Found", s)
	}
}

func TestGzipDomains(t *testing.T) {
	s := ship.New().Use(Gzip(&GZipConfig{Domains: []string{
		"www1.example.com", "*.suffix.com", "www.prefix.*",
	}}))
	s.Route("/").GET(ship.OkHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www1.example.com"
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if ce := rec.Header().Get(ship.HeaderContentEncoding); ce != "gzip" {
		t.Errorf("expect the header Conent-Encoding '%s', but got '%s'", "gzip", ce)
	} else if r, err := gzip.NewReader(rec.Body); err != nil {
		t.Errorf("got an unexpected error when newing gzip reader: %s", err)
	} else if data, err := ioutil.ReadAll(r); err != nil {
		t.Errorf("got an unexpected error when reading gzip data: %s", err)
	} else if s := string(data); s != "OK" {
		t.Errorf("expect response data '%s', but got '%s'", "OK", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www2.example.com"
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Header().Get(ship.HeaderContentEncoding) == "gzip" {
		t.Errorf("unexpect the header Conent-Encoding 'gzip'")
	} else if data, err := ioutil.ReadAll(rec.Body); err != nil {
		t.Errorf("got an unexpected error when reading response data: %s", err)
	} else if s := string(data); s != "OK" {
		t.Errorf("expect response data '%s', but got '%s'", "OK", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www.suffix.com"
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if ce := rec.Header().Get(ship.HeaderContentEncoding); ce != "gzip" {
		t.Errorf("expect the header Conent-Encoding '%s', but got '%s'", "gzip", ce)
	} else if r, err := gzip.NewReader(rec.Body); err != nil {
		t.Errorf("got an unexpected error when newing gzip reader: %s", err)
	} else if data, err := ioutil.ReadAll(r); err != nil {
		t.Errorf("got an unexpected error when reading gzip data: %s", err)
	} else if s := string(data); s != "OK" {
		t.Errorf("expect response data '%s', but got '%s'", "OK", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "www.prefix.com"
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if ce := rec.Header().Get(ship.HeaderContentEncoding); ce != "gzip" {
		t.Errorf("expect the header Conent-Encoding '%s', but got '%s'", "gzip", ce)
	} else if r, err := gzip.NewReader(rec.Body); err != nil {
		t.Errorf("got an unexpected error when newing gzip reader: %s", err)
	} else if data, err := ioutil.ReadAll(r); err != nil {
		t.Errorf("got an unexpected error when reading gzip data: %s", err)
	} else if s := string(data); s != "OK" {
		t.Errorf("expect response data '%s', but got '%s'", "OK", s)
	}
}
