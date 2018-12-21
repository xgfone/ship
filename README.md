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
	router.Route("/ping").GET(func(ctx ship.Context) error {
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
	// The route prefix, which is "" by default.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool

	// If true, it won't remove the trailing slash from the registered url path.
	KeepTrailingSlashPath bool

	// It is the default mapping to map the method into router. The default is
	//
	//     map[string]string{
	//         "Create": "POST",
	//         "Delete": "DELETE",
	//         "Update": "PUT",
	//         "Get":    "GET",
	//     }
	DefaultMethodMapping map[string]string

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

	// Create a new router, which uses echo implementation by default.
	// But you can appoint yourself customized Router implementation.
	NewRouter func() Router

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

    router.Route("/path/get").GET(getHandler)
    router.Route("/path/put").PUT(putHandler)
    router.Route("/path/post").POST(postHandler)
    router.Route("/path/patch").PATCH(patchHandler)
    router.Route("/path/delete").DELETE(deleteHandler)
    router.Route("/path/option").OPTIONS(optionHandler)
    router.Route("/path/connect").CONNECT(connectHandler)

    http.ListenAndServe(":8080", router)
}
```

Notice: you can regitser the same handler with more than one method by `Route(path string).Method(handler Handler, method ...string)`.

`R` is the alias of `Route`, so you can register the routes by `R(path string).Method(handler Handler, method ...string)`.

#### Cascade the registered routes

```go
func main() {
    router := ship.New()
    router.R("/path/to").GET(getHandler).POST(postHandler).DELETE(deleteHandler)

    http.ListenAndServe(":8080", router)
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

    http.ListenAndServe(":8080", router)
}
```

#### Naming route and building URL
You can name the route a name when registering the route, then you can build a URL by the name.

```go
func main() {
    router := ship.New()

    router.Route("/path/:id").Name("get_url").GET(func(ctx Context) error {
        fmt.Println(ctx.URL("get_url", ctx.URLParamValues()))
    })

    http.ListenAndServe(":8080", router)
}
```

#### Add the Header and Scheme filter

```go
func main() {
    router := ship.New()

    handler := func(ctx Context) error { return nil }
    router.R("/path1").Schemes("https", "wss").GET(handler)
    router.R("/path2").Headers("Content-Type", "application/json").POST(handler)

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

type TestType struct{}

func (t TestType) Create(ctx Context) error { return nil }
func (t TestType) Delete(ctx Context) error { return nil }
func (t TestType) Update(ctx Context) error { return nil }
func (t TestType) Get(ctx Context) error    { return nil }
func (t TestType) Has(ctx Context) error    { return nil }
func (t TestType) NotHandler()              {}

func main() {
    router := ship.New()

    router.Route("/v1").MapType(TestType{})

    http.ListenAndServe(":8080", router)
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

The default mapping method is `Ship.Config.DefaultMethodMapping`, which is initialized as follow if not given when creating a new router `Ship`.
```go
Ship.Config.DefaultMethodMapping = map[string]string{
    // "MethodName": "RequestMethod"
    "Create": "POST",
    "Delete": "DELETE",
    "Update": "PUT",
    "Get":    "GET",
}
```

If the default is not what you want, you can customize it, for example,
```go
router := ship.New(ship.Config{
    DefaultMethodMapping: map[string]string{
        "GetMethod": "GET",
        "PostMethod": "POST",
    },
})
```

**Notice:**
- The name of type and method will be converted to the lower.
- The mapping format of the route path is `%{prefix}/%{lower_type_name}/%{lower_method_name}`.
- The mapping format of the route name is `%{lower_type_name}_%{lower_method_name}`.
- The type of the method must be `func(ship.Context) error`, or it will be ignored.

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

    router.Route("/url/path").GET(handler)

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

    router.Route("/url/path").GET(handler)

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
    v1.Route("/get/path").GET(getHandler)

    // v2 SubRouter, which won't inherit the middlewares of the parent router.
    v2 := router.GroupNone("/v2")
    v2.Use(MyAuthMiddleware())
    v2.Route("/post/path").POST(postHandler)

    http.ListenAndServe(":8080", router)
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

    http.ListenAndServe(":8080", router)
}
```

#### Add the Virtual Host

```go
func main() {
    router := ship.New()
    router.Route("/router").GET(func(c ship.Context) error { return c.String(200, "default") })

    vhost1 := router.VHost("host1.example.com")
    vhost1.Route("/router").GET(func(c ship.Context) error { return c.String(200, "vhost1") })

    vhost2 := router.VHost("host2.example.com")
    vhost2.Route("/router").GET(func(c ship.Context) error { return c.String(200, "vhost2") })

    http.ListenAndServe(":8080", router)
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

    router.Route("/login").POST(func(ctx ship.Context) error {
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

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route, which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://godoc.org/github.com/xgfone/ship/router/echo#NewRouter).

You can appoint yourself implementation by implementing the interface [`ship.Router`](https://godoc.org/github.com/xgfone/ship/core#Router).

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
	Add(name string, path string, method string, handler Handler) (paramNum int)

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

- [x] Add Host match for `Route`, referring to [mux.Route.Host](https://godoc.org/github.com/gorilla/mux#Route.Host). We use the **Virtual Host** instead.
- [x] Add Query match for `Route`, referring to [mux.Route.Queries](https://godoc.org/github.com/gorilla/mux#Route.Queries). We use `Matcher` to operate it.
- [ ] Add JWT middleware.
- [ ] Add OAuth 2.0 middleware.
- [ ] Add CORS middlware.
- [ ] Add HTML template render.
- [ ] Add the serialization and deserialization middleware.
- [ ] Give the more capacity to the default binder.
- [ ] Add the `httprouter` router implementation.
