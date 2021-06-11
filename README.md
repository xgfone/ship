# ship [![Build Status](https://api.travis-ci.com/xgfone/ship.svg?branch=master)](https://travis-ci.com/github/xgfone/ship) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/ship)](https://pkg.go.dev/github.com/xgfone/ship) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a flexible, powerful, high performance and minimalist Go Web HTTP router framework. It is inspired by [echo](https://github.com/labstack/echo) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.

`ship` has been stable, and the current version is `v4` and support Go `1.11+`.


## Features
- Support the url parameter.
- Support the session manager.
- Support the customized router manager.
- Support the pre-route and route middlewares.
- Support the route group builder to build the route.
- Support the mulit-virtual hosts and the default host.
- Support the exact, prefix, suffix and regexp hostname.
- Support the binding of the request data, such as body and query.
- Support the renderer, such as the HTML template.
- ......


## Install

```shell
go get -u github.com/xgfone/ship/v4
```


## Quick Start

```go
// example.go
package main

import "github.com/xgfone/ship/v4"

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

### Route Path
The route path supports the parameters like `:paramName`, `*` or `*restParamName`.

- `/path/to/route` only matches the path `/path/to/route`.
- `/path/:param1/to` matches the path `/path/abc/to`, `/path/xyz/to`, etc. And `:param1` is equal to `abc` or `xyz`.
- `/path/:param1/to/:param2` matches the path `/path/p11/to/p21`, `/path/p12/to/p22`, etc. And `:parma1` is equal to `p11` or `p12`, and `:param2` is equal to `p12` or `p22`.
- `/path/to/*` or `/path/to/*all` matches the path `/path/to/abc`, `/path/to/abc/efg`, `/path/to/xyz`, `/path/to/xyz/123`, etc. And `*` or `*all` is equal to `abc`, `abc/efg`, `xyz`, or `xzy/123`. **Notice:** `*` or `*restParamName` must be the last one of the route path.
- `/path/:param/to/*` matches the path `/path/abc/to/efg`, `/path/abc/to/efg/123`, etc. And `:param` is equal to `abc`, and `*` is equal to `efg` or `efg/123`

For the parameter, it can be accessed by `Context.URLParam(paramName)`.

- For `*`, the parameter name is `*`, like `Context.URLParam("*")`.
- For `*restParamName`, the parameter name is `restParamName`, like `Context.URLParam(restParamName)`.


## API Example

### `Router`
#### Using `CONNECT`, `GET`, `POST`, `PUT`, `PATCH`, `DELETE` and `OPTION`

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


#### Cascade the registered routes

```go
func main() {
    router := ship.New()
    router.Route("/path/to").GET(getHandler).POST(postHandler).DELETE(deleteHandler)
    router.Start(":8080").Wait()
}
```

or use the mapping from method to handler:

```go
func main() {
    router := ship.New()
    router.Route("/path/to").Map(map[string]ship.Handler{
        "GET": getHandler,
        "POST": postHandler,
        "DELETE": deleteHandler,
    })
    router.Start(":8080").Wait()
}
```

#### Naming route and building URL
When registering the route, it can be named with a name, then build a url path by the name.

```go
func main() {
    router := ship.New()
    router.Route("/path/:id").Name("get_url").GET(func(ctx *ship.Context) error {
        fmt.Println(ctx.URLPath("get_url", ctx.URLParam("id")))
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
    router.Route("/path2").HasHeader("Content-Type", "application/json").POST(handler)
    router.Start(":8080").Wait()
}
```

#### Using `SubRouter`

```go
func main() {
    router := ship.New().Use(middleware.Logger(nil), middleware.Recover())

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

    app.Group("/prefix").Route("/name").Name("test").GET(handler) // Register the route
    app.Group("/prefix").Route("/noname").GET(handler)            // Don't register the route
    app.Route("/no_group").GET(handler)                           // Don't register the route
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
    app.Route("/path").Name("test").GET(handler)
}
```

### Using `Middleware`

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship/v4"
    "github.com/xgfone/ship/v4/middleware"
)

func main() {
    // We disable the default error log because we have used the Logger middleware.
    app := ship.New().Use(middleware.Logger(nil), middleware.Recover())
    app.Use(MyAuthMiddleware())
    app.Route("/url/path").GET(handler)
    app.Start(":8080").Wait()
}
```

You can register a **Before** middleware to be run before finding the router to affect the route match. For example,

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship/v4"
    "github.com/xgfone/ship/v4/middleware"
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
    router.Use(middleware.Logger(nil))
    router.Pre(RemovePathPrefix("/static"))
    router.Use(middleware.Recover())

    router.Route("/url/path").GET(handler)
    router.Start(":8080").Wait()
}
```

### Add the Virtual Host

```go
package main

import "github.com/xgfone/ship/v4"

func main() {
	router := ship.New()
	router.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "default") })

	// Exact Match Host
	vhost1 := router.Host("www.host1.example.com")
	vhost1.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost1") })

	// Suffix Match Host
	vhost2 := router.Host("*.host2.example.com")
	vhost2.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost2") })

	// Prefix Match Host
	vhost3 := router.Host("www.host3.*")
	vhost3.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost3") })

	// Regexp Match Host by using Go regexp package
	vhost4 := router.Host(`www\.[a-zA-z0-9]+\.example\.com`)
	vhost4.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost4") })

	router.Start(":8080").Wait()
}
```

```shell
$ curl http://127.0.0.1:8080/
default

$ curl http://127.0.0.1:8080/ -H 'Host: www.host1.example.com' # Exact
vhost1

$ curl http://127.0.0.1:8080/ -H 'Host: www.host2.example.com' # Suffix
vhost2

$ curl http://127.0.0.1:8080/ -H 'Host: www.host3.example.com' # Prefix
vhost3

$ curl http://127.0.0.1:8080/ -H 'Host: www.host4.example.com' # Regexp
vhost4
```

### Handle the complex response

```go
package main

import (
	"net/http"

	"github.com/xgfone/ship/v4"
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

	"github.com/xgfone/ship/v4"
	"github.com/xgfone/ship/v4/render"
	"github.com/xgfone/ship/v4/render/template"
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
		}
		return ctx.JSON(200, map[string]interface{}{"msg": "json"})
	})

	// For XML
	router.Route("/xml").GET(func(ctx *ship.Context) error {
		if ctx.QueryParam("pretty") == "1" {
			return ctx.XMLPretty(200, []string{"msg", "xml"}, "    ")
		}
		return ctx.XML(200, []string{"msg", "xml"})
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
	"github.com/xgfone/ship/v4"
)

func main() {
	app := ship.New()
	app.Route("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	app.Start(":8080").Wait()
}
```

The default collectors can be disabled or removed like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/ship/v4"
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
	app.Route("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	app.Start(":8080").Wait()
}
```

The default prometheus HTTP handler, `promhttp.Handler()`, will collect two metrics: `promhttp_metric_handler_requests_in_flight` and `promhttp_metric_handler_requests_total{code="200/500/503"}`. However, it can be rewrote like this.
```go
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/xgfone/ship/v4"
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
	ship.New().Route("/metrics").GET(Prometheus()).Ship().Start(":8080").Wait()
}
```

### OpenTracing
```go
import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/xgfone/ship/v4"
)

// OpenTracingOption is used to configure the OpenTracingServer.
type OpenTracingOption struct {
	Tracer        opentracing.Tracer // Default: opentracing.GlobalTracer()
	ComponentName string             // Default: "net/http"

	// URLTagFunc is used to get the value of the tag "http.url".
	// Default: url.String()
	URLTagFunc func(*url.URL) string

	// SpanFilter is used to filter the span if returning true.
	// Default: return false
	SpanFilter func(*http.Request) bool

	// OperationNameFunc is used to the operation name.
	// Default: fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
	OperationNameFunc func(*http.Request) string

	// SpanObserver is used to do extra things of the span for the request.
	//
	// For example,
	//    OpenTracingOption {
	//        SpanObserver: func(*http.Request, opentracing.Span) {
	//            ext.PeerHostname.Set(span, req.Host)
	//        },
	//    }
	//
	// Default: Do nothing.
	SpanObserver func(*http.Request, opentracing.Span)
}

// Init initializes the OpenTracingOption.
func (o *OpenTracingOption) Init() {
	if o.ComponentName == "" {
		o.ComponentName = "net/http"
	}
	if o.URLTagFunc == nil {
		o.URLTagFunc = func(u *url.URL) string { return u.String() }
	}
	if o.SpanFilter == nil {
		o.SpanFilter = func(r *http.Request) bool { return false }
	}
	if o.SpanObserver == nil {
		o.SpanObserver = func(*http.Request, opentracing.Span) {}
	}
	if o.OperationNameFunc == nil {
		o.OperationNameFunc = func(r *http.Request) string {
			return fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
		}
	}
}

// GetTracer returns the OpenTracing tracker.
func (o *OpenTracingOption) GetTracer() opentracing.Tracer {
	if o.Tracer == nil {
		return opentracing.GlobalTracer()
	}
	return o.Tracer
}

// NewOpenTracingRoundTripper returns a new OpenTracingRoundTripper.
func NewOpenTracingRoundTripper(rt http.RoundTripper, opt *OpenTracingOption) *OpenTracingRoundTripper {
	var o OpenTracingOption
	if opt != nil {
		o = *opt
	}
	o.Init()
	return &OpenTracingRoundTripper{RoundTripper: rt, OpenTracingOption: o}
}

// WrappedRoundTripper returns the wrapped http.RoundTripper.
func (rt *OpenTracingRoundTripper) WrappedRoundTripper() http.RoundTripper {
	return rt.RoundTripper
}

func (rt *OpenTracingRoundTripper) roundTrip(req *http.Request) (*http.Response, error) {
	if rt.RoundTripper == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return rt.RoundTripper.RoundTrip(req)
}

// RoundTrip implements the interface http.RounderTripper.
func (rt *OpenTracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.SpanFilter(req) {
		return rt.roundTrip(req)
	}

	operationName := rt.OperationNameFunc(req)
	sp, ctx := opentracing.StartSpanFromContext(req.Context(), operationName)
	ext.HTTPMethod.Set(sp, req.Method)
	ext.Component.Set(sp, rt.ComponentName)
	ext.HTTPUrl.Set(sp, rt.URLTagFunc(req.URL))
	rt.SpanObserver(req, sp)
	defer sp.Finish()

	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	rt.GetTracer().Inject(sp.Context(), opentracing.HTTPHeaders, carrier)

	return rt.roundTrip(req.WithContext(ctx))
}

// OpenTracing is a middleware to support the OpenTracing.
func OpenTracing(opt *OpenTracingOption) Middleware {
	var o OpenTracingOption
	if opt != nil {
		o = *opt
	}
	o.Init()

	const format = opentracing.HTTPHeaders
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			req := ctx.Request()
			if o.SpanFilter(req) {
				return next(ctx)
			}

			tracer := o.GetTracer()
			sc, _ := tracer.Extract(format, opentracing.HTTPHeadersCarrier(req.Header))
			sp := tracer.StartSpan(o.OperationNameFunc(req), ext.RPCServerOption(sc))

			ext.HTTPMethod.Set(sp, req.Method)
			ext.Component.Set(sp, o.ComponentName)
			ext.HTTPUrl.Set(sp, o.URLTagFunc(req.URL))
			o.SpanObserver(req, sp)

			req = req.WithContext(opentracing.ContextWithSpan(req.Context(), sp))
			ctx.SetRequest(req)

			defer func() {
				if e := recover(); e != nil {
					ext.Error.Set(sp, true)
					sp.Finish()
					panic(e)
				}

				statusCode := ctx.StatusCode()
				if !ctx.IsResponded() {
					switch e := err.(type) {
					case nil:
					case ship.HTTPError:
						statusCode = e.Code
					default:
						statusCode = 500
					}
				}

				ext.HTTPStatusCode.Set(sp, uint16(statusCode))
				if statusCode >= 500 {
					ext.Error.Set(sp, true)
				}
				sp.Finish()
			}()

			err = next(ctx)
			return err
		}
	}
}

func init() {
	// TODO: Initialize the global OpenTracing tracer.

	// Replace the default global RoundTripper.
	http.DefaultTransport = NewOpenTracingRoundTripper(http.DefaultTransport, nil)
}

func main() {
	app := ship.Default()
	app.Use(OpenTracing(nil))
	app.Route("/").GET(func(c *ship.Context) error {
		ctx := c.Request().Context() // ctx contains the parent span context
		req, err := http.NewRequestWithContext(ctx, METHOD, URL, BODY)
		if err != nil {
			return
		}
		// TODO with req ...
	})
	app.Start(":8080").Wait()
}
```


## Route Management

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route with **Zero Garbage** (See [Benchmark](#benchmark)), which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://pkg.go.dev/github.com/xgfone/ship/v4/router/echo?tab=doc#NewRouter).

You can appoint your own implementation by implementing the interface [`Router`](https://pkg.go.dev/github.com/xgfone/ship/v4/router?tab=doc#Router).

```go
type Router interface {
	// Routes uses the filter to filter and return the routes if it returns true.
	//
	// Return all the routes if filter is nil.
	Routes(filter func(name, path, method string) bool) []Route

	// Path generates a url path by the path name and parameters.
	//
	// Return "" if there is not the route path named name.
	Path(name string, params ...interface{}) string

	// Add adds the route and returns the number of the parameters
	// if there are the parameters in the route path.
	//
	// name is the name of the path, which is optional and must be unique
	// if not empty.
	//
	// If method is empty, handler is the handler of all the methods supported
	// by the implementation. Or, it is only that of the given method.
	//
	// For the parameter in the path, the format is determined by the implementation.
	Add(name, path, method string, handler interface{}) (paramNum int, err error)

	// Del deletes the given route.
	//
	// If method is empty, deletes all the routes associated with the path.
	// Or, only delete the given method for the path.
	Del(path, method string) (err error)

	// Match matches the route by path and method, puts the path parameters
	// into pnames and pvalues, then returns the handler and the number
	// of the path paramethers.
	//
	// If pnames or pvalues is empty, it will ignore the path paramethers
	// when finding the route handler.
	//
	// Return (nil, 0) if not found the route handler.
	Match(path, method string, pnames, pvalues []string) (handler interface{}, pn int)
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
|Benchmark**Gin**Static-4         |  23368 | 49788 |    8278   |   157
|Benchmark**Gin**GitHubAPI-4      |  15684 | 75104 |   10849   |   203
|Benchmark**Gin**GplusAPI-4       | 276224 |  4184 |     686   |    13
|Benchmark**Gin**ParseAPI-4       | 157810 |  7537 |    1357   |    26
|Benchmark**Echo**Static-4        |  29432 | 39989 |    2432   |   157
|Benchmark**Echo**GitHubAPI-4     |  20527 | 56857 |    2468   |   203
|Benchmark**Echo**GplusAPI-4      | 387421 |  3179 |     193   |    13
|Benchmark**Echo**ParseAPI-4      | 220273 |  5575 |     365   |    26
|Benchmark**ShipEcho**Static-4    |  34054 | 35548 |    1016   | **0**
|Benchmark**ShipEcho**GitHubAPI-4 |  21842 | 54962 |    1585   | **0**
|Benchmark**ShipEcho**GplusAPI-4  | 402898 |  2996 |      85   | **0**
|Benchmark**ShipEcho**ParseAPI-4  | 223581 |  5478 |     154   | **0**

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
|Benchmark**Gin**Static-4         |  18085 | 62380 |    8494   |   157
|Benchmark**Gin**GitHubAPI-4      |  12646 | 93052 |   11115   |   203
|Benchmark**Gin**GplusAPI-4       | 224404 |  5222 |     701   |    13
|Benchmark**Gin**ParseAPI-4       | 124138 |  9442 |    1387   |    26
|Benchmark**Echo**Static-4        |  22624 | 47401 |    2021   |   157
|Benchmark**Echo**GitHubAPI-4     |  16822 | 69059 |    2654   |   203
|Benchmark**Echo**GplusAPI-4      | 326142 |  3759 |     157   |    13
|Benchmark**Echo**ParseAPI-4      | 178182 |  6713 |     402   |    26
|Benchmark**ShipEcho**Static-4    |  27048 | 43713 |     640   | **0**
|Benchmark**ShipEcho**GitHubAPI-4 |  17545 | 66953 |     987   | **0**
|Benchmark**ShipEcho**GplusAPI-4  | 318595 |  3698 |      54   | **0**
|Benchmark**ShipEcho**ParseAPI-4  | 175984 |  6807 |     196   | **0**
