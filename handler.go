// Copyright 2018 xgfone <xgfone@126.com>
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
	"errors"
	"net/http"

	"github.com/xgfone/ship/core"
)

// Handler is the alias of core.Handler, which stands for a request & response
// Handler.
//
// Type:
//   type Handler func(*Context) error
type Handler = core.Handler

// Middleware is the alias of core.Middleware.
//
// Type:
//   type Middleware func(Handler) Handler
type Middleware = core.Middleware

type httpHandlerBridge struct {
	ship    *Ship
	Handler Handler
}

func newHTTPHandlerBridge(s *Ship, h Handler) httpHandlerBridge {
	if s == nil {
		panic(errors.New("ship must not be nil"))
	}
	return httpHandlerBridge{ship: s, Handler: h}
}

func (h httpHandlerBridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.ship.AcquireContext(r, w).(*contextT)
	ctx.setShip(h.ship)
	if h.Handler == nil {
		h.ship.config.NotFoundHandler(ctx)
	} else {
		h.Handler(ctx)
	}
}

// ToHTTPHandler converts the Handler to http.Handler
func ToHTTPHandler(s *Ship, h Handler) http.Handler {
	return newHTTPHandlerBridge(s, h)
}

// FromHTTPHandler converts http.Handler to Handler.
func FromHTTPHandler(h http.Handler) Handler {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

// FromHTTPHandlerFunc converts http.HandlerFunc to Handler.
func FromHTTPHandlerFunc(h http.HandlerFunc) Handler {
	return func(ctx Context) error {
		h(ctx.Response(), ctx.Request())
		return nil
	}
}

func nothingHandler(ctx Context) error {
	return nil
}

// NothingHandler returns a Handler doing nothing.
func NothingHandler() Handler {
	return nothingHandler
}

func okHandler(ctx Context) error {
	return ctx.String(http.StatusOK, "OK")
}

// OkHandler returns a Handler only sending the response "200 OK"
func OkHandler() Handler {
	return okHandler
}

func notFoundHandler(ctx Context) error {
	return ctx.String(http.StatusNotFound, "Not Found")
}

// NotFoundHandler returns a NotFound handler.
func NotFoundHandler() Handler {
	return notFoundHandler
}

func methodNotAllowedHandler(ctx Context) error {
	return ctx.NoContent(http.StatusMethodNotAllowed)
}

// MethodNotAllowedHandler returns a MethodNotAllowed handler.
func MethodNotAllowedHandler() Handler {
	return methodNotAllowedHandler
}

func optionsHandler(ctx Context) error {
	return ctx.NoContent(http.StatusOK)
}

// OptionsHandler returns OPTIONS handler.
func OptionsHandler() Handler {
	return optionsHandler
}
