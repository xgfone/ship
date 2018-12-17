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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	if _, err = resp.Write([]byte(ctx.URLParamByName("id"))); err != nil {
		code := http.StatusInternalServerError
		err = NewHTTPError(code).SetInnerError(err)
	}
	return
}

var params2Handler = func(ctx Context) (err error) {
	resp := ctx.Response()
	get := ctx.URLParamByName
	if _, err = resp.Write([]byte(get("p1") + "|" + get("p2"))); err != nil {
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
			p.Route(tt.path, tt.handler).GET()
		case http.MethodPost:
			p.Route(tt.path, tt.handler).POST()
		case http.MethodHead:
			p.Route(tt.path, tt.handler).HEAD()
		case http.MethodPut:
			p.Route(tt.path, tt.handler).PUT()
		case http.MethodDelete:
			p.Route(tt.path, tt.handler).DELETE()
		case http.MethodConnect:
			p.Route(tt.path, tt.handler).CONNECT()
		case http.MethodOptions:
			p.Route(tt.path, tt.handler).OPTIONS()
		case http.MethodPatch:
			p.Route(tt.path, tt.handler).PATCH()
		case http.MethodTrace:
			p.Route(tt.path, tt.handler).TRACE()
		default:
			p.Route(tt.path, tt.handler).Method(tt.method)
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
	p2.Route("/test", defaultHandler).Any()

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
		p.Route(route.path, func(ctx Context) error {
			if _, err := ctx.Response().Write([]byte(ctx.Request().URL.Path)); err != nil {
				panic(err)
			}
			return nil
		}).Method(route.method)
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

	p.Route("/home/", defaultHandler).PUT()
	p.Route("/home/", defaultHandler).POST()
	p.Route("/home/", defaultHandler).HEAD()
	p.Route("/home/", defaultHandler).DELETE()
	p.Route("/home/", defaultHandler).CONNECT()
	p.Route("/home/", defaultHandler).OPTIONS()
	p.Route("/home/", defaultHandler).PATCH()
	p.Route("/home/", defaultHandler).TRACE()
	p.Route("/home/", defaultHandler).Method("PROPFIND")

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

	p.Route("/home/", defaultHandler).GET()
	p.Route("/home/", defaultHandler).HEAD()

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

	p.Route("/home", defaultHandler).GET()
	p.Route("/home", defaultHandler).POST()
	p.Route("/home", defaultHandler).DELETE()
	p.Route("/home", defaultHandler).HEAD()
	p.Route("/home", defaultHandler).PUT()
	p.Route("/home", defaultHandler).CONNECT()
	p.Route("/home", defaultHandler).PATCH()
	p.Route("/home", defaultHandler).TRACE()
	p.Route("/home", defaultHandler).Method("PROPFIND")

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
	p.Route("/home/", defaultHandler).GET()
	p.Route("/home/", defaultHandler).POST()
	p.Route("/users/:id", defaultHandler).GET()
	p.Route("/users/:id/:id2/:id3", defaultHandler).GET()

	code, _ := sendTestRequest("BAD_METHOD", "/home/", p)
	assert.Equal(t, code, http.StatusNotFound)

	code, _ = sendTestRequest(http.MethodGet, "/users/14/more", p)
	assert.Equal(t, code, http.StatusNotFound)
}

func TestBasePath(t *testing.T) {
	p := New()
	p.Route("/", defaultHandler).GET()

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
