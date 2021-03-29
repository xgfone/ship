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
		ctx.Response().Write([]byte("test"))
		return nil
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
	} else if rec.Header().Get(ship.HeaderContentEncoding) != "" {
		t.Fail()
	} else if rec.Header().Get(ship.HeaderContentType) != "" {
		t.Fail()
	} else if rec.Body.Len() != 0 {
		t.Fail()
	}
}

func TestGzipErrorReturned(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	s := ship.New()
	s.Logger = ship.NewLoggerFromWriter(buf, "", 0)
	s.Use(Gzip(nil))
	s.Route("/").GET(func(ctx *ship.Context) error { return ship.ErrNotFound })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ship.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, rec.Code)
	} else if rec.Header().Get(ship.HeaderContentEncoding) != "" {
		t.Fail()
	}
}
