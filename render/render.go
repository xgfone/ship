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

package render

import (
	"fmt"

	"github.com/xgfone/ship/core"
)

type funcRenderer func(core.Context, string, int, interface{}) error

func (f funcRenderer) Render(ctx core.Context, name string, code int, data interface{}) error {
	return f(ctx, name, code, data)
}

// RendererFunc converts a function to Renderer.
func RendererFunc(f func(ctx core.Context, name string, code int, v interface{}) error) core.Renderer {
	return funcRenderer(f)
}

// Marshaler is used to marshal a value to []byte.
type Marshaler func(data interface{}) ([]byte, error)

// SimpleRenderer returns a simple renderer, which is the same as follow:
//
//     b, err := encode(data)
//     if err != nil {
//         return err
//     }
//     return ctx.Blob(code, contentType, b)
//
func SimpleRenderer(name string, contentType string, marshaler Marshaler) core.Renderer {
	return RendererFunc(func(ctx core.Context, _name string, code int, v interface{}) error {
		if name != _name {
			return fmt.Errorf("not support the renderer named '%s'", _name)
		}
		b, err := marshaler(v)
		if err != nil {
			return err
		}
		return ctx.Blob(code, contentType, b)
	})
}
