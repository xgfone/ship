# ship [![Build Status](https://travis-ci.org/xgfone/ship.svg?branch=master)](https://travis-ci.org/xgfone/ship) [![Coverage Status](https://coveralls.io/repos/github/xgfone/ship/badge.svg?branch=master)](https://coveralls.io/github/xgfone/ship?branch=master) [![GoDoc](https://godoc.org/github.com/xgfone/ship?status.svg)](http://godoc.org/github.com/xgfone/ship) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a flexible, powerful, high performance and minimalist Go Web HTTP router framework.

`ship` is inspired by [echo](https://github.com/labstack/echo) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.


## Install

```shell
go get -u github.com/xgfone/ship
```

For the core functions, **it has no any third-party dependencies.**


## Prerequisite

Now `ship` requires Go `1.9+`.


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
	router.Route("/ping", func(ctx ship.Context) error {
		return ctx.JSON(200, map[string]interface{}{"message": "pong"})
	}).GET()

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
	// The route prefix, which is "" by default.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool

	// The router management, which uses echo implementation by default.
	// But you can appoint yourself customized Router implementation.
	Router Router

	// The logger management, which is `NewNoLevelLogger(os.Stdout)` by default.
	// But you can appoint yourself customized Logger implementation.
	Logger Logger
	// Binder is used to bind the request data to the given value,
	// which is `NewBinder()` by default.
	// But you can appoint yourself customized Binder implementation
	Binder Binder
	// Rendered is used to render the response to the peer, which has no
	// the default implementation.
	Renderer Renderer

	// Handle the error at last.
	//
	// The default will send the response to the peer if the error is a HTTPError.
	// Or only log it. So the handler and the middleware return a HTTPError,
	// instead of sending the response to the peer.
	HandleError func(Context, error)

	// You can appoint the NotFound handler. The default is NotFoundHandler().
	NotFoundHandler Handler

	// OPTIONS and MethodNotAllowed handler, which are used for the default router.
	OptionsHandler          Handler
	MethodNotAllowedHandler Handler
}
```

```go
func main() {
    config := ship.Config{
        ...
    }
    router := ship.New(config)

    ...
}
```


## API Example

See [GoDOC](https://godoc.org/github.com/xgfone/ship).

### `Router`
#### Using `Connect`, `Get`, `Post`, `Put`, `Patch`, `Delete` and `Option`

```go
func main() {
    router := ship.New()

    router.Route("/path/get", getHandler).GET()
    router.Route("/path/put", putHandler).PUT()
    router.Route("/path/post", postHandler).POST()
    router.Route("/path/patch", patchHandler).PATCH()
    router.Route("/path/delete", deleteHandler).DELETE()
    router.Route("/path/option", optionHandler).OPTIONS()
    router.Route("/path/connect", connectHandler).CONNECT()

    http.ListenAndServe(":8080", router)
}
```

Notice: you can regitser the customized method by `Route(path string, handler Handler).Method(method ...string)`.

#### Naming route and building URL
You can name the route a name when registering the route, then you can build a URL by the name.

```go
func main() {
    router := ship.New()

    router.Route("/path/:id", func(ctx Context) error {
        fmt.Println(ctx.URL("get_url", ctx.URLParam()))
    }).Name("get_url").GET()

    http.ListenAndServe(":8080", router)
}
```

#### Add the Header and Scheme filter

```go
func main() {
    router := ship.New()

    handler := func(ctx Context) error { return nil }
    router.R("/path1", handler).Schemes("https", "wss").GET()
    router.R("/path2", handler).Headers("Content-Type", "application/json").POST()

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
    router := ship.New()

    ship.MapMethodIntoRouter(router, TestStruct{}, "/v1")

    http.ListenAndServe(":8080", router)
}
```

`ship.MapMethodIntoRouter(router, TestStruct{}, "/v1")` is equal to

```go
ts := TestStruct{}
router.Route("/v1/teststruct/get", ts.Get).Name("teststruct_get").GET()
router.Route("/v1/teststruct/update", ts.Update).Name("teststruct_update").PUT()
router.Route("/v1/teststruct/create", ts.Create).Name("teststruct_create").POST()
router.Route("/v1/teststruct/delete", ts.Delete).Name("teststruct_delete").DELETE()
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

    router.Route("/url/path", handler).GET()

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
    router := ship.New()

    // Use and Before have no interference each other.
    router.Use(middleware.Logger())
    router.Pre(RemovePathPrefix("/static"))
    router.Use(middleware.Recover())

    router.Route("/url/path", handler).GET()

    http.ListenAndServe(":8080", router)
}
```

The sub-packages [`middleware`](https://godoc.org/github.com/xgfone/ship/middleware) has implemented some middlewares as follow:

- [CSRF](https://godoc.org/github.com/xgfone/ship/middleware#CSRF)
- [Gzip](https://godoc.org/github.com/xgfone/ship/middleware#Gzip)
- [Logger](https://godoc.org/github.com/xgfone/ship/middleware#Logger)
- [Recover](https://godoc.org/github.com/xgfone/ship/middleware#Recover)
- [CleanPath](https://godoc.org/github.com/xgfone/ship/middleware#CleanPath)
- [BodyLimit](https://godoc.org/github.com/xgfone/ship/middleware#BodyLimit)
- [TokenAuth](https://godoc.org/github.com/xgfone/ship/middleware#TokenAuth)
- [ResetResponse](https://godoc.org/github.com/xgfone/ship/middleware#ResetResponse)
- [RemoveTrailingSlash](https://godoc.org/github.com/xgfone/ship/middleware#RemoveTrailingSlash)

#### Using `SubRouter`

```go
func main() {
    router := ship.New()

    router.Use(middleware.Logger())
    router.Use(middleware.Recover())

    // v1 SubRouter, which will inherit the middlewares of the parent router.
    v1 := router.Group("/v1")
    v1.Route("/get/path", getHandler).GET()

    // v2 SubRouter, which won't inherit the middlewares of the parent router.
    v2 := router.GroupNone("/v2")
    v2.Use(MyAuthMiddleware())
    v2.Route("/post/path", postHandler).POST()

    http.ListenAndServe(":8080", router)
}
```

#### Traverse the registered route

```go
func main() {
    router := ship.New()

    router.Route("/get/path", getHandler).Name("get_name").GET()
    router.Route("/post/path", posttHandler).Name("post_name").POST()

    router.Traverse(func(name, method, path string) {
        fmt.Println(name, method, path)
        // Output:
        // get_name GET /get/path
        // post_name POST /post/path
    })

    http.ListenAndServe(":8080", router)
}
```


## `Context`

See [the interface doc](https://godoc.org/github.com/xgfone/ship/core#Context).


## Bind JSON, XML or Form data form payload

`ship` supply a default data binding to bind the JSON, XML or Form data from payload.

```go
type Login struct {
    Username string `json:"username" xml:"username"`
    Password string `json:"password" xml:"password"`
}

func main() {
    router := ship.New()

    router.Route("/login", func(ctx ship.Context) error {
        var login Login
        if err := ctx.Bind(&login); err != nil {
            return err
        }
        ...
    }).POST()

    http.ListenAndServe(":8080", router)
}
```


## Route Management

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route, which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://godoc.org/github.com/xgfone/ship/router/echo#NewRouter).

You can appoint yourself implementation by implementing the interface [`ship.Router`](https://godoc.org/github.com/xgfone/ship/core#Router).

```go
type Router interface {
	// Generate a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add a route with name, method , path and handler,
	// and return the number of the parameters if there are the parameters
	// in the route. Or return 0.
	//
	// If the router does not support the parameter, it should panic.
	//
	// Notice: for keeping consistent, the parameter should start with ":"
	// or "*". ":" stands for a single parameter, and "*" stands for
	// a wildcard parameter.
	Add(name string, path string, methods []string, handler Handler) (paramNum int)

	// Find a route handler by the method and path of the request.
	//
	// Return nil if the route does not exist.
	//
	// If the route has more than one parameter, the name and value
	// of the parameters should be stored `pnames` and `pvalues` respectively.
	Find(method string, path string, pnames []string, pvalues []string) (handler Handler)

	// Traverse each route.
	Each(func(name string, method string, path string))
}
```

```go
func main() {
    config := ship.Config{Router: NewMyRouter(...)}
    router := ship.New(config)
    ...
}
```

## TODO

- Add Host match for `Route`, referring to [mux.Route.Host](https://godoc.org/github.com/gorilla/mux#Route.Host).
- Add Query match for `Route`, referring to [mux.Route.Queries](https://godoc.org/github.com/gorilla/mux#Route.Queries).
- Add JWT middleware.
- Add OAuth 2.0 middleware.
- Add CORS middlware.
- Add HTML template render.
- Add the serialization and deserialization middleware.
- Give the more capacity to the default binder.
- Add the `httprouter` router implementation.
