// Copyright 2022 xgfone
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

import "net/http"

// Middleware represents a middleware.
type Middleware func(Handler) Handler

// NewMiddleware returns a common middleware with the handler.
//
// Notice: the wrapped http.Handler has implemented the interface
//
//   type interface {
//       HandleHTTP(http.ResponseWriter, *http.Request) error
//   }
//
// So it can be used to wrap the error returned by other middleware handlers.
func NewMiddleware(handle func(http.Handler, http.ResponseWriter, *http.Request) error) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			req := c.Request()
			if ctx := req.Context(); GetContext(ctx) == nil {
				req = req.WithContext(SetContext(ctx, c))
				c.SetRequest(req)
			}
			return handle(mwHandler(next), c.Response(), req)
		}
	}
}

type mwHandler Handler

func (h mwHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleHTTP(w, r)
}

func (h mwHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	c := GetContext(r.Context())
	if r != c.Request() {
		c.SetRequest(r)
	}
	if resp, ok := w.(*Response); !ok || resp != c.Response() {
		c.SetResponse(w)
	}
	return Handler(h)(c)
}
