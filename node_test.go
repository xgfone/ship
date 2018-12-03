// Copyright 2018 xgfone <xgfone@126.com>.
// Copyright 2016 Dean Karn.
// Copyright 2013 Julien Schmidt.
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file at https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE.

package ship

import (
	"fmt"
	"net/http"
	"testing"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions,
//   those, marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//

func TestAddChain(t *testing.T) {
	r := NewRouter()
	r.Get("/home", defaultHandler)

	testMatchPanic(t, func() { r.Get("/home", defaultHandler) },
		fmt.Errorf("handlers are already registered for path '/home'"))
}

func TestBadWildcard(t *testing.T) {
	r := NewRouter()
	testMatchPanic(t, func() { r.Get("/test/:test*test", defaultHandler) },
		fmt.Errorf("only one wildcard per path segment is allowed, has: ':test*test' in path '/test/:test*test'"))

	r.Get("/users/:id/contact-info/:cid", defaultHandler)
	testMatchPanic(t, func() { r.Get("/users/:id/*", defaultHandler) },
		fmt.Errorf("wildcard route '*' conflicts with existing children in path '/users/:id/*'"))
	testMatchPanic(t, func() { r.Get("/admin/:/", defaultHandler) },
		fmt.Errorf("wildcards must be named with a non-empty name in path '/admin/:/'"))
	testMatchPanic(t, func() { r.Get("/admin/events*", defaultHandler) },
		fmt.Errorf("no / before catch-all in path '/admin/events*'"))

	l2 := NewRouter()
	l2.Get("/", defaultHandler)
	testMatchPanic(t, func() { l2.Get("/*", defaultHandler) },
		fmt.Errorf("catch-all conflicts with existing handle for the path segment root in path '/*'"))

	code, _ := sendTestRequest(http.MethodGet, "/home", l2)
	testEqual(t, code, http.StatusNotFound)

	l3 := NewRouter()
	l3.Get("/testers/:id", defaultHandler)

	code, _ = sendTestRequest(http.MethodGet, "/testers/13/test", l3)
	testEqual(t, code, http.StatusNotFound)
}

func TestDuplicateParams(t *testing.T) {
	r := NewRouter()
	r.Get("/store/:id", defaultHandler)
	testMatchPanic(t, func() { r.Get("/store/:id/employee/:id", defaultHandler) },
		fmt.Errorf("Duplicate param name ':id' detected for route '/store/:id/employee/:id'"))

	r.Get("/company/:id/", defaultHandler)
	testMatchPanic(t, func() { r.Get("/company/:id/employee/:id/", defaultHandler) },
		fmt.Errorf("Duplicate param name ':id' detected for route '/company/:id/employee/:id/'"))
}

func TestWildcardParam(t *testing.T) {
	r := NewRouter()
	r.Get("/users/*", func(ctx Context) (err error) {
		resp := ctx.Response()
		ups := GetURLParam(ctx.Request())
		if _, err = resp.Write([]byte(ups.Get(WildcardParam))); err != nil {
			code := http.StatusInternalServerError
			err = NewHTTPError(code).SetInnerError(err)
		}
		return
	})

	code, body := sendTestRequest(http.MethodGet, "/users/testwild", r)
	testEqual(t, code, http.StatusOK)
	testEqual(t, body, "testwild")

	code, body = sendTestRequest(http.MethodGet, "/users/testwildslash/", r)
	testEqual(t, code, http.StatusOK)
	testEqual(t, body, "testwildslash/")
}

func TestBadRoutes(t *testing.T) {
	r := NewRouter()

	testMatchPanic(t, func() { r.Get("/users//:id", defaultHandler) },
		fmt.Errorf("Bad path '/users//:id' contains duplicate // at index:6"))
}
