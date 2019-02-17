// Copyright 2019 xgfone <xgfone@126.com>
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

package ship

import (
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/xgfone/ship/binder"
	"github.com/xgfone/ship/render"
	"github.com/xgfone/ship/router/echo"
	"github.com/xgfone/ship/session/memory"
)

// Config is used to configure the router used by the default implementation.
type Config struct {
	// The name of the router, which is used when starting the http server.
	Name string

	// The route prefix, which is "" by default.
	Prefix string

	// If ture, it will enable the debug mode.
	Debug bool

	// If true, it won't remove the trailing slash from the registered url path.
	KeepTrailingSlashPath bool

	// The size of the buffer initialized by the buffer pool.
	//
	// The default is 2KB.
	BufferSize int

	// The initializing size of the store, which is a map essentially,
	// used by the context.
	//
	// The default is 0. If you use the store, such as Get(), Set(), you should
	// set it to a appropriate value.
	ContextStoreSize int

	// The maximum number of the middlewares, which is 256 by default.
	MiddlewareMaxNum int

	// It is the default mapping to map the method into router. The default is
	//
	//     map[string]string{
	//         "Create": "POST",
	//         "Delete": "DELETE",
	//         "Update": "PUT",
	//         "Get":    "GET",
	//     }
	DefaultMethodMapping map[string]string

	// The signal set that built-in http server will wrap and handle.
	// The default is
	//
	//     []os.Signal{
	//         os.Interrupt,
	//         syscall.SIGTERM,
	//         syscall.SIGQUIT,
	//         syscall.SIGABRT,
	//         syscall.SIGINT,
	//     }
	//
	// In order to disable the signals, you can set it to []os.Signal{}.
	Signals []os.Signal

	// BindQuery binds the request query to v.
	BindQuery func(queries url.Values, v interface{}) error

	// The logger management, which is `NewNoLevelLogger(os.Stdout)` by default.
	// But you can appoint yourself customized Logger implementation.
	Logger Logger
	// Binder is used to bind the request data to the given value,
	// which is `NewBinder()` by default.
	// But you can appoint yourself customized Binder implementation
	Binder Binder
	// Rendered is used to render the response to the peer.
	//
	// The default is MuxRender, and adds some renderer, for example,
	// json, jsonpretty, xml, xmlpretty, etc, as follow.
	//
	//     renderer := NewMuxRender()
	//     renderer.Add("json", render.JSON())
	//     renderer.Add("jsonpretty", render.JSONPretty("    "))
	//     renderer.Add("xml", render.XML())
	//     renderer.Add("xmlpretty", render.XMLPretty("    "))
	//
	// So you can use it by the four ways:
	//
	//     renderer.Render(ctx, "json", 200, data)
	//     renderer.Render(ctx, "jsonpretty", 200, data)
	//     renderer.Render(ctx, "xml", 200, data)
	//     renderer.Render(ctx, "xmlpretty", 200, data)
	//
	// You can use the default, then add yourself renderer as follow.
	//
	///    router := New()
	//     mr := router.MuxRender()
	//     mr.Add("html", HtmlRenderer)
	//
	Renderer Renderer
	// Session is used to acquire and store the session information.
	Session Session

	// Create a new router, which uses echo implementation by default.
	// But you can appoint yourself customized Router implementation.
	NewRouter func() Router

	// Handle the error at last.
	//
	// The default will send the response to the peer if the error is a HTTPError.
	// Or only log it. So the handler and the middleware return a HTTPError,
	// instead of sending the response to the peer.
	HandleError func(Context, error)

	// NewCtxData news a value to correlate to the context when newing the context.
	NewCtxData func(Context) Resetter

	// The default global handler of Context.
	CtxHandler func(Context, ...interface{}) error

	// You can appoint the NotFound handler. The default is NotFoundHandler().
	NotFoundHandler Handler

	// OPTIONS and MethodNotAllowed handler, which are used for the default router.
	OptionsHandler          Handler
	MethodNotAllowedHandler Handler
}

func (c *Config) init(s *Ship) {
	c.Prefix = strings.TrimSuffix(c.Prefix, "/")

	if c.BufferSize <= 0 {
		c.BufferSize = 2048
	}

	if c.ContextStoreSize < 0 {
		c.ContextStoreSize = 0
	}

	if c.MiddlewareMaxNum <= 0 {
		c.MiddlewareMaxNum = 256
	}

	if c.DefaultMethodMapping == nil {
		c.DefaultMethodMapping = map[string]string{
			"Create": "POST",
			"Delete": "DELETE",
			"Update": "PUT",
			"Get":    "GET",
		}
	}

	if c.Signals == nil {
		c.Signals = []os.Signal{
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGABRT,
			syscall.SIGINT,
		}
	}

	if c.Logger == nil {
		c.Logger = NewNoLevelLogger(os.Stdout)
	}

	if c.NotFoundHandler == nil {
		c.NotFoundHandler = NotFoundHandler()
	}

	if c.HandleError == nil {
		c.HandleError = s.handleError
	}

	if c.Session == nil {
		c.Session = memory.NewSession()
	}

	if c.Binder == nil {
		c.Binder = binder.NewBinder()
	}

	if c.BindQuery == nil {
		c.BindQuery = binder.BindQuery
	}

	if c.Renderer == nil {
		mr := NewMuxRender()
		mr.Add("json", render.JSON())
		mr.Add("jsonpretty", render.JSONPretty("    "))
		mr.Add("xml", render.XML())
		mr.Add("xmlpretty", render.XMLPretty("    "))
		c.Renderer = mr
	}

	if c.NewRouter == nil {
		c.NewRouter = func() Router { return echo.NewRouter(c.MethodNotAllowedHandler, c.OptionsHandler) }
	}
}

// Option is used to configure Ship.
type Option func(*Config)

// SetName resets Config.Name.
func SetName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

// SetLogger resets Config.Logger.
func SetLogger(log Logger) Option {
	return func(c *Config) {
		if log != nil {
			c.Logger = log
		}
	}
}

// SetCtxData resets Config.NewCtxData.
func SetCtxData(newCtxData func(Context) Resetter) Option {
	return func(c *Config) {
		c.NewCtxData = newCtxData
	}
}

// SetSignal resets Config.Signals.
func SetSignal(sigs []os.Signal) Option {
	return func(c *Config) {
		c.Signals = sigs
	}
}

// SetSession resets Config.Session.
func SetSession(session Session) Option {
	return func(c *Config) {
		c.Session = session
	}
}

// SetRenderer resets Config.Renderer.
func SetRenderer(r Renderer) Option {
	return func(c *Config) {
		c.Renderer = r
	}
}

// SetContextStoreSize resets Config.ContextStoreSize.
func SetContextStoreSize(size int) Option {
	return func(c *Config) {
		if size >= 0 {
			c.ContextStoreSize = size
		}
	}
}

// SetCtxHandler resets Config.SetCtxHandler.
func SetCtxHandler(h func(Context, ...interface{}) error) Option {
	return func(c *Config) {
		c.CtxHandler = h
	}
}

// SetNotFoundHandler resets Config.NotFoundHandler.
func SetNotFoundHandler(h Handler) Option {
	return func(c *Config) {
		if h != nil {
			c.NotFoundHandler = h
		}
	}
}
