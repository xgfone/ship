# ship [![Build Status](https://github.com/xgfone/ship/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/ship/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/ship/v5)](https://pkg.go.dev/github.com/xgfone/ship/v5) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/ship/master/LICENSE)

`ship` is a flexible, powerful, high performance and minimalist Go Web HTTP router framework supporting Go `1.11+`. It is inspired by [echo](https://github.com/labstack/echo) and [httprouter](https://github.com/julienschmidt/httprouter). Thanks for those contributors.


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


### Components
- `Ship` is the pure router framework based on the method and the path, including `Middleware`, `Context`, `Router`, etc.
- `HostManager` and `HostHandler` are the vhost manager and the standard http handler with the vhost manager.
- `Runner` is the runner to start the http server with the standard http handler.


## Install

```shell
go get -u github.com/xgfone/ship/v5
```


## Quick Start

```go
// example.go
package main

import (
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

func main() {
	router := ship.New()
	router.Use(middleware.Logger(), middleware.Recover()) // Use the middlewares.

	router.Route("/ping").GET(func(c *ship.Context) error {
		return c.JSON(200, map[string]interface{}{"message": "pong"})
	})

	group := router.Group("/group")
	group.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(200, "group")
	})

	subgroup := group.Group("/subgroup")
	subgroup.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(200, "subgroup")
	})

	// Start the HTTP server.
	ship.StartServer(":8080", router)
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

$ curl http://127.0.0.1:8080/group/ping
group

$ curl http://127.0.0.1:8080/group/subgroup/ping
subgroup
```

### Route Path
The route path supports the parameters like `:paramName`, `*` or `*restParamName`.

- `/path/to/route` only matches the path `/path/to/route`.
- `/path/:param1/to` matches the path `/path/abc/to`, `/path/xyz/to`, etc. And `:param1` is equal to `abc` or `xyz`.
- `/path/:param1/to/:param2` matches the path `/path/p11/to/p21`, `/path/p12/to/p22`, etc. And `:parma1` is equal to `p11` or `p12`, and `:param2` is equal to `p12` or `p22`.
- `/path/to/*` or `/path/to/*all` matches the path `/path/to/abc`, `/path/to/abc/efg`, `/path/to/xyz`, `/path/to/xyz/123`, etc. And `*` or `*all` is equal to `abc`, `abc/efg`, `xyz`, or `xzy/123`. **Notice:** `*` or `*restParamName` must be the last one of the route path.
- `/path/:param/to/*` matches the path `/path/abc/to/efg`, `/path/abc/to/efg/123`, etc. And `:param` is equal to `abc`, and `*` is equal to `efg` or `efg/123`

For the parameter, it can be accessed by `Context.Param(paramName)`.

- For `*`, the parameter name is `*`, like `Context.Param("*")`.
- For `*restParamName`, the parameter name is `restParamName`, like `Context.Param(restParamName)`.


## API Example

### Route Builder
#### Using `CONNECT`, `HEAD`, `GET`, `POST`, `PUT`, `PATCH`, `DELETE` and `OPTIONS`

```go
func main() {
	router := ship.New()
	router.Route("/path/get").GET(getHandler)
	router.Route("/path/put").PUT(putHandler)
	router.Route("/path/head").HEAD(headHandler)
	router.Route("/path/post").POST(postHandler)
	router.Route("/path/patch").PATCH(patchHandler)
	router.Route("/path/delete").DELETE(deleteHandler)
	router.Route("/path/option").OPTIONS(optionHandler)
	router.Route("/path/connect").CONNECT(connectHandler)
	ship.StartServer(":8080", router)
}
```

Notice: you can register the same handler with more than one method by `Route(path string).Method(handler Handler, method ...string)`.


#### Cascade the registered routes

```go
func main() {
	router := ship.New()
	router.Route("/path/to").GET(getHandler).POST(postHandler).DELETE(deleteHandler)
	ship.StartServer(":8080", router)
}
```


#### Use the mapping of the route methods
```go
func main() {
	router := ship.New()
	router.Route("/path/to").Map(map[string]ship.Handler{
		"GET": getHandler,
		"POST": postHandler,
		"DELETE": deleteHandler,
	})
	ship.StartServer(":8080", router)
}
```


#### Name the route
When registering the route, it can be named with a name.

```go
func main() {
	router := ship.New()
	router.Route("/path/:id").Name("get_url").GET(func(c *ship.Context) error {
		fmt.Println(c.URL("get_url", c.Param("id")))
		return nil
	})
	ship.StartServer(":8080", router)
}
```


#### Use the route group

```go
package main

import (
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

// MyAuthMiddleware returns a middleare to authenticate the request.
func MyAuthMiddleware() ship.Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(c *ship.Context) error {
			// TODO: authenticate the request.
			return next(c)
		}
	}
}

func main() {
	router := ship.New()
	router.Use(middleware.Logger(), middleware.Recover())

	// v1 Group, which will inherit the middlewares of the parent router.
	v1 := router.Group("/v1")
	v1.Route("/get").GET(func(c *ship.Context) error { return nil }) // Route: GET /v1/get

	// v2 Group, which won't inherit the middlewares of the parent router.
	v2 := router.Group("/v2").ResetMiddlewares(MyAuthMiddleware())
	v2.Route("/post").POST(func(c *ship.Context) error { return nil }) // Route: POST /v2/post

	// For sub-group of v2 Group.
	v2g := v2.Group("/child")
	v2g.Route("/path").GET(func(c *ship.Context) error { return nil }) // Route: GET /v2/child/path

	ship.StartServer(":8080", router)
}
```

#### Filter the unacceptable routes
```go
package main

import (
	"strings"

	"github.com/xgfone/ship/v5"
)

func filter(ri ship.Route) bool {
	if ri.Name == "" || !strings.HasPrefix(ri.Path, "/prefix/") {
		return true
	}
	return false
}

func main() {
	handler := func(c *ship.Context) error { return nil }

	router := ship.New()
	router.RouteFilter = filter // Don't register the router without name.

	router.Group("/prefix").Route("/name").Name("test").GET(handler) // Register the route
	router.Group("/prefix").Route("/noname").GET(handler)            // Don't register the route
	router.Route("/no_group").GET(handler)                           // Don't register the route

	ship.StartServer(":8080", router)
}
```

#### Modify the route before registering it
```go
package main

import "github.com/xgfone/ship/v5"

func modifier(ri ship.Route) ship.Route {
	ri.Path = "/prefix" + ri.Path
	return ri
}

func main() {
	handler := func(c *ship.Context) error { return nil }

	router := ship.New()
	router.RouteModifier = modifier
	router.Route("/path").Name("test").GET(handler) // Register the path as "/prefix/path".

	ship.StartServer(":8080", router)
}
```


### Use `Middleware`
```go
package main

import (
	"fmt"
	"strings"

	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

// RemovePathPrefix returns a middleware to remove the prefix from the request path.
func RemovePathPrefix(prefix string) ship.Middleware {
	if len(prefix) < 2 || prefix[len(prefix)-1] == '/' {
		panic(fmt.Errorf("invalid prefix: '%s'", prefix))
	}

	return func(next ship.Handler) ship.Handler {
		return func(c *ship.Context) error {
			req := c.Request()
			req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
			return next(c)
		}
	}
}

func main() {
	router := ship.New()

	// Execute the middlewares before finding the route.
	router.Pre(RemovePathPrefix("/static"))

	// Execute the middlewares after finding the route.
	router.Use(middleware.Logger(), middleware.Recover())

	handler := func(c *ship.Context) error { return nil }
	router.Route("/path1").GET(handler)
	router.Route("/path2").GET(handler)
	router.Route("/path3").GET(handler)

	ship.StartServer(":8080", router)
}
```

### Use the virtual host

```go
package main

import (
	"github.com/xgfone/ship/v5"
)

func main() {
	vhosts := ship.NewHostManagerHandler(nil)

	_default := ship.New()
	_default.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "default") })
	vhosts.SetDefaultHost("", _default)

	// Exact Match Host
	vhost1 := ship.New()
	vhost1.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost1") })
	vhosts.AddHost("www.host1.example.com", vhost1)

	// Suffix Match Host
	vhost2 := ship.New()
	vhost2.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost2") })
	vhosts.AddHost("*.host2.example.com", vhost2)

	// Prefix Match Host
	vhost3 := ship.New()
	vhost3.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost3") })
	vhosts.AddHost("www.host3.*", vhost3)

	// Regexp Match Host by using Go regexp package
	vhost4 := ship.New()
	vhost4.Route("/").GET(func(c *ship.Context) error { return c.Text(200, "vhost4") })
	vhosts.AddHost(`www\.[a-zA-z0-9]+\.example\.com`, vhost4)

	ship.StartServer(":8080", vhosts)
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

import "github.com/xgfone/ship/v5"

func responder(c *ship.Context, args ...interface{}) error {
	switch len(args) {
	case 0:
		return c.NoContent(200)
	case 1:
		switch v := args[0].(type) {
		case int:
			return c.NoContent(v)
		case string:
			return c.Text(200, v)
		}
	case 2:
		switch v0 := args[0].(type) {
		case int:
			return c.Text(v0, "%v", args[1])
		}
	}
	return c.NoContent(500)
}

func main() {
	router := ship.New()
	router.Responder = responder
	router.Route("/path1").GET(func(c *ship.Context) error { return c.Respond() })
	router.Route("/path2").GET(func(c *ship.Context) error { return c.Respond(200) })
	router.Route("/path3").GET(func(c *ship.Context) error { return c.Respond("Hello, World") })
	router.Route("/path4").GET(func(c *ship.Context) error { return c.Respond(200, "Hello, World") })
	ship.StartServer(":8080", router)
}
```

### Bind JSON, XML or Form data from the request payload
```go
package main

import "github.com/xgfone/ship/v5"

// Login is the login information.
type Login struct {
	Username string `json:"username" xml:"username"`
	Password string `json:"password" xml:"password"`
}

func main() {
	router := ship.Default()
	router.Route("/login").POST(func(c *ship.Context) (err error) {
		var login Login
		if err = c.Bind(&login); err != nil {
			return ship.ErrBadRequest.New(err)
		}
		return c.Text(200, "username=%s, password=%s", login.Username, login.Password)
	})

	ship.StartServer(":8080", router)
}
```

```shell
$ curl http://127.0.0.1:8080/login \
    -H 'Content-Type: application/json' \
    -d '{"username":"xgfone","password":"123456"}'
username=xgfone, password=123456

$ curl http://127.0.0.1:8080/login \
    -H 'Content-Type: application/xml' \
    -d '<login><username>xgfone</username><password>123456</password></login>'
username=xgfone, password=123456
```

### Render HTML template

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

```go
package main

import (
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/render/template"
)

func main() {
	// It will recursively load all the files in the directory as the templates.
	loader := template.NewDirLoader("/path/to/templates")
	tmplRender := template.NewHTMLTemplateRender(loader)

	router := ship.Default()
	router.Renderer.(*ship.MuxRenderer).Add(".tmpl", tmplRender)
	router.Route("/html").GET(func(c *ship.Context) error {
		return c.RenderOk("index.tmpl", "Hello World")
	})

	// Start the HTTP server.
	ship.StartServer(":8080", router)
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


## Route Management

`ship` supply a default implementation based on [Radix tree](https://en.wikipedia.org/wiki/Radix_tree) to manage the route with **Zero Garbage** (See [Benchmark](#benchmark)), which refers to [echo](https://github.com/labstack/echo), that's, [`NewRouter()`](https://pkg.go.dev/github.com/xgfone/ship/v5/router/echo?tab=doc#NewRouter).

You can appoint your own implementation by implementing the interface [`Router`](https://pkg.go.dev/github.com/xgfone/ship/v5/router?tab=doc#Router).

```go
type Router interface {
	// Range traverses all the registered routes.
	Range(func(name, path, method string, handler interface{}))

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
	NewMyRouter := func() (router ship.Router) {
		// TODO: new a Router.
		return
	}

	router := ship.New()
	router.Router = NewMyRouter()
	// ...
}
```


## Benchmark

```
HP Laptop 14s-dr2014TU
go:     1.16.4
goos:   windows
goarch: amd64
memory: 16GB DDR4-3200
cpu:    11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
```


|            Framework          | Version
|-------------------------------|---------
| `github.com/gin-gonic/gin`    | v1.7.2
| `github.com/labstack/echo/v4` | v4.4.0
| `github.com/xgfone/ship/v5`   | v5.0.0


|           Function           |  ops   | ns/op | B/opt | allocs/op
|------------------------------|--------|-------|-------|-----------
| Benchmark**Echo**Static-8        |  43269 | 27676 |  2056 | 157
| Benchmark**Echo**GitHubAPI-8     |  29738 | 40773 |  2788 | 203
| Benchmark**Echo**GplusAPI-8      | 668731 |  1967 |   207 |  13
| Benchmark**Echo**ParseAPI-8      | 362774 |  3369 |   398 |  26
| Benchmark**Gin**Static-8         |  47384 | 24037 |  8267 | 157
| Benchmark**Gin**GitHubAPI-8      |  33747 | 34648 | 10771 | 203
| Benchmark**Gin**GplusAPI-8       | 598628 |  1830 |   681 |  13
| Benchmark**Gin**ParseAPI-8       | 356298 |  3314 |  1442 |  26
| Benchmark**ShipEcho**Static-8    |  51788 | 23219 |   668 | **0**
| Benchmark**ShipEcho**GitHubAPI-8 |  32854 | 35759 |  1054 | **0**
| Benchmark**ShipEcho**GplusAPI-8  | 746049 |  1809 |    92 | **0**
| Benchmark**ShipEcho**ParseAPI-8  | 396067 |  3310 |   174 | **0**


|             Function           |    ops   | ns/op | B/opt | allocs/op
|--------------------------------|----------|-------|-------|-----------
| BenchmarkShipWithoutVHost-8    | 19691887 | 54.53 |   0   |    0
| BenchmarkShipWithExactVHost-8  | 17158249 | 64.19 |   0   |    0
| BenchmarkShipWithPrefixVHost-8 | 13445091 | 90.81 |   0   |    0
| BenchmarkShipWithRegexpVHost-8 |  4668913 | 248.0 |   0   |    0
