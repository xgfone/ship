// Copyright 2021 xgfone
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

var noHandler = NothingHandler()

func BenchmarkShipWithoutVHost(b *testing.B) {
	router := New()
	router.Route("/path1").GET(noHandler)
	router.Route("/path2").GET(noHandler)

	req, err := http.NewRequest(http.MethodGet, "http://www.example.com/path2", nil)
	rec := httptest.NewRecorder()
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rec, req)
	}
}

func BenchmarkShipWithExactVHost(b *testing.B) {
	vhost := New()
	vhost.Route("/path1").GET(noHandler)
	vhost.Route("/path2").GET(noHandler)

	vhosts := NewHostManagerHandler(nil)
	vhosts.AddHost("www1.example.com", vhost)
	vhosts.AddHost("www2.example.com", New())

	req, err := http.NewRequest(http.MethodGet, "http://www1.example.com/path2", nil)
	rec := httptest.NewRecorder()
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vhosts.ServeHTTP(rec, req)
	}
}

func BenchmarkShipWithPrefixVHost(b *testing.B) {
	vhost := New()
	vhost.Route("/path1").GET(noHandler)
	vhost.Route("/path2").GET(noHandler)

	vhosts := NewHostManagerHandler(nil)
	vhosts.AddHost("*.example.com", vhost)

	req, err := http.NewRequest(http.MethodGet, "http://www.example.com/path2", nil)
	rec := httptest.NewRecorder()
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vhosts.ServeHTTP(rec, req)
	}
}

func BenchmarkShipWithRegexpVHost(b *testing.B) {
	vhost := New()
	vhost.Route("/path1").GET(noHandler)
	vhost.Route("/path2").GET(noHandler)

	vhosts := NewHostManagerHandler(nil)
	vhosts.AddHost(`[a-zA-z0-9]+\.example\.com`, vhost)

	req, err := http.NewRequest(http.MethodGet, "http://www.example.com/path2", nil)
	rec := httptest.NewRecorder()
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vhosts.ServeHTTP(rec, req)
	}
}
