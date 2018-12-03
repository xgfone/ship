# ship [![Build Status](https://travis-ci.org/xgfone/ship.svg?branch=master)](https://travis-ci.org/xgfone/ship) [![Coverage Status](https://coveralls.io/repos/github/xgfone/ship/badge.svg?branch=master)](https://coveralls.io/github/xgfone/ship?branch=master) [![GoDoc](https://godoc.org/github.com/xgfone/ship?status.svg)](http://godoc.org/github.com/xgfone/ship) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a A flexible and powerful HTTP router implemented by golang, which uses the interface to define the router, so you can supply yourself implementation and combine them with the default implementation.

`ship` is inspired by [echo](https://github.com/labstack/echo), [pure](https://github.com/go-playground/pure) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.


## Install

```shell
go get -u github.com/xgfone/ship
```

For the core functions, **it has no any third-party dependencies.**


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
	router := ship.NewRouter()
	router.Get("/ping", func(ctx ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	})

	http.ListenAndServe(":8080", router)
}
```

```shell
$ go run example.go
```

```shell
$ curl http://127.0.0.1:8080/ping
{"message":"pong"}
```


## Configure `Router`

```go
type Config struct {
	// The route prefix.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool
	// If ture, it will clean the request path before finding the route.
	CleanPath bool

	// You can customize the logger implementation, or use NewNoLevelLogger(os.Stdout).
	Logger Logger
	// You can customize the Binder implementation, or use NewBinder().
	Binder Binder
	// You can customize the Renderer implementation.
	Renderer Renderer

	// You can customize the Route management, or use NewRoute().
	// The default implementation is based on Radix tree,
	// which refers to https://github.com/go-playground/pure.
	NewRoute func() Route
	// You can customize the Context implementation, or use NewContext().
	NewContext func() Context
	// You can customize the URLParam implementation, or use NewURLParam().
	NewURLParam func(int) URLParam
	// You can check and filter the response output before sending the peer.
	FilterOutput func([]byte) []byte

	// You can appoint the error handler, or use HandleHTTPError().
	HandleError func(ctx Context, err error)
	// You can appoint the panic handler.
	HandlePanic func(ctx Context, panicValue interface{})

	// You can appoint the OPTIONS handler.
	OptionsHandler Handler
	// You can appoint the NotFound handler. Or use NotFoundHandler.
	NotFoundHandler Handler
	// You can appoint the MethodNotAllowed handler.
	MethodNotAllowedHandler Handler
}
```

```go
func main() {
    config := ship.Config{}
    router := ship.NewRouter(config)

    ...
}
```


## API Example

See [GoDOC](https://godoc.org/github.com/xgfone/ship).

### `Router`
#### Using `Connect`, `Get`, `Post`, `Put`, `Patch`, `Delete` and `Option`

```go
func main() {
    router := ship.NewRouter()

    router.Get("/path/get", getHandler)
    router.Put("/path/put", putHandler)
    router.Post("/path/post", postHandler)
    router.Patch("/path/patch", patchHandler)
    router.Delete("/path/delete", deleteHandler)
    router.Option("/path/option", optionHandler)
    router.Connect("/path/connect", connectHandler)

    http.ListenAndServe(":8080", router)
}
```

Notice: you can regitser the customized method by `Methods(methods []string, path string, handler Handler)`.

#### Naming route and building URL
You can name the route a name when registering the route, then you can build a URL by the name.

```go
func main() {
    router := ship.NewRouter()

    router.Get("/path/:id", func(ctx Context) error {
        ctx.URL("get_url", ctx.URLParam())
    }, "get_url")

    http.ListenAndServe(":8080", router)
}
```

#### Map methods into Router

```go
package main

import (
    "net/http"

    "github.com/xgfone/ship"
)

type TestStruct struct{}

func (t TestStruct) Create(ctx ship.Context) error { return nil }
func (t TestStruct) Delete(ctx ship.Context) error { return nil }
func (t TestStruct) Update(ctx ship.Context) error { return nil }
func (t TestStruct) Get(ctx ship.Context) error    { return nil }
func (t TestStruct) Has(ctx ship.Context) error    { return nil }
func (t TestStruct) NotHandler()                   {}

func main() {
    router := ship.NewRouter()

    ship.MapMethodIntoRouter(router, TestStruct{}, "/v1")

    http.ListenAndServe(":8080", router)
}
```

`ship.MapMethodIntoRouter(router, TestStruct{}, "/v1")` is equal to

```go
ts := TestStruct{}
router.Get("/v1/teststruct/get", ts.Get, "teststruct_get")
router.Put("/v1/teststruct/update", ts.Update, "teststruct_update")
router.Post("/v1/teststruct/create", ts.Create, "teststruct_create")
router.Delete("/v1/teststruct/delete", ts.Delete, "teststruct_delete")
```

The default mapping method is `DefaultMethodMapping`, which is defined as follow.
```go
var DefaultMethodMapping = map[string]string{
    // "MethodName": "RequestMethod"
    "Create": "POST",
    "Delete": "DELETE",
    "Update": "PUT",
    "Get":    "GET",
}
```

If the default is not what you want, you can customize it, for example,
```go
ship.MapMethodIntoRouter(router, TestStruct{}, "/v1", map[string]string{
    "GetMethod": "GET",
    "PostMethod": "POST",
})
```

**Notice:**
- The name of type and method will be converted to the lower.
- The mapping format of the route path is `%{prefix}/%{lower_type_name}/%{lower_method_name}`.
- The mapping format of the route name is `%{lower_type_name}_%{lower_method_name}`.

#### Using `Middleware`

```go
func main() {
    router := ship.NewRouter()

    router.Use(ship.NewLoggerMiddleware(), ship.NewRecoverMiddleware())
    router.Use(MyAuthMiddleware())

    router.Get("/url/path", handler)

    http.ListenAndServe(":8080", router)
}
```

You can register a middleware to run it before finding the router. So you maybe affect the router finding by registering **Before** middlewares. For example,

```go
func RemovePathPrefix(prefix string) ship.Middleware {
    if len(prefix) < 2 || prefix[len(prefix)-1] == "/" {
        panic(fmt.Errorf("invalid prefix: '%s'", prefix))
    }

    return func(next ship.Handler) Handler {
        return func(ctx ship.Context) error {
            req := ctx.Request()
            req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
        }
    }
}

func main() {
    router := ship.NewRouter()

    // Use and Before have no interference each other.
    router.Use(ship.NewLoggerMiddleware())
    router.Before(RemovePathPrefix("/static"))
    router.Use(ship.NewRecoverMiddleware())

    router.Get("/url/path", handler)

    http.ListenAndServe(":8080", router)
}
```

**After** Middleware is the same as **Before**, except running the middlewares after routing and before executing the found handler.

#### Using `SubRouter`

```go
func main() {
    router := ship.NewRouter()

    router.Use(ship.NewLoggerMiddleware())
    router.Use(ship.NewRecoverMiddleware())

    // v1 SubRouter, which will inherit the middlewares of the parent router.
    v1 := router.SubRouter("/v1")
    v1.Get("/get/path", getHandler)

    // v2 SubRouter, which won't inherit the middlewares of the parent router.
    v2 := router.SubRouterNone("/v2")
    v2.Use(MyAuthMiddleware())
    v2.Post("/post/path", postHandler)

    http.ListenAndServe(":8080", router)
}
```

#### Traverse the registered route

```go
func main() {
    router := ship.NewRouter()

    router.Get("/get/path", getHandler, "get_name")
    router.Post("/post/path", posttHandler, "post_name")

    router.Each(func(name, method, path string, handler Handler) {
        fmt.Println(name, method, path)
        // Output:
        // get_name GET /get/path
        // post_name POST /post/path
    })

    http.ListenAndServe(":8080", router)
}
```


## `Context`

See [the interface doc](https://godoc.org/github.com/xgfone/ship#Context).


## Bind JSON, XML or Form data form payload

`ship` supply a default data binding to bind the JSON, XML or Form data from payload.

```go
type Login struct {
    Username string `json:"username" xml:"username"`
    Password string `json:"password" xml:"password"`
}

func main() {
    router := ship.NewRouter()

    router.Post("/login", func(ctx ship.Context) error {
        var login Login
        if err := ctx.Bind(&login); err != nil {
            return err
        }
        ...
    })

    http.ListenAndServe(":8080", router)
}
```


## Route Management

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route, which refers to [pure](https://github.com/go-playground/pure), that's, `ship.NewRoute()`.

You can appoint yourself implementation by implementing the interface `ship.Route`.

```go
type Route interface {
    AddRoute(name string, method string, path string, handler Handler) (paramMaxNum int)
    FindRoute(method string, path string, newURLParam func() URLParam) (Handler, URLParam)
    URL(name string, params URLParam) string
}
```

```go
func main() {
    config := ship.Config{NewRoute: MyNewRoute}
    router := ship.NewRouter(config)
    ...
}
```
