# ship [![Build Status](https://travis-ci.org/xgfone/ship.svg?branch=master)](https://travis-ci.org/xgfone/ship) [![GoDoc](https://godoc.org/github.com/xgfone/ship?status.svg)](http://godoc.org/github.com/xgfone/ship) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a flexible, powerful, high performance and minimalist Go Web HTTP router framework. It is inspired by [echo](https://github.com/labstack/echo) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.

`ship` has been stable, and the current version is `v1`.


## Install

```shell
go get -u github.com/xgfone/ship
```

For the core functions, **it has no any third-party dependencies**, except for `github.com/xgfone/go-tools`.


## Prerequisite

Now `ship` requires Go `1.7+`.


## Quick Start

```go
// example.go

package main

import (
	"net/http"

	"github.com/xgfone/ship"
)

func main() {
	router := ship.New()
	router.Route("/ping").GET(func(ctx *ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	// Start the HTTP server.
	router.Start(":8080")
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

See [GoDOC](https://godoc.org/github.com/xgfone/ship).

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

    // Start the HTTP server.
    router.Start(":8080")
}
```

Notice: you can register the same handler with more than one method by `Route(path string).Method(handler Handler, method ...string)`.

`R` is the alias of `Route`, you can register the routes by `R(path string).Method(handler Handler, method ...string)`.

There is a default global ship, `DefaultShip`, and some its methods as the global functions, such as `Pre`, `Use`, `G`, `R`, `GroupWithoutMiddleware`, `RouteWithoutMiddleware`, `URL`, `Start`, `StartServer`, `Wait`, `Shutdown`. **Suggest:** You should use the default `DefaultShip` preferentially.

#### Cascade the registered routes

```go
func main() {
    router := ship.New()
    router.R("/path/to").GET(getHandler).POST(postHandler).DELETE(deleteHandler)

    // Start the HTTP server.
    router.Start(":8080")
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

    // Start the HTTP server.
    router.Start(":8080")
}
```

#### Naming route and building URL
You can name the route when registering it, then you can build a URL by the name.

```go
func main() {
    router := ship.New()

    router.Route("/path/:id").Name("get_url").GET(func(ctx *ship.Context) error {
        fmt.Println(ctx.URL("get_url", ctx.Param("id)))
    })

    // Start the HTTP server.
    router.Start(":8080")
}
```

#### Add the Header and Scheme filter

```go
func main() {
    router := ship.New()

    handler := func(ctx *ship.Context) error { return nil }
    router.R("/path1").Schemes("https", "wss").GET(handler)
    router.R("/path2").Headers("Content-Type", "application/json").POST(handler)

    // Start the HTTP server.
    router.Start(":8080")
}
```

#### Map methods into Router

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship"
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
    router.Start(":8080")
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

The default method mapping is as follow, which can be reset by `SetDefaultMethodMapping` when configuring `Ship`.
```go
opt := SetDefaultMethodMapping (map[string]string{
    // "MethodName": "RequestMethod"
    "Create": "POST",
    "Delete": "DELETE",
    "Update": "PUT",
    "Get":    "GET",
})
router.Configure(opt)
```

**Notice:**
- The name of type and method will be converted to the lower.
- The mapping format of the route path is `%{prefix}/%{lower_type_name}/%{lower_method_name}`.
- The mapping format of the route name is `%{lower_type_name}_%{lower_method_name}`.
- The type of the method must be `func(*ship.Context) error` or it will be ignored.

#### Using `SubRouter`

```go
func main() {
    router := ship.New()

    router.Use(middleware.Logger())
    router.Use(middleware.Recover())

    // v1 SubRouter, which will inherit the middlewares of the parent router.
    v1 := router.Group("/v1")
    v1.Route("/get/path").GET(getHandler)

    // v2 SubRouter, which won't inherit the middlewares of the parent router.
    v2 := router.GroupWithoutMiddleware("/v2")
    v2.Use(MyAuthMiddleware())
    v2.Route("/post/path").POST(postHandler)

    router.Start(":8080")
}
```

#### Traverse the registered route

```go
func main() {
    router := ship.New()

    router.Route("/get/path").Name("get_name").GET(getHandler)
    router.Route("/post/path").Name("post_name").POST(posttHandler)

    router.Traverse(func(name, method, path string) {
        fmt.Println(name, method, path)
        // Output:
        // get_name GET /get/path
        // post_name POST /post/path
    })

    router.Start(":8080")
}
```

### Using `Middleware`

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship"
    "github.com/xgfone/ship/middleware"
)

func main() {
    router := ship.New()
    router.Use(middleware.Logger(), middleware.Recover())
    router.Use(MyAuthMiddleware())
    router.Route("/url/path").GET(handler)
    router.Start(":8080")
}
```

You can register a middleware to run before finding the router. You may affect the router finding by registering **Before** middleware. For example,

```go
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
    router.Start(":8080")
}
```

The sub-packages [`middleware`](https://godoc.org/github.com/xgfone/ship/middleware) has implemented some middleware as follows:

- [CORS](https://godoc.org/github.com/xgfone/ship/middleware#CORS)
- [CSRF](https://godoc.org/github.com/xgfone/ship/middleware#CSRF)
- [Flat](https://godoc.org/github.com/xgfone/ship/middleware#Flat)
- [Gzip](https://godoc.org/github.com/xgfone/ship/middleware#Gzip)
- [Logger](https://godoc.org/github.com/xgfone/ship/middleware#Logger)
- [Recover](https://godoc.org/github.com/xgfone/ship/middleware#Recover)
- [Matchers](https://godoc.org/github.com/xgfone/ship/middleware#Matchers)
- [CleanPath](https://godoc.org/github.com/xgfone/ship/middleware#CleanPath)
- [BodyLimit](https://godoc.org/github.com/xgfone/ship/middleware#BodyLimit)
- [RequestID](https://godoc.org/github.com/xgfone/ship/middleware#RequestID)
- [TokenAuth](https://godoc.org/github.com/xgfone/ship/middleware#TokenAuth)
- [MaxRequests](https://godoc.org/github.com/xgfone/ship/middleware#MaxRequests)
- [ResetResponse](https://godoc.org/github.com/xgfone/ship/middleware#ResetResponse)
- [SetCtxHandler](https://godoc.org/github.com/xgfone/ship/middleware#SetCtxHandler)
- [RemoveTrailingSlash](https://godoc.org/github.com/xgfone/ship/middleware#RemoveTrailingSlash)

### Add the Virtual Host

```go
func main() {
    router := ship.New()
    router.Route("/router").GET(func(c *ship.Context) error { return c.String(200, "default") })

    vhost1 := router.VHost("host1.example.com")
    vhost1.Route("/router").GET(func(c *ship.Context) error { return c.String(200, "vhost1") })

    vhost2 := router.VHost("host2.example.com")
    vhost2.Route("/router").GET(func(c *ship.Context) error { return c.String(200, "vhost2") })

    router.Start(":8080")
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

### Many HTTP Server

```go
package main

import (
	"github.com/xgfone/ship"
)

func main() {
	app1 := ship.New(ship.SetName("app1"))
	app1.Route("/").GET(func(ctx *ship.Context) error { return ctx.String(200, "app1") })

	app2 := app1.Clone("app2").Link(app1)
	app2.Route("/").GET(func(ctx *ship.Context) error { return ctx.String(200, "app2") })

	go app2.Start(":8001")
	app1.Start(":8000").Wait()
}
```

When you runs it, the console will output like this:
```shell
2019/03/07 22:25:41.701704 ship.go:571: [I] The HTTP Server [app1] is running on :8000
2019/03/07 22:25:41.702308 ship.go:571: [I] The HTTP Server [app2] is running on :8001
```

Then you can access `http://127.0.0.1:8000` and `http://127.0.0.1:8001`. The servers returns `app1` and `app2`.

When you keys `Ctrl+C`, the two servers will exit and output like this.
```shell
2019/03/07 22:26:16.243693 once.go:44: [I] The HTTP Server [app1] is shutdown
2019/03/07 22:26:16.243998 once.go:44: [I] The HTTP Server [app2] is shutdown
```

**Notice:**
1. The router app returned by `ship.New()` will listen on the signals.
2. For `Clone()`, `app2` will also exit when `app1` exits.
3. For `Link()`, both `app1` and `app2` will exit when either exits.

### Handle the complex response

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/xgfone/ship"
	"github.com/xgfone/ship/middleware"
)

func Responder := func(ctx *ship.Context, args ...interface{}) error {
	switch len(args) {
	case 0:
		return ctx.NoContent(http.StatusOK)
	case 1:
		switch v := args[0].(type) {
		case int:
			return ctx.NoContent(v)
		case string:
			return ctx.String(http.StatusOK, v)
		}
	case 2:
		switch v0 := args[0].(type) {
		case int:
			return ctx.String(v0, "%v", args[1])
		}
	}
	return ctx.NoContent(http.StatusInternalServerError)
}

func main() {
	router := ship.New()
	router.Use(middleware.SetCtxHandler(Responder))
	router.Route("/path1").GET(func(c *ship.Context) error { return c.Handle() })
	router.Route("/path2").GET(func(c *ship.Context) error { return c.Handle(200) })
	router.Route("/path3").GET(func(c *ship.Context) error { return c.Handle("Hello, World") })
	router.Route("/path4").GET(func(c *ship.Context) error { return c.Handle(200, "Hello, World") })

	router.Start(":8080")
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
    router := ship.New()

    router.Route("/login").POST(func(ctx *ship.Context) error {
        var login Login
        if err := ctx.Bind(&login); err != nil {
            return err
        }
        ...
    })

    router.Start(":8080")
}
```

### Render JSON, XML, HTML or other format data

```go
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
	router.MuxRenderer().Add(engine.Ext(), engine)

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

	// Start the HTTP server.
	router.Start(":8080")
}
```

### Prometheus Metric
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/ship"
)

func main() {
	app := ship.New()
	app.R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	app.Start(":8080")
	app.Wait()

	// You can write it like this:
	//
	// ship.New().R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler())).Ship().Start(":8080").Wait()
}
```

You can disable or remove the default collectors like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/ship"
)

// DisableBuiltinCollector removes the collectors that the default prometheus
// register registered.
func DisableBuiltinCollector() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

func main() {
	DisableBuiltinCollector()
	ship.New().R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler())).Ship().Start(":8080").Wait()
}
```

The default prometheus HTTP handler, `promhttp.Handler()`, will collect two metrics: `promhttp_metric_handler_requests_in_flight` and `promhttp_metric_handler_requests_total{code="200/500/503"}`. However, you can rewrite it like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/xgfone/ship"
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
				ctx.Logger().Error("failed to encode prometheus metric: %s", err)
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

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route, which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://godoc.org/github.com/xgfone/ship/router/echo#NewRouter).

You can appoint your own implementation by implementing the interface [`ship.Router`](https://godoc.org/github.com/xgfone/ship/core#Router).

```go
type Router interface {
	// Generate a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add a route with name, path, method and handler,
	// and return the number of the parameters if there are the parameters
	// in the route. Or return 0.
	//
	// If the name has been added for the same path, it should be allowed.
	// Or it should panic.
	//
	// If the router does not support the parameter, it should panic.
	//
	// Notice: for keeping consistent, the parameter should start with ":"
	// or "*". ":" stands for a single parameter, and "*" stands for
	// a wildcard parameter.
	Add(name string, path string, method string, handler interface{}) (paramNum int)

	// Find a route handler by the method and path of the request.
	//
	// Return nil if the route does not exist.
	//
	// If the route has more than one parameter, the name and value
	// of the parameters should be stored `pnames` and `pvalues` respectively.
	Find(method string, path string, pnames []string, pvalues []string) (handler interface{})

	// Traverse each route.
	Each(func(name string, method string, path string))
}
```

```go
func main() {
    NewMyRouter := func() ship.Router { return ... }
    router := ship.New(ship.SetNewRouter(NewMyRouter))
    ...
}
```

## TODO

- [x] Add Host match for `Route`, referring to [mux.Route.Host](https://godoc.org/github.com/gorilla/mux#Route.Host). _We use the **Virtual Host** instead._
- [x] Add Query match for `Route`, referring to [mux.Route.Queries](https://godoc.org/github.com/gorilla/mux#Route.Queries). _We use `Matcher` to operate it._
- [x] Add the serialization and deserialization middleware. _We use `Binder` and `Renderer`._
- [x] Add HTML template render.
- [x] Add CORS middlware.
- [ ] Add JWT middleware.
- [ ] Add OAuth 2.0 middleware.
