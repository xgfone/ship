package echo

import (
	"fmt"
	"testing"

	"github.com/xgfone/ship/core"
)

func TestRouter(t *testing.T) {
	router := NewRouter(nil, nil)
	router.Add("static", "/static", []string{"GET"}, func(ctx core.Context) error { return ctx.String(200, "STATIC") })
	router.Add("param", "/test/:name", []string{"POST"}, func(ctx core.Context) error {
		return ctx.String(200, fmt.Sprintf("hello %s", ctx.URLParamByName("name")))
	})

	router.Each(func(name, method, path string) {
		switch name {
		case "static":
			if method != "GET" || path != "/static" {
				t.Fail()
			}
		case "param":
			if method != "POST" || path != "/test/:name" {
				t.Fail()
			}
		}
	})

	if router.URL("param", "Aaron") != "/test/Aaron" {
		t.Fail()
	}

	if router.Find("GET", "/static", nil, nil) == nil {
		t.Fail()
	}

	pnames := make([]string, 1)
	pvalues := make([]string, 1)
	if router.Find("POST", "/test/Aaron", pnames, pvalues) == nil {
		t.Fail()
	}
	if pnames[0] != "name" || pvalues[0] != "Aaron" {
		t.Fail()
	}

	pnames[0] = ""
	pvalues[0] = ""
	router.Add("", "/static/*path", []string{"GET"}, func(ctx core.Context) error { return nil })
	if router.Find("GET", "/static/path/to/file", pnames, pvalues) == nil {
		t.Fail()
	}
	if len(pnames) != 1 || pnames[0] != "path" || pvalues[0] != "path/to/file" {
		t.Fail()
	}
}
