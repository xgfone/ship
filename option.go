// Copyright 2019 xgfone
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

	"github.com/xgfone/go-tools/v6/pools"
)

// Option is used to configure Ship.
type Option func(*Ship)

// SetName sets the name, which is "" by default.
func SetName(name string) Option {
	return func(s *Ship) {
		s.name = name
	}
}

// SetDebug sets whether to enable the debug mode, which is false by default.
func SetDebug(debug bool) Option {
	return func(s *Ship) {
		s.debug = debug
	}
}

// SetPrefix sets the route prefix, which is "" by default.
func SetPrefix(prefix string) Option {
	return func(s *Ship) {
		s.prefix = strings.TrimSuffix(prefix, "/")
	}
}

// SetLogger sets the Logger, which is `NewNoLevelLogger(os.Stderr)` by default.
func SetLogger(log Logger) Option {
	return func(s *Ship) {
		if log != nil {
			s.logger = log
		}
	}
}

// SetBinder sets the Binder, which is used to bind a value from the request.
//
// It's `MuxBinder` by default, which has added some binders, for example,
// "application/json", "application/xml", "text/xml", "multipart/form-data",
// and "application/x-www-form-urlencoded" as follow.
//
//     mb := New().MuxBinder()  // or NewMuxBinder(), but you need to set it.
//     mb.Add(MIMEApplicationJSON, JSONBinder())
//     mb.Add(MIMETextXML, XMLBinder())
//     mb.Add(MIMEApplicationXML, XMLBinder())
//     mb.Add(MIMEMultipartForm, FormBinder())
//     mb.Add(MIMEApplicationForm, FormBinder())
//
// Then, you can add yourself binder, for example,
//
//     mb.Add("Content-Type", binder)
//
// So you can use it by the four ways:
//
//     mb.Bind(ctx, ptr)
//
// In the context, you can call it like this:
//
//     ctx.Bind(ptr)
//
func SetBinder(b Binder) Option {
	return func(s *Ship) {
		if b != nil {
			s.binder = b
		}
	}
}

// SetSession sets the Session, which is `NewMemorySession()` by default.
func SetSession(session Session) Option {
	return func(s *Ship) {
		if session != nil {
			s.session = session
		}
	}
}

// SetRenderer sets the Renderer, which is used to render the response to the peer.
//
// It's `MuxRenderer` by default, which has added some renderers, such as json,
// jsonpretty, xml, and xmlpretty as follow.
//
//     mr := New().MuxRenderer()  // or NewMuxRenderer(), but you need to set it.
//     mr.Add("json", JSONRenderer())
//     mr.Add("jsonpretty", JSONPrettyRenderer("    "))
//     mr.Add("xml", XMLRenderer())
//     mr.Add("xmlpretty", XMLPrettyRenderer("    "))
//
// Then, you can add yourself renderer, for example,
//
//     engine := django.New(".", ".html"))
//     mr.Add(engine.Ext(), HTMLTemplate(engine))
//
// So you can use it by the four ways:
//
//     mr.Render(ctx, "json", 200, data)
//     mr.Render(ctx, "jsonpretty", 200, data)
//     mr.Render(ctx, "xml", 200, data)
//     mr.Render(ctx, "xmlpretty", 200, data)
//     mr.Render(ctx, "index.html", 200, data)
//
// In the context, you can call it like this:
//
//     ctx.Render("json", 200, data)
//     ctx.Render("jsonpretty", 200, data)
//     ctx.Render("xml", 200, data)
//     ctx.Render("xmlpretty", 200, data)
//     ctx.Render("index.html", 200, data)
//
func SetRenderer(r Renderer) Option {
	return func(s *Ship) {
		if r != nil {
			s.renderer = r
		}
	}
}

// SetSignal sets the signals.
//
// Notice: the signals will be wrapped and handled by the http server
// if running it.
//
// The default is
//
//   []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGINT}
//
// In order to disable the signals, you can set it to []os.Signal{}.
func SetSignal(sigs []os.Signal) Option {
	return func(s *Ship) {
		s.signals = sigs
	}
}

// SetBufferSize sets the buffer size, which is used to initializing the buffer pool.
//
// The default is 2048.
func SetBufferSize(size int) Option {
	return func(s *Ship) {
		if size > 0 {
			s.bufferSize = size
			s.bufpool = pools.NewBufferPool(size)
		}
	}
}

// SetCtxDataSize sets the initializing size of the context data, that's, Context.Data.
func SetCtxDataSize(size int) Option {
	return func(s *Ship) {
		if size >= 0 {
			s.ctxDataSize = size
		}
	}
}

// SetMaxMiddlewareNum sets the maximum number of the middlewares,
// which is 256 by default.
func SetMaxMiddlewareNum(num int) Option {
	return func(s *Ship) {
		if num >= 0 {
			s.middlewareMaxNum = num
		}
	}
}

// EnableCtxHTTPContext sets whether to inject the Context into the HTTP request
// as the http context, then you can use `GetContext(httpReq)` to get the Context.
func EnableCtxHTTPContext(enable bool) Option {
	return func(s *Ship) {
		s.enableCtxHTTPContext = enable
	}
}

// SetEnableCtxHTTPContext is the alias of EnableCtxHTTPContext.
//
// DEPRECATED! In order to keep the backward compatibility, it does not removed
// until the next major version.
func SetEnableCtxHTTPContext(enable bool) Option {
	return EnableCtxHTTPContext(enable)
}

// KeepTrailingSlashPath sets whether to remove the trailing slash
// from the registered url path.
func KeepTrailingSlashPath(keep bool) Option {
	return func(s *Ship) {
		s.keepTrailingSlashPath = keep
	}
}

// SetKeepTrailingSlashPath is the alias of KeepTrailingSlashPath.
//
// DEPRECATED! In order to keep the backward compatibility, it does not removed
// until the next major version.
func SetKeepTrailingSlashPath(keep bool) Option {
	return KeepTrailingSlashPath(keep)
}

// SetDefaultMethodMapping sets the default mapping when registering the struct route.
//
// The default is
//
//     map[string]string{
//         "Create": "POST",
//         "Delete": "DELETE",
//         "Update": "PUT",
//         "Get":    "GET",
//     }
//
func SetDefaultMethodMapping(m map[string]string) Option {
	return func(s *Ship) { s.defaultMethodMapping = m }
}

// SetNotFoundHandler sets the handler to handle NotFound,
// which is `NotFoundHandler()` by default.
func SetNotFoundHandler(h Handler) Option {
	return func(s *Ship) {
		if h != nil {
			s.notFoundHandler = h
		}
	}
}

// SetOptionsHandler sets the handler to handle the OPTIONS request.
//
// It is used by the default router implemetation.
func SetOptionsHandler(h Handler) Option {
	return func(s *Ship) {
		if h != nil {
			s.optionsHandler = h
			if s.isDefaultRouter {
				s.router = s.newRouter()
			}
		}
	}
}

// SetMethodNotAllowedHandler sets the handler to handle MethodNotAllowed.
//
// It is used by the default router implemetation.
func SetMethodNotAllowedHandler(h Handler) Option {
	return func(s *Ship) {
		if h != nil {
			s.methodNotAllowedHandler = h
			if s.isDefaultRouter {
				s.router = s.newRouter()
			}
		}
	}
}

// SetNewRouter sets the NewRouter to generate the new router.
//
// if newRouter is nil, it will be reset to the default.
func SetNewRouter(newRouter func() Router) Option {
	return func(s *Ship) {
		if newRouter != nil {
			s.newRouter = newRouter
			s.isDefaultRouter = false
		} else {
			s.newRouter = s.defaultNewRouter
			s.isDefaultRouter = true
		}
		s.router = s.newRouter()
	}
}

// SetNewCtxData sets the creator of the request data to correlating to
// the context when newing the context.
//
// Notice: The lifecycle of the request data is the same as its context.
// When the context is created, the request data is also created.
// When finishing to handle the request, it will be reset by the context,
// not destroyed.
func SetNewCtxData(newCtxData func(*Context) Resetter) Option {
	return func(s *Ship) {
		s.newCtxData = newCtxData
	}
}

// SetCtxHandler sets the default global context handler.
func SetCtxHandler(handler func(*Context, ...interface{}) error) Option {
	return func(s *Ship) {
		s.ctxHandler = handler
	}
}

// SetErrorHandler sets default error handler, which will handle the error
// returned by the handler or the middleware at last.
//
// The default will send the response to the peer if the error is a HTTPError.
// Or only send 500 if no response.
func SetErrorHandler(handler func(*Context, error)) Option {
	return func(s *Ship) {
		if handler != nil {
			s.handleError = handler
		}
	}
}

// DisableErrorLog disables the default error handler to log the error.
func DisableErrorLog(disabled bool) Option {
	return func(s *Ship) {
		s.disableErrorLog = disabled
	}
}

// SetBindQuery sets the query binder to bind the query to a value,
// which is `BindURLValues(v, d, "query")` by default.
func SetBindQuery(bind func(interface{}, url.Values) error) Option {
	return func(s *Ship) {
		if bind != nil {
			s.bindQuery = bind
		}
	}
}

// SetRouteFilter sets the route filter, which will ignore the route and
// not register it if the filter returns false.
//
// For matching the group, you maybe check whether the path has the prefix,
// that's, the group name.
func SetRouteFilter(filter func(name, path, method string) bool) Option {
	return func(s *Ship) {
		if filter != nil {
			s.filter = filter
		}
	}
}

// SetRouteModifier sets the route modifier, which will modify the route
// before registering it.
//
// The modifier maybe return the new name, path and method.
//
// Notice: the modifier will be run before filter.
func SetRouteModifier(modifier func(name, path, method string) (string, string, string)) Option {
	return func(s *Ship) {
		if modifier != nil {
			s.modifier = modifier
		}
	}
}
