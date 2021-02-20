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
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/xgfone/ship/v3/router"
	"github.com/xgfone/ship/v3/router/echo"
)

func sortRouteInfos(ris []RouteInfo) {
	sort.Slice(ris, func(i, j int) bool {
		return ris[i].CtxData.(int) < ris[j].CtxData.(int)
	})
}

func TestRoute(t *testing.T) {
	s := New()
	handler := OkHandler()
	routes := []RouteInfo{
		{Host: "", Name: "name", Path: "/path", Method: http.MethodGet, Handler: handler},
		{Host: "host1", Name: "name1", Path: "/path1", Method: http.MethodGet, Handler: handler},
		{Host: "host1", Name: "name2", Path: "/path2", Method: http.MethodGet, Handler: handler},
		{Host: "host1", Name: "name3", Path: "/path3", Method: http.MethodGet, Handler: handler},
		{Host: "host2", Name: "name4", Path: "/path4", Method: http.MethodGet, Handler: handler},
		{Host: "host2", Name: "name5", Path: "/path5", Method: http.MethodGet, Handler: handler},
		{Host: "host2", Name: "name6", Path: "/path6", Method: http.MethodGet, Handler: handler},
	}

	for i, r := range routes {
		s.Route(r.Path).Name(r.Name).Host(r.Host).CtxData(i).Method(r.Handler, r.Method)
	}

	if rs := s.Routes(); len(rs) != 7 {
		t.Errorf("the number of the registered routes is %d, not 7\n", len(rs))
	} else {
		sortRouteInfos(rs)
		for i, r := range rs {
			if i != r.CtxData.(int) {
				t.Errorf("%d: %+v", i, r)
			}

			switch r.Name {
			case "name":
			case "name1":
			case "name2":
			case "name3":
			case "name4":
			case "name5":
			case "name6":
				route := routes[6]
				if r.Name != route.Name || r.Host != route.Host ||
					r.Path != route.Path || r.Method != route.Method {
					t.Errorf("expected %v, got %v\n", route, r)
				}
			default:
				t.Errorf("unknown route: %v", r)
			}
		}
	}

	hosts := make([]string, 0, 3)
	for host := range s.Routers() {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	if len(hosts) != 3 {
		t.Errorf("the number of routers is not 3: %v", hosts)
	} else {
		_hosts := []string{"", "host1", "host2"}
		for i := range hosts {
			if hosts[i] != _hosts[i] {
				t.Errorf("%dth: expected '%s', got '%s'", i, _hosts[i], hosts[i])
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////

var defaultHandler = func(ctx *Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.Request().Method)); err != nil {
		code := http.StatusInternalServerError
		err = HTTPError{Code: code, Err: err}
	}
	return
}

var idHandler = func(ctx *Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.URLParam("id"))); err != nil {
		code := http.StatusInternalServerError
		err = HTTPError{Code: code, Err: err}
	}
	return
}

var params2Handler = func(ctx *Context) (err error) {
	resp := ctx.Response()
	_, err = resp.Write([]byte(ctx.URLParam("p1") + "|" + ctx.URLParam("p2")))
	if err != nil {
		code := http.StatusInternalServerError
		err = HTTPError{Code: code, Err: err}
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
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/path/to", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAllMethods(t *testing.T) {
	p := New()
	p.Use(func(next Handler) Handler {
		return func(c *Context) error {
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
		if err != nil {
			t.Error(err)
		}

		res := httptest.NewRecorder()
		p.ServeHTTP(res, req)

		if tt.code != res.Code {
			t.Errorf("StatusCode: expect %d, got %d", tt.code, res.Code)
		}

		if len(tt.body) > 0 {
			if b, err := ioutil.ReadAll(res.Body); err != nil {
				t.Error(err)
			} else if s := string(b); s != tt.body {
				t.Errorf("Body: expect '%s', got '%s'", tt.body, s)
			}
		}
	}

	// test any

	p2 := New()
	p2.Route("/test").Any(defaultHandler)

	test2 := []struct{ method string }{
		{method: http.MethodConnect},
		{method: http.MethodDelete},
		{method: http.MethodGet},
		{method: http.MethodHead},
		{method: http.MethodOptions},
		{method: http.MethodPatch},
		{method: http.MethodPost},
		{method: http.MethodPut},
		{method: http.MethodTrace},
	}

	for _, tt := range test2 {
		req, err := http.NewRequest(tt.method, "/test", nil)
		if err != nil {
			t.Error(err)
		}

		res := httptest.NewRecorder()
		p2.ServeHTTP(res, req)

		if http.StatusOK != res.Code {
			t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, res.Code)
		}

		if b, err := ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		} else if s := string(b); s != tt.method {
			t.Errorf("Body: expect '%s', got '%s'", tt.method, s)
		}
	}
}

func TestRouterAPI(t *testing.T) {
	p := New()

	for _, route := range githubAPI {
		p.Route(route.path).Method(func(ctx *Context) error {
			if _, err := ctx.Response().Write([]byte(ctx.Request().URL.Path)); err != nil {
				panic(err)
			}
			return nil
		}, route.method)
	}

	for _, route := range githubAPI {
		code, body := sendTestRequest(route.method, route.path, p)
		if body != route.path {
			t.Errorf("Body: expect '%s', got '%s'", route.path, body)
		}
		if code != http.StatusOK {
			t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, code)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	p := New().SetNewRouter(func() router.Router {
		return echo.NewRouter(nil, RouteInfo{Handler: MethodNotAllowedHandler()})
	})

	p.Route("/home").PUT(defaultHandler)
	p.Route("/home").POST(defaultHandler)
	p.Route("/home").HEAD(defaultHandler)
	p.Route("/home").DELETE(defaultHandler)
	p.Route("/home").CONNECT(defaultHandler)
	p.Route("/home").OPTIONS(defaultHandler)
	p.Route("/home").PATCH(defaultHandler)
	p.Route("/home").TRACE(defaultHandler)
	p.Route("/home").Method(defaultHandler, "PROPFIND")

	code, _ := sendTestRequest(http.MethodPut, "/home", p)
	if code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, code)
	}

	r, _ := http.NewRequest(http.MethodGet, "/home", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	r, _ = http.NewRequest("PROPFIND2", "/home/1", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestMethodNotAllowed2(t *testing.T) {
	p := New().SetNewRouter(func() router.Router {
		return echo.NewRouter(nil, RouteInfo{Handler: MethodNotAllowedHandler()})
	})

	p.Route("/home").GET(defaultHandler)
	p.Route("/home").HEAD(defaultHandler)

	code, _ := sendTestRequest(http.MethodGet, "/home", p)
	if code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, code)
	}

	r, _ := http.NewRequest(http.MethodPost, "/home", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

}

func TestNotFound(t *testing.T) {
	notFound := func(ctx *Context) error {
		http.Error(ctx.Response(), http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return nil
	}

	p := New()
	p.NotFound = notFound
	p.Route("/home/").GET(defaultHandler)
	p.Route("/home/").POST(defaultHandler)
	p.Route("/users/:id").GET(defaultHandler)
	p.Route("/users/:id/:id2/:id3").GET(defaultHandler)

	code, _ := sendTestRequest("BAD_METHOD", "/home/", p)
	if code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, code)
	}

	code, _ = sendTestRequest(http.MethodGet, "/users/14/more", p)
	if code != http.StatusNotFound {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusNotFound, code)
	}
}

func TestBasePath(t *testing.T) {
	p := New()
	p.Route("/").GET(defaultHandler)

	code, _ := sendTestRequest(http.MethodGet, "/", p)
	if code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, code)
	}
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

func (t TestType) Create(ctx *Context) error { return nil }
func (t TestType) Delete(ctx *Context) error { return nil }
func (t TestType) Update(ctx *Context) error { return nil }
func (t TestType) Get(ctx *Context) error    { return nil }
func (t TestType) Has(ctx *Context) error    { return nil }
func (t TestType) NotHandler()               {}

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
	for _, r := range router1.Routes() {
		name, method, path := r.Name, r.Method, r.Path
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
	}

	router2 := New()
	router2.Route("/").MapType(TestType{})
	for _, r := range router2.Routes() {
		name, method, path := r.Name, r.Method, r.Path
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
	}
}

func TestShipHost(t *testing.T) {
	s := New()
	s.Route("/router").GET(func(c *Context) error { return c.Text(200, "default") })
	s.Route("/router").Host("*.host1.example.com").
		GET(func(c *Context) error { return c.Text(200, "vhost1") })
	s.Route("/router").Host(`[a-zA-z0-9]+\.example\.com`).
		GET(func(c *Context) error { return c.Text(200, "vhost2") })

	req := httptest.NewRequest(http.MethodGet, "/router", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "default" {
		t.Errorf("Body: expect '%s', got '%s'", "default", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "www.host1.example.com"
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "vhost1" {
		t.Errorf("Body: expect '%s', got '%s'", "vhost1", s)
	}

	req = httptest.NewRequest(http.MethodGet, "/router", nil)
	req.Host = "host2.example.com"
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	if s := rec.Body.String(); s != "vhost2" {
		t.Errorf("Body: expect '%s', got '%s'", "vhost2", s)
	}
}

func TestRouteStaticFile(t *testing.T) {
	s := New()
	s.Route("/README.md").StaticFile("./README.md")

	req := httptest.NewRequest(http.MethodHead, "/README.md", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if rec.Body.Len() != 0 {
		t.Error("the body is not empty")
	} else if rec.Header().Get(HeaderEtag) == "" {
		t.Error("no ETAG")
	}

	req = httptest.NewRequest(http.MethodGet, "/README.md", nil)
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if rec.Body.Len() == 0 {
		t.Error("the body is empty")
	} else if ct := rec.Header().Get(HeaderContentType); ct != "text/markdown; charset=utf-8" {
		t.Errorf("ContentType: expect '%s', got '%s'", "text/markdown; charset=utf-8", ct)
	}
}

func TestRouteHasHeader(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	s := New().SetLogger(NewLoggerFromWriter(buf, ""))

	s.Route("/path").HasHeader("Content-Type", "application/json").GET(
		func(ctx *Context) error { return ctx.Text(200, "OK") })

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set("Content-Type", "application/xml")
	rec = httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestContextBindQuery(t *testing.T) {
	type V struct {
		A string `query:"a" default:"xyz"`
		B int    `query:"b"`
	}
	vs := V{}

	s := Default()
	s.Route("/path").GET(func(ctx *Context) error { return ctx.BindQuery(&vs) })
	req := httptest.NewRequest(http.MethodGet, "/path?b=2", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	} else if vs.A != "xyz" {
		t.Errorf("expect '%s', got '%s'", "xyz", vs.A)
	} else if vs.B != 2 {
		t.Errorf("expect %d, got %d", 2, vs.B)
	}
}

func TestContextAccept(t *testing.T) {
	expected := []string{"text/html", "application/xhtml+xml", "image/webp", "application/", ""}
	var accepts []string
	s := New()
	s.R("/path").GET(func(ctx *Context) error {
		accepts = ctx.Accept()
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.Header.Set(HeaderAccept, "text/html, application/xhtml+xml, application/*;q=0.9, image/webp, */*;q=0.8")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode: expect %d, got %d", http.StatusOK, rec.Code)
	}
	for i := range expected {
		if expected[i] != accepts[i] {
			t.Errorf("expect '%s', got '%s'", expected[i], accepts[i])
		}
	}
}

func TestSetRouteFilter(t *testing.T) {
	app := New()
	app.RouteFilter = func(ri RouteInfo) bool {
		if ri.Name == "" {
			return true
		} else if !strings.HasPrefix(ri.Path, "/group/") {
			return true
		}
		return false
	}

	handler := func(ctx *Context) error { return nil }
	app.Group("/group").R("/name").Name("test").GET(handler)
	app.R("/noname").GET(handler)

	routes := app.Routes()
	if len(routes) == 0 {
		t.Error("no routes")
	}
	for _, ri := range routes {
		if ri.Name != "test" {
			t.Error(ri)
		}
	}
}

func TestSetRouteModifier(t *testing.T) {
	app := New()
	app.RouteModifier = func(ri RouteInfo) RouteInfo {
		if !strings.HasPrefix(ri.Path, "/prefix/") {
			ri.Path = "/prefix" + ri.Path
		}
		return ri
	}

	handler := func(ctx *Context) error { return nil }
	app.R("/path").GET(handler)

	noRoute := true
	for _, ri := range app.Routes() {
		noRoute = false
		if ri.Path != "/prefix/path" {
			t.Error(ri.Path)
		}
	}

	if noRoute {
		t.Fail()
	}
}

const middlewareoutput = `
pre m1 start
pre m2 start
use m1 start
use m2 start
group m1 start
group m2 start
route m1 start
route m2 start
route m2 end
route m1 end
group m2 end
group m1 end
use m2 end
use m1 end
pre m2 end
pre m1 end
`

func TestMiddleware(t *testing.T) {
	bs := bytes.NewBufferString("\n")
	s := New()
	s.Pre(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("pre m1 start\n")
			err := next(ctx)
			bs.WriteString("pre m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("pre m2 start\n")
			err := next(ctx)
			bs.WriteString("pre m2 end\n")
			return err
		}
	})

	s.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("use m1 start\n")
			err := next(ctx)
			bs.WriteString("use m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("use m2 start\n")
			err := next(ctx)
			bs.WriteString("use m2 end\n")
			return err
		}
	})

	group := s.Group("/v1")
	group.Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("group m1 start\n")
			err := next(ctx)
			bs.WriteString("group m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("group m2 start\n")
			err := next(ctx)
			bs.WriteString("group m2 end\n")
			return err
		}
	})

	group.R("/route").Use(func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("route m1 start\n")
			err := next(ctx)
			bs.WriteString("route m1 end\n")
			return err
		}
	}, func(next Handler) Handler {
		return func(ctx *Context) error {
			bs.WriteString("route m2 start\n")
			err := next(ctx)
			bs.WriteString("route m2 end\n")
			return err
		}
	}).GET(OkHandler())

	req := httptest.NewRequest(http.MethodGet, "/v1/route", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if bs.String() != middlewareoutput {
		t.Error(bs.String())
		t.Fail()
	}
}

func TestShipAddRoutes(t *testing.T) {
	app := New()
	app.RouteModifier = func(ri RouteInfo) RouteInfo {
		ri.Host = "www.example.com"
		return ri
	}
	app.Group("/test").AddRoutes(HTTPPprofToRouteInfo()...)

	routes := app.Routes()
	if len(routes) == 0 {
		t.Errorf("no routes")
	}

	for _, ri := range routes {
		switch ri.Path {
		case "/test/debug/pprof/*":
		case "/test/debug/pprof/cmdline":
		case "/test/debug/pprof/profile":
		case "/test/debug/pprof/symbol":
			switch ri.Method {
			case http.MethodGet, http.MethodPost:
			default:
				t.Error(ri)
			}
		case "/test/debug/pprof/trace":
		default:
			t.Error(ri)
		}
	}

	rs := HTTPPprofToRouteInfo()
	for i := range rs {
		rs[i].Handler = nil
	}
	app.Group("/test").DelRoutes(rs...)
	if routes := app.Routes(); len(routes) > 0 {
		t.Error(routes)
	}
}

func TestRoute_RemoveAny(t *testing.T) {
	h := OkHandler()
	app := New()
	app.Route("/path1").GET(h).POST(h).DELETE(h)
	if routes := app.Routes(); len(routes) != 3 {
		t.Error(routes)
	}

	app.Route("/path1").RemoveAny()
	if routes := app.Routes(); len(routes) != 0 {
		t.Error(routes)
	}
}
