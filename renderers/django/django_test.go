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

package django

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xgfone/ship"
)

func TestEngine(t *testing.T) {
	htmlData := `<html><head></head><body>{{ data }}</body></html>`
	filename := "_test_django_engine_.html"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		t.Fail()
		return
	}
	file.WriteString(htmlData)
	file.Close()
	defer os.Remove(filename)

	buf := bytes.NewBuffer(nil)
	engine := New(".")
	err = engine.Execute(buf, filename, map[string]interface{}{"data": "abc"}, nil)
	if err != nil || buf.String() != "<html><head></head><body>abc</body></html>" {
		t.Fail()
	}
}

func TestRouterTemplate(t *testing.T) {
	htmlData := `<html><head></head><body>{{ data }}</body></html>`
	filename := "_test_django_engine_for_router_.html"

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		t.Fail()
		return
	}
	file.WriteString(htmlData)
	file.Close()
	defer os.Remove(filename)

	engine := New(".")

	s := ship.New()
	mr := s.MuxRenderer()
	mr.Add(engine.Ext(), ship.HTMLTemplateRenderer(engine))

	s.Route("/django").GET(func(ctx *ship.Context) error {
		return ctx.Render(filename, 200, map[string]interface{}{"data": "django"})
	})
	req := httptest.NewRequest(http.MethodGet, "/django", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "<html><head></head><body>django</body></html>", rec.Body.String())
}
