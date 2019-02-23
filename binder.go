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
	"encoding/json"
	"encoding/xml"
	"errors"
)

// Binder is the interface to bind the value to v from ctx.
type Binder interface {
	Bind(ctx *Context, v interface{}) error
}

// MuxBinder is a multiplexer for kinds of Binders.
type MuxBinder struct {
	binders map[string]Binder
}

// NewMuxBinder returns a new MuxBinder.
func NewMuxBinder() *MuxBinder {
	return &MuxBinder{binders: make(map[string]Binder, 8)}
}

// Add adds a binder to bind the content for the header Content-Type.
func (mb *MuxBinder) Add(contentType string, binder Binder) {
	if binder == nil {
		panic(errors.New("the binder is nil"))
	}
	mb.binders[contentType] = binder
}

// Get returns the corresponding binder by the header Content-Type.
//
// Return nil if not found.
func (mb *MuxBinder) Get(contentType string) Binder {
	return mb.binders[contentType]
}

// Del removes the corresponding binder by the header Content-Type.
func (mb *MuxBinder) Del(contentType string) {
	delete(mb.binders, contentType)
}

// Bind implements the interface Binder, which will call the registered binder
// to bind the request to v by the request header Content-Type.
func (mb *MuxBinder) Bind(ctx *Context, v interface{}) error {
	ct := ctx.ContentType()
	if ct == "" {
		return ErrMissingContentType
	}
	if binder := mb.Get(ct); binder != nil {
		return binder.Bind(ctx, v)
	}
	return ErrUnsupportedMediaType.NewMsg("not support Content-Type '%s'", ct)
}

type binderFunc func(*Context, interface{}) error

func (f binderFunc) Bind(ctx *Context, v interface{}) error {
	return f(ctx, v)
}

// BinderFunc converts a function to Binder.
func BinderFunc(f func(*Context, interface{}) error) Binder {
	return binderFunc(f)
}

// JSONBinder returns a JSON binder to bind the JSON request.
func JSONBinder() Binder {
	return BinderFunc(func(ctx *Context, v interface{}) error {
		return json.NewDecoder(ctx.req.Body).Decode(v)
	})
}

// XMLBinder returns a XML binder to bind the XML request.
func XMLBinder() Binder {
	return BinderFunc(func(ctx *Context, v interface{}) error {
		return xml.NewDecoder(ctx.req.Body).Decode(v)
	})
}

// FormBinder returns a Form binder to bind the Form request.
//
// Notice: The bound value must be a pointer to a struct.
// You can modify the name of the field by the tag, which is "form" by default.
func FormBinder(tag ...string) Binder {
	_tag := "form"
	if len(tag) > 0 && tag[0] != "" {
		_tag = tag[0]
	}

	return BinderFunc(func(ctx *Context, v interface{}) error {
		form, err := ctx.FormParams()
		if err != nil {
			return err
		}
		return BindURLValues(v, form, _tag)
	})
}

// QueryBinder returns a query binder to bind the query parameters..
//
// Notice: The bound value must be a pointer to a struct.
// You can modify the name of the field by the tag, which is "query" by default.
func QueryBinder(tag ...string) Binder {
	_tag := "query"
	if len(tag) > 0 && tag[0] != "" {
		_tag = tag[0]
	}

	return BinderFunc(func(ctx *Context, v interface{}) error {
		return BindURLValues(v, ctx.QueryParams(), _tag)
	})
}
