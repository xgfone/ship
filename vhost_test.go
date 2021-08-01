// Copyright 2020 xgfone
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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVHost(t *testing.T) {
	vhosts := NewHostManagerHandler(nil)

	dhost := New()
	dhost.Route("/router").GET(func(c *Context) error { return c.Text(200, "default") })
	vhosts.SetDefaultHost("", dhost)

	vhost1 := New()
	vhost1.Route("/router").GET(func(c *Context) error { return c.Text(200, "vhost1") })
	vhosts.AddHost("*.host1.example.com", vhost1)

	vhost2 := New()
	vhost2.Route("/router").GET(func(c *Context) error { return c.Text(200, "vhost2") })
	vhosts.AddHost(`[a-zA-z0-9]+\.example\.com`, vhost2)

	req := httptest.NewRequest(http.MethodGet, "/router", nil)
	rec := httptest.NewRecorder()
	vhosts.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "default" {
		t.Errorf("Body: expect '%s', got '%s'", "default", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "www.host1.example.com"
	rec = httptest.NewRecorder()
	vhosts.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "vhost1" {
		t.Errorf("Body: expect '%s', got '%s'", "vhost1", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host2.example.com"
	rec = httptest.NewRecorder()
	vhosts.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "vhost2" {
		t.Errorf("Body: expect '%s', got '%s'", "vhost2", s)
	}
}
