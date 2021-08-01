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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/ship/v5"
)

func TestBodyLimitReader(t *testing.T) {
	bs := []byte("Hello, World")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))

	// reader all should return ErrStatusRequestEntityTooLarge
	reader := &limitedReader{limit: 6}
	reader.Reset(req.Body)
	_, err := ioutil.ReadAll(reader)
	he := err.(ship.HTTPServerError)
	if he.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("StatusCode: expect %d, got %d",
			http.StatusRequestEntityTooLarge, he.Code)
	}

	// reset reader and read six bytes must succeed.
	buf := make([]byte, 6)
	reader.Reset(ioutil.NopCloser(bytes.NewReader(bs)))
	n, err := reader.Read(buf)
	if n != 6 {
		t.Fail()
	} else if err != nil {
		t.Error(err)
	} else if s := string(buf); s != "Hello," {
		t.Errorf("expect '%s', got '%s'", "Hello,", s)
	}
}

func TestBodyLimit(t *testing.T) {
	bs := "Hello, World"
	limit := int64(2 * 1024 * 1024) // 2M
	s := ship.New()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(bs)))
	rec := httptest.NewRecorder()
	ctx := s.AcquireContext(req, rec)

	handler := func(ctx *ship.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		return ctx.Text(http.StatusOK, string(body))
	}

	// Based on content length (within limit)
	if err := BodyLenLimit(limit)(handler)(ctx); err != nil {
		t.Error(err)
	} else if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if s := rec.Body.String(); s != bs {
		t.Errorf("Body: expect '%s', got '%s'", bs, s)
	}

	// Based on content read (overlimit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(bs)))
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	he := BodyLenLimit(6)(handler)(ctx).(ship.HTTPServerError)
	if he.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("StatusCode: expect %d, got %d",
			http.StatusRequestEntityTooLarge, he.Code)
	}

	// Based on content read (within limit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(bs)))
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	if err := BodyLenLimit(limit)(handler)(ctx); err != nil {
		t.Error(err)
	} else if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if s := rec.Body.String(); s != bs {
		t.Errorf("Body: expect '%s', got '%s'", bs, s)
	}

	// Based on content read (overlimit)'
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(bs)))
	rec = httptest.NewRecorder()
	ctx = s.AcquireContext(req, rec)
	he = BodyLenLimit(6)(handler)(ctx).(ship.HTTPServerError)
	if he.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("StatusCode: expect %d, got %d",
			http.StatusRequestEntityTooLarge, he.Code)
	}
}
