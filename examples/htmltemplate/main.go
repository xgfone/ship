package main

import (
	"net/http"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/render"
	"github.com/xgfone/ship/render/django"
)

var filename = "test_django_engine.html"

func main() {
	tm := render.NewHTMLTemplateManager()
	tm.Register(django.New(".", ".html"))

	router := ship.New()
	router.MuxRender().Add(".html", tm)

	// For JSON
	router.Route("/json").GET(func(ctx ship.Context) error {
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
	router.Route("/xml").GET(func(ctx ship.Context) error {
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
	router.Route("/html").GET(func(ctx ship.Context) error {
		return ctx.Render(filename, 200, map[string]interface{}{"name": "django"})
		// Or
		// return ctx.HTML(200, `<html>...</html>`)
	})

	// For others
	// ...

	http.ListenAndServe(":8080", router)
}
