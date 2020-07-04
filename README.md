# ship [![Build Status](https://travis-ci.org/xgfone/ship.svg?branch=master)](https://travis-ci.org/xgfone/ship) [![GoDoc](https://godoc.org/github.com/xgfone/ship?status.svg)](https://pkg.go.dev/github.com/xgfone/ship/v2) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a flexible, powerful, high performance and minimalist Go Web HTTP router framework. It is inspired by [echo](https://github.com/labstack/echo) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.

`ship` has been stable, and the current version is `v2` and support Go `1.11+`.


## Install

```shell
go get -u github.com/xgfone/ship/v2
```


## Quick Start

```go
// example.go
package main

import "github.com/xgfone/ship/v2"

func main() {
	router := ship.New()
	router.Route("/ping").GET(func(ctx *ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	// Start the HTTP server.
	router.Start(":8080").Wait()
	// or
	// http.ListenAndServe(":8080", router)
}
```

```shell
$ go run example.go
```

```shell
$ curl http://127.0.0.1:8080/ping
{"message":"pong"}
```


## API Example

### `Router`
#### Using `Connect`, `Get`, `Post`, `Put`, `Patch`, `Delete` and `Option`

```go
func main() {
    router := ship.New()
    router.Route("/path/get").GET(getHandler)
    router.Route("/path/put").PUT(putHandler)
    router.Route("/path/post").POST(postHandler)
    router.Route("/path/patch").PATCH(patchHandler)
    router.Route("/path/delete").DELETE(deleteHandler)
    router.Route("/path/option").OPTIONS(optionHandler)
    router.Route("/path/connect").CONNECT(connectHandler)
    router.Start(":8080").Wait()
}
```

Notice: you can register the same handler with more than one method by `Route(path string).Method(handler Handler, method ...string)`.

`R` is the alias of `Route`, and you can register the routes by `R(path string).Method(handler Handler, method ...string)`.

#### Cascade the registered routes

```go
func main() {
    router := ship.New()
    router.R("/path/to").GET(getHandler).POST(postHandler).DELETE(deleteHandler)
    router.Start(":8080").Wait()
}
```

or use the mapping from method to handler:

```go
func main() {
    router := ship.New()
    router.R("/path/to").Map(map[string]ship.Handler{
        "GET": getHandler,
        "POST": postHandler,
        "DELETE": deleteHandler,
    })
    router.Start(":8080").Wait()
}
```

#### Naming route and building URL
You can name the route when registering it, then you can build a URL by the name.

```go
func main() {
    router := ship.New()
    router.Route("/path/:id").Name("get_url").GET(func(ctx *ship.Context) error {
        fmt.Println(ctx.URL("get_url", ctx.URLParam("id")))
        return nil
    })
    router.Start(":8080").Wait()
}
```

#### Add the Header filter

```go
func main() {
    router := ship.New()
    handler := func(ctx *ship.Context) error { return nil }

    // The Content-Type header of the request to /path2 must be application/json,
    // Or it will return 404.
    router.R("/path2").HasHeader("Content-Type", "application/json").POST(handler)
    router.Start(":8080").Wait()
}
```

#### Map methods into Router

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship/v2"
)

type TestType struct{}

func (t TestType) Create(ctx *ship.Context) error { return nil }
func (t TestType) Delete(ctx *ship.Context) error { return nil }
func (t TestType) Update(ctx *ship.Context) error { return nil }
func (t TestType) Get(ctx *ship.Context) error    { return nil }
func (t TestType) Has(ctx *ship.Context) error    { return nil }
func (t TestType) NotHandler()              {}

func main() {
    router := ship.New()
    router.Route("/v1").MapType(TestType{})
    router.Start(":8080").Wait()
}
```

`router.Route("/v1").MapType(TestType{})` is equal to

```go
tv := TestType{}
router.Route("/v1/testtype/get").Name("testtype_get").GET(tv.Get)
router.Route("/v1/testtype/update").Name("testtype_update").PUT(tv.Update)
router.Route("/v1/testtype/create").Name("testtype_create").POST(tv.Create)
router.Route("/v1/testtype/delete").Name("testtype_delete").DELETE(tv.Delete)
```

**Notice:**
- The name of type and method will be converted to the lower.
- The mapping format of the route path is `%{prefix}/%{lower_type_name}/%{lower_method_name}`.
- The mapping format of the route name is `%{lower_type_name}_%{lower_method_name}`.
- The type of the method must be `func(*ship.Context) error` or it will be ignored.

#### Using `SubRouter`

```go
func main() {
    router := ship.New().Use(middleware.Logger(), middleware.Recover())

    // v1 SubRouter, which will inherit the middlewares of the parent router.
    v1 := router.Group("/v1")
    v1.Route("/get/path").GET(getHandler)

    // v2 SubRouter, which won't inherit the middlewares of the parent router.
    v2 := router.Group("/v2").NoMiddlewares().Use(MyAuthMiddleware())
    v2.Route("/post/path").POST(postHandler)

    router.Start(":8080").Wait()
}
```

#### Filter the unacceptable route
```go
func filter(ri ship.RouteInfo) bool {
    if ri.Name == "" {
        return true
    } else if !strings.HasPrefix(ri.Path, "/prefix/") {
        return true
    }
    return false
}

func main() {
    // Don't register the router without name.
    app := ship.New()
    app.RouteFilter = filter

    app.Group("/prefix").R("/name").Name("test").GET(handler) // Register the route
    app.Group("/prefix").R("/noname").GET(handler)            // Don't register the route
    app.R("/no_group").GET(handler)                           // Don't register the route
}
```

#### Modify the registered route
```go
func modifier(ri ship.RouteInfo) ship.RouteInfo {
    ri.Path = "/prefix" + ri.Path
    return ri
}

func main() {
    app := ship.New()
    app.RouteModifier = modifier

    // Register the path as "/prefix/path".
    app.R("/path").Name("test").GET(handler)
}
```

### Using `Middleware`

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship/v2"
    "github.com/xgfone/ship/v2/middleware"
)

func main() {
    // We disable the default error log because we have used the Logger middleware.
    app := ship.New().Use(middleware.Logger(), middleware.Recover())
    app.Use(MyAuthMiddleware())
    app.Route("/url/path").GET(handler)
    app.Start(":8080").Wait()
}
```

You can register a middleware to run before finding the router. You may affect the router finding by registering **Before** middleware. For example,

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship/v2"
    "github.com/xgfone/ship/v2/middleware"
)

func RemovePathPrefix(prefix string) ship.Middleware {
    if len(prefix) < 2 || prefix[len(prefix)-1] == "/" {
        panic(fmt.Errorf("invalid prefix: '%s'", prefix))
    }

    return func(next ship.Handler) Handler {
        return func(ctx *ship.Context) error {
            req := ctx.Request()
            req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
        }
    }
}

func main() {
    router := ship.New()

    // Use and Before have no interference each other.
    router.Use(middleware.Logger())
    router.Pre(RemovePathPrefix("/static"))
    router.Use(middleware.Recover())

    router.Route("/url/path").GET(handler)
    router.Start(":8080").Wait()
}
```

### Add the Virtual Host

```go
package main

import "github.com/xgfone/ship/v2"

func main() {
	router := ship.New()
	router.Route("/router").GET(func(c *ship.Context) error { return c.Text(200, "default") })

	vhost1 := router.Host("host1.example.com") // It is a RouteGroup with the host.
	vhost1.Route("/router").GET(func(c *ship.Context) error { return c.Text(200, "vhost1") })

	vhost2 := router.Host("host2.example.com") // It is a RouteGroup with the host.
	vhost2.Route("/router").GET(func(c *ship.Context) error { return c.Text(200, "vhost2") })

	router.Start(":8080").Wait()
}
```

```shell
$ curl http://127.0.0.1:8080/router
default

$ curl http://127.0.0.1:8080/router -H 'Host: host1.example.com'
vhost1

$ curl http://127.0.0.1:8080/router -H 'Host: host2.example.com'
vhost2
```

### Handle the complex response

```go
package main

import (
	"net/http"

	"github.com/xgfone/ship/v2"
)

func responder(ctx *ship.Context, args ...interface{}) error {
	switch len(args) {
	case 0:
		return ctx.NoContent(http.StatusOK)
	case 1:
		switch v := args[0].(type) {
		case int:
			return ctx.NoContent(v)
		case string:
			return ctx.Text(http.StatusOK, v)
		}
	case 2:
		switch v0 := args[0].(type) {
		case int:
			return ctx.Text(v0, "%v", args[1])
		}
	}
	return ctx.NoContent(http.StatusInternalServerError)
}

func main() {
	app := ship.New()
	app.Responder = responder
	app.Route("/path1").GET(func(c *ship.Context) error { return c.Respond() })
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Respond(200) })
	app.Route("/path3").GET(func(c *ship.Context) error { return c.Respond("Hello, World") })
	app.Route("/path4").GET(func(c *ship.Context) error { return c.Respond(200, "Hello, World") })
	app.Start(":8080").Wait()
}
```

### Bind JSON, XML or Form data form payload

`ship` supply a default data binding to bind the JSON, XML or Form data from payload.

```go
type Login struct {
    Username string `json:"username" xml:"username"`
    Password string `json:"password" xml:"password"`
}

func main() {
    router := ship.Default()

    router.Route("/login").POST(func(ctx *ship.Context) error {
        var login Login
        if err := ctx.Bind(&login); err != nil {
            return err
        }
        ...
    })

    router.Start(":8080").Wait()
}
```

### Render JSON, XML, HTML or other format data

In the directory `/path/to/templates`, there is a template file named `index.tmpl` as follow:
```html
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        This is the body content: </pre>{{ . }}</pre>
    </body>
</html>
```
So we load it as the template by the stdlib `html/template`, and render it as the HTML content.

```go
package main

import (
	"fmt"

	"github.com/xgfone/ship/v2"
	"github.com/xgfone/ship/v2/render"
	"github.com/xgfone/ship/v2/render/template"
)

func main() {
	// It will recursively load all the files in the directory as the templates.
	loader := template.NewDirLoader("/path/to/templates")
	tmplRender := template.NewHTMLTemplateRender(loader)

	router := ship.Default()
	router.Renderer.(*render.MuxRenderer).Add(".tmpl", tmplRender)

	// For JSON
	router.Route("/json").GET(func(ctx *ship.Context) error {
		if ctx.QueryParam("pretty") == "1" {
			return ctx.JSONPretty(200, map[string]interface{}{"msg": "json"}, "    ")
			// Or
			// return ctx.RenderOk("jsonpretty", map[string]interface{}{"msg": "json"})
		}
		return ctx.JSON(200, map[string]interface{}{"msg": "json"})
		// Or
		// return ctx.RenderOk("json", map[string]interface{}{"msg": "json"})
	})

	// For XML
	router.Route("/xml").GET(func(ctx *ship.Context) error {
		if ctx.QueryParam("pretty") == "1" {
			return ctx.XMLPretty(200, []string{"msg", "xml"}, "    ")
			// Or
			// return ctx.RenderOk("xmlpretty", []string{"msg", "xml"})
		}
		return ctx.XML(200, []string{"msg", "xml"})
		// Or
		// return ctx.RenderOk("xml", []string{"msg", "xml"})
	})

	// For HTML
	router.Route("/html").GET(func(ctx *ship.Context) error {
		return ctx.RenderOk("index.tmpl", "Hello World")
	})

	// Start the HTTP server.
	router.Start(":8080").Wait()
}
```

When accessing `http://127.0.0.1:8080/html`, it returns
```html
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        This is the body content: </pre>Hello World</pre>
    </body>
</html>
```

### Prometheus Metric
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/ship/v2"
)

func main() {
	app := ship.New()
	app.R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	app.Start(":8080").Wait()
}
```

You can disable or remove the default collectors like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/ship/v2"
)

// DisableBuiltinCollector removes the collectors that the default prometheus
// register registered.
func DisableBuiltinCollector() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

func main() {
	DisableBuiltinCollector()
	app := ship.New()
	app.R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	app.Start(":8080").Wait()
}
```

The default prometheus HTTP handler, `promhttp.Handler()`, will collect two metrics: `promhttp_metric_handler_requests_in_flight` and `promhttp_metric_handler_requests_total{code="200/500/503"}`. However, you can rewrite it like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/xgfone/ship/v2"
)

// DisableBuiltinCollector removes the collectors that the default prometheus
// register registered.
func DisableBuiltinCollector() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

// Prometheus returns a prometheus handler.
//
// if missing gatherer, it is prometheus.DefaultGatherer by default.
func Prometheus(gatherer ...prometheus.Gatherer) ship.Handler {
	gather := prometheus.DefaultGatherer
	if len(gatherer) > 0 && gatherer[0] != nil {
		gather = gatherer[0]
	}

	return func(ctx *ship.Context) error {
		mfs, err := gather.Gather()
		if err != nil {
			return err
		}

		ct := expfmt.Negotiate(ctx.Request().Header)
		ctx.SetContentType(string(ct))
		enc := expfmt.NewEncoder(ctx, ct)

		for _, mf := range mfs {
			if err = enc.Encode(mf); err != nil {
				ctx.Logger().Errorf("failed to encode prometheus metric: %s", err)
			}
		}

		return nil
	}
}

func main() {
	DisableBuiltinCollector()
	ship.New().R("/metrics").GET(Prometheus()).Ship().Start(":8080").Wait()
}
```


## Route Management

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route with **Zero Garbage** (See [Benchmark](#benchmark)), which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://pkg.go.dev/github.com/xgfone/ship/v2/router/echo?tab=doc#NewRouter).

You can appoint your own implementation by implementing the interface [`Router`](https://pkg.go.dev/github.com/xgfone/ship/v2/router?tab=doc#Router).

```go
type Router interface {
	// Routes returns the list of all the routes.
	Routes() []Route

	// URL generates a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add adds a route and returns the number of the parameters
	// if there are the parameters in the route.
	//
	// For keeping consistent, the parameter should start with ":" or "*".
	// ":" stands for the single parameter, and "*" stands for the wildcard.
	Add(name, method, path string, handler interface{}) (paramNum int, err error)

	// Del deletes the given route.
	//
	// If name is not empty, lookup the path by it instead.
	//
	// If method is empty, deletes all the routes associated with the path.
	// Or only delete the given method for the path.
	Del(name, method, path string) (err error)

	// Find searchs the handler and the number of the url path paramethers.
	// For the url path paramethers, they are put into pnames and pvalues.
	//
	// Return (nil, 0) if not found the route handler.
	Find(method, path string, pnames, pvalues []string) (handler interface{}, pn int)
}
```

```go
func main() {
    NewMyRouter := func() ship.Router { return ... }
    router := ship.New().SetNewRouter(NewMyRouter)
    // ...
}
```


## Benchmark

### Test 1
```
Dell Vostro 3470
Intel Core i5-7400 3.0GHz
8GB DDR4 2666MHz
Windows 10
Go 1.13.4
```

|           Function          |  ops   | ns/op | bytes/opt | allocs/op
|-----------------------------|--------|-------|-----------|-----------
|BenchmarkGinStatic-4         |  23368 | 49788 |    8278   |   157
|BenchmarkGinGitHubAPI-4      |  15684 | 75104 |   10849   |   203
|BenchmarkGinGplusAPI-4       | 276224 |  4184 |     686   |    13
|BenchmarkGinParseAPI-4       | 157810 |  7537 |    1357   |    26
|BenchmarkEchoStatic-4        |  29432 | 39989 |    2432   |   157
|BenchmarkEchoGitHubAPI-4     |  20527 | 56857 |    2468   |   203
|BenchmarkEchoGplusAPI-4      | 387421 |  3179 |     193   |    13
|BenchmarkEchoParseAPI-4      | 220273 |  5575 |     365   |    26
|BenchmarkShipEchoStatic-4    |  34054 | 35548 |    1016   |     0
|BenchmarkShipEchoGitHubAPI-4 |  21842 | 54962 |    1585   |     0
|BenchmarkShipEchoGplusAPI-4  | 402898 |  2996 |      85   |     0
|BenchmarkShipEchoParseAPI-4  | 223581 |  5478 |     154   |     0

### Test 2
```
MacBook Pro(Retina, 13-inch, Mid 2014)
Intel Core i5 2.6GHz
8GB DDR3 1600MHz
macOS Mojave
Go 1.13.4
```

|           Function          |  ops   | ns/op | bytes/opt | allocs/op
|-----------------------------|--------|-------|-----------|-----------
|BenchmarkGinStatic-4         |  18085 | 62380 |    8494   |   157
|BenchmarkGinGitHubAPI-4      |  12646 | 93052 |   11115   |   203
|BenchmarkGinGplusAPI-4       | 224404 |  5222 |     701   |    13
|BenchmarkGinParseAPI-4       | 124138 |  9442 |    1387   |    26
|BenchmarkEchoStatic-4        |  22624 | 47401 |    2021   |   157
|BenchmarkEchoGitHubAPI-4     |  16822 | 69059 |    2654   |   203
|BenchmarkEchoGplusAPI-4      | 326142 |  3759 |     157   |    13
|BenchmarkEchoParseAPI-4      | 178182 |  6713 |     402   |    26
|BenchmarkShipEchoStatic-4    |  27048 | 43713 |     640   |     0
|BenchmarkShipEchoGitHubAPI-4 |  17545 | 66953 |     987   |     0
|BenchmarkShipEchoGplusAPI-4  | 318595 |  3698 |      54   |     0
|BenchmarkShipEchoParseAPI-4  | 175984 |  6807 |     196   |     0
