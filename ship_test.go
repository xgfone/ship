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

package ship

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultHandler = func(ctx Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.Request().Method)); err != nil {
		code := http.StatusInternalServerError
		err = NewHTTPError(code).SetInnerError(err)
	}
	return
}

var idHandler = func(ctx Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.Param("id"))); err != nil {
		code := http.StatusInternalServerError
		err = NewHTTPError(code).SetInnerError(err)
	}
	return
}

var params2Handler = func(ctx Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.Param("p1") + "|" + ctx.Param("p2"))); err != nil {
		code := http.StatusInternalServerError
		err = NewHTTPError(code).SetInnerError(err)
	}
	return
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func sendTestRequest(method, path string, s *Ship) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}

	s.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

func TestRouteMap(t *testing.T) {
	m2h := map[string]Handler{"GET": defaultHandler, "POST": defaultHandler}

	s := New()
	s.Route("/path/to").Map(m2h)

	req := httptest.NewRequest(http.MethodGet, "/path/to", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)

	req = httptest.NewRequest(http.MethodPost, "/path/to", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, rec.Code, http.StatusOK)
}

func TestAllMethods(t *testing.T) {
	p := New()
	p.Use(func(next Handler) Handler {
		return func(c Context) error {
			return next(c)
		}
	})

	tests := []struct {
		method  string
		path    string
		url     string
		handler Handler
		code    int
		body    string
		// panicExpected bool
		// panicMsg      string
	}{
		{
			method:  http.MethodGet,
			path:    "/get",
			url:     "/get",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodGet,
		},
		{
			method:  http.MethodPost,
			path:    "/post",
			url:     "/post",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPost,
		},
		{
			method:  http.MethodHead,
			path:    "/head",
			url:     "/head",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodHead,
		},
		{
			method:  http.MethodPut,
			path:    "/put",
			url:     "/put",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPut,
		},
		{
			method:  http.MethodDelete,
			path:    "/delete",
			url:     "/delete",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodDelete,
		},
		{
			method:  http.MethodConnect,
			path:    "/connect",
			url:     "/connect",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodConnect,
		},
		{
			method:  http.MethodOptions,
			path:    "/options",
			url:     "/options",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodOptions,
		},
		{
			method:  http.MethodPatch,
			path:    "/patch",
			url:     "/patch",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodPatch,
		},
		{
			method:  http.MethodTrace,
			path:    "/trace",
			url:     "/trace",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    http.MethodTrace,
		},
		{
			method:  "PROPFIND",
			path:    "/propfind",
			url:     "/propfind",
			handler: defaultHandler,
			code:    http.StatusOK,
			body:    "PROPFIND",
		},
		{
			method:  http.MethodGet,
			path:    "/users/:id",
			url:     "/users/13",
			handler: idHandler,
			code:    http.StatusOK,
			body:    "13",
		},
		{
			method:  http.MethodGet,
			path:    "/2params/:p1",
			url:     "/2params/10",
			handler: params2Handler,
			code:    http.StatusOK,
			body:    "10|",
		},
		{
			method:  http.MethodGet,
			path:    "/2params/:p1/params/:p2",
			url:     "/2params/13/params/12",
			handler: params2Handler,
			code:    http.StatusOK,
			body:    "13|12",
		},
	}

	for _, tt := range tests {
		switch tt.method {
		case http.MethodGet:
			p.Route(tt.path).GET(tt.handler)
		case http.MethodPost:
			p.Route(tt.path).POST(tt.handler)
		case http.MethodHead:
			p.Route(tt.path).HEAD(tt.handler)
		case http.MethodPut:
			p.Route(tt.path).PUT(tt.handler)
		case http.MethodDelete:
			p.Route(tt.path).DELETE(tt.handler)
		case http.MethodConnect:
			p.Route(tt.path).CONNECT(tt.handler)
		case http.MethodOptions:
			p.Route(tt.path).OPTIONS(tt.handler)
		case http.MethodPatch:
			p.Route(tt.path).PATCH(tt.handler)
		case http.MethodTrace:
			p.Route(tt.path).TRACE(tt.handler)
		default:
			p.Route(tt.path).Method(tt.handler, tt.method)
		}
	}

	for _, tt := range tests {
		req, err := http.NewRequest(tt.method, tt.url, nil)
		assert.NoError(t, err)

		res := httptest.NewRecorder()
		p.ServeHTTP(res, req)

		assert.Equal(t, tt.code, res.Code)
		if len(tt.body) > 0 {
			b, err := ioutil.ReadAll(res.Body)
			assert.NoError(t, err)

			s := string(b)
			assert.Equal(t, tt.body, s)
		}
	}

	// test any

	p2 := New()
	p2.Route("/test").Any(defaultHandler)

	test2 := []struct {
		method string
	}{
		{
			method: http.MethodConnect,
		},
		{
			method: http.MethodDelete,
		},
		{
			method: http.MethodGet,
		},
		{
			method: http.MethodHead,
		},
		{
			method: http.MethodOptions,
		},
		{
			method: http.MethodPatch,
		},
		{
			method: http.MethodPost,
		},
		{
			method: http.MethodPut,
		},
		{
			method: http.MethodTrace,
		},
	}

	for _, tt := range test2 {
		req, err := http.NewRequest(tt.method, "/test", nil)
		assert.NoError(t, err)

		res := httptest.NewRecorder()
		p2.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		b, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		s := string(b)
		assert.Equal(t, tt.method, s)
	}
}

func TestRouterAPI(t *testing.T) {
	p := New()

	for _, route := range githubAPI {
		p.Route(route.path).Method(func(ctx Context) error {
			if _, err := ctx.Response().Write([]byte(ctx.Request().URL.Path)); err != nil {
				panic(err)
			}
			return nil
		}, route.method)
	}

	for _, route := range githubAPI {
		code, body := sendTestRequest(route.method, route.path, p)
		assert.Equal(t, body, route.path)
		assert.Equal(t, code, http.StatusOK)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	p := New(Config{
		KeepTrailingSlashPath:   true,
		MethodNotAllowedHandler: MethodNotAllowedHandler(),
	})

	p.Route("/home/").PUT(defaultHandler)
	p.Route("/home/").POST(defaultHandler)
	p.Route("/home/").HEAD(defaultHandler)
	p.Route("/home/").DELETE(defaultHandler)
	p.Route("/home/").CONNECT(defaultHandler)
	p.Route("/home/").OPTIONS(defaultHandler)
	p.Route("/home/").PATCH(defaultHandler)
	p.Route("/home/").TRACE(defaultHandler)
	p.Route("/home/").Method(defaultHandler, "PROPFIND")

	code, _ := sendTestRequest(http.MethodPut, "/home/", p)
	assert.Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodGet, "/home/", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)

	assert.Equal(t, w.Code, http.StatusMethodNotAllowed)

	sallow, ok := w.Header()[HeaderAllow]
	if len(sallow) == 0 {
		t.Fail()
	} else {
		allow := strings.Split(sallow[0], ", ")
		assert.Equal(t, ok, true)
		assert.Equal(t, len(allow), 9)
	}

	r, _ = http.NewRequest("PROPFIND2", "/home/1", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, r)

	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestMethodNotAllowed2(t *testing.T) {
	p := New(Config{
		KeepTrailingSlashPath:   true,
		MethodNotAllowedHandler: MethodNotAllowedHandler(),
	})

	p.Route("/home/").GET(defaultHandler)
	p.Route("/home/").HEAD(defaultHandler)

	code, _ := sendTestRequest(http.MethodGet, "/home/", p)
	assert.Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodPost, "/home/", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)

	assert.Equal(t, w.Code, http.StatusMethodNotAllowed)

	sallow, ok := w.Header()[HeaderAllow]
	if len(sallow) == 0 {
		t.Fail()
	}
	allow := strings.Split(sallow[0], ", ")

	// Sometimes this array is out of order for whatever reason?
	if allow[0] == http.MethodGet {
		assert.Equal(t, ok, true)
		assert.Equal(t, allow[0], http.MethodGet)
		assert.Equal(t, allow[1], http.MethodHead)
	} else {
		assert.Equal(t, ok, true)
		assert.Equal(t, allow[1], http.MethodGet)
		assert.Equal(t, allow[0], http.MethodHead)
	}
}

func TestAutomaticallyHandleOPTIONS(t *testing.T) {
	p := New(Config{
		OptionsHandler:          OptionsHandler(),
		MethodNotAllowedHandler: MethodNotAllowedHandler(),
	})

	p.Route("/home").GET(defaultHandler)
	p.Route("/home").POST(defaultHandler)
	p.Route("/home").DELETE(defaultHandler)
	p.Route("/home").HEAD(defaultHandler)
	p.Route("/home").PUT(defaultHandler)
	p.Route("/home").CONNECT(defaultHandler)
	p.Route("/home").PATCH(defaultHandler)
	p.Route("/home").TRACE(defaultHandler)
	p.Route("/home").Method(defaultHandler, "PROPFIND")

	code, _ := sendTestRequest(http.MethodGet, "/home", p)
	assert.Equal(t, code, http.StatusOK)

	r, _ := http.NewRequest(http.MethodOptions, "/home", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)

	assert.Equal(t, w.Code, http.StatusOK)
	allow, ok := w.Header()[HeaderAllow]
	assert.Equal(t, ok, true)
	assert.Equal(t, len(strings.Split(allow[0], ", ")), 9)
}

func TestNotFound(t *testing.T) {
	notFound := func(ctx Context) error {
		http.Error(ctx.Response(), http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}

	p := New(Config{NotFoundHandler: notFound})
	p.Route("/home/").GET(defaultHandler)
	p.Route("/home/").POST(defaultHandler)
	p.Route("/users/:id").GET(defaultHandler)
	p.Route("/users/:id/:id2/:id3").GET(defaultHandler)

	code, _ := sendTestRequest("BAD_METHOD", "/home/", p)
	assert.Equal(t, code, http.StatusNotFound)

	code, _ = sendTestRequest(http.MethodGet, "/users/14/more", p)
	assert.Equal(t, code, http.StatusNotFound)
}

func TestBasePath(t *testing.T) {
	p := New()
	p.Route("/").GET(defaultHandler)

	code, _ := sendTestRequest(http.MethodGet, "/", p)
	assert.Equal(t, code, http.StatusOK)
}

type route struct {
	method string
	path   string
}

var githubAPI = []route{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	//{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	//{"GET", "/gists/public"},
	//{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	//{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	//{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	//{"GET", "/repos/:owner/:repo/issues/events"},
	//{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	//{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	//{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	//{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	//{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	//{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	//{"GET", "/repos/:owner/:repo/contents/*path"},
	//{"PUT", "/repos/:owner/:repo/contents/*path"},
	//{"DELETE", "/repos/:owner/:repo/contents/*path"},
	//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	//{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	//{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	//{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	//{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}

type TestType struct{}

func (t TestType) Create(ctx Context) error { return nil }
func (t TestType) Delete(ctx Context) error { return nil }
func (t TestType) Update(ctx Context) error { return nil }
func (t TestType) Get(ctx Context) error    { return nil }
func (t TestType) Has(ctx Context) error    { return nil }
func (t TestType) NotHandler()              {}

func strIsInSlice(s string, ss []string) bool {
	for _, _s := range ss {
		if _s == s {
			return true
		}
	}
	return false
}

func TestRouteMapType(t *testing.T) {
	router1 := New()
	router1.Route("/v1").MapType(TestType{})
	router1.Traverse(func(name, method, path string) {
		switch method {
		case "GET":
			if name != "testtype_get" || path != "/v1/testtype/get" {
				t.Fail()
			}
		case "POST":
			if name != "testtype_create" || path != "/v1/testtype/create" {
				t.Fail()
			}
		case "PUT":
			if name != "testtype_update" || path != "/v1/testtype/update" {
				t.Fail()
			}
		case "DELETE":
			if name != "testtype_delete" || path != "/v1/testtype/delete" {
				t.Fail()
			}
		default:
			t.Fail()
		}
	})

	router2 := New()
	router2.Route("").MapType(TestType{})
	router2.Traverse(func(name, method, path string) {
		switch method {
		case "GET":
			if name != "testtype_get" || path != "/testtype/get" {
				t.Fail()
			}
		case "POST":
			if name != "testtype_create" || path != "/testtype/create" {
				t.Fail()
			}
		case "PUT":
			if name != "testtype_update" || path != "/testtype/update" {
				t.Fail()
			}
		case "DELETE":
			if name != "testtype_delete" || path != "/testtype/delete" {
				t.Fail()
			}
		default:
			t.Fail()
		}
	})
}

func TestShipVHost(t *testing.T) {
	s := New()
	s.Route("/router").GET(func(c Context) error { return c.String(200, "default") })

	vhost1 := s.VHost("host1.example.com")
	vhost1.Route("/router").GET(func(c Context) error { return c.String(200, "vhost1") })

	vhost2 := s.VHost("host2.example.com")
	vhost2.Route("/router").GET(func(c Context) error { return c.String(200, "vhost2") })

	req := httptest.NewRequest(http.MethodGet, "/router", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "default", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host1.example.com"
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "vhost1", rec.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host2.example.com"
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "vhost2", rec.Body.String())
}

func TestRouteStaticFile(t *testing.T) {
	s := New()
	s.Route("/README.md").StaticFile("./README.md")

	req := httptest.NewRequest(http.MethodHead, "/README.md", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 0, rec.Body.Len())
	assert.NotZero(t, rec.Header().Get(HeaderEtag))
	assert.NotZero(t, rec.Header().Get(HeaderContentLength))

	req = httptest.NewRequest(http.MethodGet, "/README.md", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.NotZero(t, rec.Body.Len())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/markdown; charset=utf-8", rec.Header().Get(HeaderContentType))
}

func TestRouteStaticFS(t *testing.T) {
	s := New()
	s.Route("/ship").StaticFS(http.Dir("."))

	req := httptest.NewRequest(http.MethodHead, "/ship/", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.NotZero(t, rec.Body.Len())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get(HeaderContentType))
	assert.Contains(t, rec.Body.String(), `"README.md"`)

	req = httptest.NewRequest(http.MethodGet, "/ship/README.md", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotZero(t, rec.Body.Len())
	assert.NotZero(t, rec.Header().Get(HeaderContentLength))

	req = httptest.NewRequest(http.MethodHead, "/ship/core/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.NotZero(t, rec.Body.Len())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get(HeaderContentType))
	assert.Contains(t, rec.Body.String(), `"router.go"`)

	req = httptest.NewRequest(http.MethodGet, "/ship/core/router.go", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotZero(t, rec.Body.Len())
	assert.NotZero(t, rec.Header().Get(HeaderContentLength))
}

func TestRouteStatic(t *testing.T) {
	s := New()
	s.Route("/ship").Static(".")

	req := httptest.NewRequest(http.MethodHead, "/ship/", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.NotZero(t, rec.Body.Len())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get(HeaderContentType))
	assert.NotContains(t, rec.Body.String(), `"README.md"`)

	req = httptest.NewRequest(http.MethodGet, "/ship/README.md", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotZero(t, rec.Body.Len())
	assert.NotZero(t, rec.Header().Get(HeaderContentLength))

	req = httptest.NewRequest(http.MethodHead, "/ship/core/", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.NotZero(t, rec.Body.Len())
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get(HeaderContentType))
	assert.NotContains(t, rec.Body.String(), `"router.go"`)

	req = httptest.NewRequest(http.MethodGet, "/ship/core/router.go", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotZero(t, rec.Body.Len())
	assert.NotZero(t, rec.Header().Get(HeaderContentLength))
}

func TestRouteMatcher(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	s := New(Config{Logger: NewNoLevelLogger(buf)})

	s.Route("/path1").Header("Content-Type", "application/json").GET(
		func(ctx Context) error { return ctx.String(200, "OK") })
	req := httptest.NewRequest(http.MethodGet, "/path1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	s.Route("/path3").Header("Content-Type").GET(
		func(ctx Context) error { return ctx.String(200, "OK") })
	req = httptest.NewRequest(http.MethodGet, "/path3", nil)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	s.Route("/path4").Header("Content-Type").GET(
		func(ctx Context) error { return ctx.String(200, "OK") })
	req = httptest.NewRequest(http.MethodGet, "/path4", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "code=404, msg=Not Found, err=missing the header 'Content-Type'", rec.Body.String())
}

func TestContextBindQuery(t *testing.T) {
	type V struct {
		A string `query:"a"`
		B int    `query:"b"`
	}
	vs := V{}

	s := New()
	s.Route("/path").GET(func(ctx Context) error { return ctx.BindQuery(&vs) })
	req := httptest.NewRequest(http.MethodGet, "/path?a=xyz&b=2", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "xyz", vs.A)
	assert.Equal(t, 2, vs.B)
}

func TestContextAccept(t *testing.T) {
	expected := []string{"text/html", "application/xhtml+xml", "image/webp", "application/", ""}
	var accepts []string
	s := New()
	s.R("/path").GET(func(ctx Context) error {
		accepts = ctx.Accept()
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set(HeaderAccept, "text/html, application/xhtml+xml, application/*;q=0.9, image/webp, */*;q=0.8")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expected, accepts)
}

type safeBufferWriter struct {
	buf  *bytes.Buffer
	lock *sync.Mutex
}

func (bw *safeBufferWriter) Write(p []byte) (int, error) {
	bw.lock.Lock()
	n, err := bw.buf.Write(p)
	bw.lock.Unlock()
	return n, err
}

func TestShipLink(t *testing.T) {
	buf := &safeBufferWriter{buf: bytes.NewBuffer(nil), lock: new(sync.Mutex)}
	logger := NewNoLevelLogger(buf, 0)

	prouter := New(Config{Name: "parent", Logger: logger})
	crouter1 := New(Config{Name: "child1", Logger: logger}).Link(prouter)
	crouter2 := prouter.Clone("child2").Link(prouter)

	prouter.RegisterOnShutdown(func() { time.Sleep(time.Millisecond) })
	crouter1.RegisterOnShutdown(func() { time.Sleep(time.Millisecond) })
	crouter2.RegisterOnShutdown(func() { time.Sleep(time.Millisecond) })

	go func() {
		time.Sleep(time.Millisecond * 100)
		prouter.Shutdown(context.Background())
	}()
	go crouter1.Start("127.0.0.1:11111")
	go crouter2.Start("127.0.0.1:11112")
	prouter.Start("127.0.0.1:11113")

	prouter.Wait()
	time.Sleep(time.Millisecond * 100)
	buf.lock.Lock()
	lines := strings.Split(strings.TrimSpace(buf.buf.String()), "\n")
	buf.lock.Unlock()

	assert.Equal(t, 6, len(lines))
	if len(lines) != 6 {
		return
	}
	sort.Strings(lines[:3])
	sort.Strings(lines[3:])

	assert.Equal(t, "[I] The HTTP Server [child1] is running on 127.0.0.1:11111", lines[0])
	assert.Equal(t, "[I] The HTTP Server [child2] is running on 127.0.0.1:11112", lines[1])
	assert.Equal(t, "[I] The HTTP Server [parent] is running on 127.0.0.1:11113", lines[2])
	assert.Equal(t, "[I] The HTTP Server [child1] is shutdown", lines[3])
	assert.Equal(t, "[I] The HTTP Server [child2] is shutdown", lines[4])
	assert.Equal(t, "[I] The HTTP Server [parent] is shutdown", lines[5])
}
