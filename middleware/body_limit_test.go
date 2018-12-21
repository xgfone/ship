// Copyright 2018 xgfone <xgfone@126.com>
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

	"github.com/stretchr/testify/assert"
	"github.com/xgfone/ship"
)

func TestBodyLimitReader(t *testing.T) {
	bs := []byte("Hello, World")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))

	// reader all should return ErrStatusRequestEntityTooLarge
	reader := &limitedReader{limit: 6}
	reader.Reset(req.Body)
	_, err := ioutil.ReadAll(reader)
	he := err.(ship.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code())

	// reset reader and read six bytes must succeed.
	buf := make([]byte, 6)
	reader.Reset(ioutil.NopCloser(bytes.NewReader(bs)))
	n, err := reader.Read(buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte("Hello,"), buf)
}

func TestBodyLimit(t *testing.T) {
	bs := []byte("Hello, World")
	limit := int64(2 * 1024 * 1024) // 2M

	assert := assert.New(t)
	s := ship.New()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))
	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)

	handler := func(ctx ship.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		return ctx.String(http.StatusOK, string(body))
	}

	// Based on content length (within limit)
	if assert.NoError(BodyLimit(limit)(handler)(ctx)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(bs, rec.Body.Bytes())
	}

	// Based on content read (overlimit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))
	rec = httptest.NewRecorder()
	ctx = s.NewContext(req, rec)
	he := BodyLimit(6)(handler)(ctx).(ship.HTTPError)
	assert.Equal(http.StatusRequestEntityTooLarge, he.Code())

	// Based on content read (within limit)
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))
	rec = httptest.NewRecorder()
	ctx = s.NewContext(req, rec)
	if assert.NoError(BodyLimit(limit)(handler)(ctx)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(bs, rec.Body.Bytes())
	}

	// Based on content read (overlimit)'
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bs))
	rec = httptest.NewRecorder()
	ctx = s.NewContext(req, rec)
	he = BodyLimit(6)(handler)(ctx).(ship.HTTPError)
	assert.Equal(http.StatusRequestEntityTooLarge, he.Code())
}
