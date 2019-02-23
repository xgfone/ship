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

package main

import (
	"net/http"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/renderers/django"
)

var filename = "test_django_engine.html"

func main() {
	engine := django.New(".", ".html")

	router := ship.New()
	router.MuxRenderer().Add(engine.Ext(), ship.HTMLTemplateRenderer(engine))

	// For JSON
	router.Route("/json").GET(func(ctx *ship.Context) error {
		if ctx.QueryParam("pretty") == "1" {
			return ctx.JSONPretty(200, map[string]interface{}{"msg": "json"}, "    ")
			// Or
			// return ctx.Render("jsonpretty", 200, map[string]interface{}{"msg": "json"})
		}
		return ctx.JSON(200, map[string]interface{}{"msg": "json"})
		// Or
		// return ctx.Render("json", 200, map[string]interface{}{"msg": "json"})
	})

	// For XML
	router.Route("/xml").GET(func(ctx *ship.Context) error {
		if ctx.QueryParam("pretty") == "1" {
			return ctx.XMLPretty(200, []string{"msg", "xml"}, "    ")
			// Or
			// return ctx.Render("xmlpretty", 200, []string{"msg", "xml"})
		}
		return ctx.XML(200, []string{"msg", "xml"})
		// Or
		// return ctx.Render("xml", 200, []string{"msg", "xml"})
	})

	// For HTML
	router.Route("/html").GET(func(ctx *ship.Context) error {
		return ctx.Render(filename, 200, map[string]interface{}{"name": "django"})
		// Or
		// return ctx.HTML(200, `<html>...</html>`)
	})

	// For others
	// ...

	http.ListenAndServe(":8080", router)
}
