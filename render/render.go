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

package render

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
)

// Renderer is the interface to render the response.
type Renderer interface {
	Render(w http.ResponseWriter, name string, code int, data interface{}) error
}

type rendererFunc func(http.ResponseWriter, string, int, interface{}) error

func (f rendererFunc) Render(w http.ResponseWriter, name string, code int, data interface{}) error {
	return f(w, name, code, data)
}

// RendererFunc converts a function to Renderer.
func RendererFunc(f func(http.ResponseWriter, string, int, interface{}) error) Renderer {
	return rendererFunc(f)
}

// MuxRenderer is a multiplexer for kinds of Renderers.
type MuxRenderer struct {
	renders map[string]Renderer
}

// NewMuxRenderer returns a new MuxRenderer.
func NewMuxRenderer() *MuxRenderer {
	return &MuxRenderer{renders: make(map[string]Renderer, 8)}
}

func (mr *MuxRenderer) fmtSuffix(suffix string) string {
	if s := path.Ext(suffix); s != "" {
		suffix = s[1:]
	}
	if suffix == "" || suffix == "." {
		panic(errors.New("MuxRenderer: the suffix is empty"))
	}
	return suffix
}

// Add adds a renderer with a suffix identifier.
func (mr *MuxRenderer) Add(suffix string, renderer Renderer) {
	mr.renders[mr.fmtSuffix(suffix)] = renderer
}

// Get returns the corresponding renderer by the suffix.
//
// Return nil if not found.
func (mr *MuxRenderer) Get(suffix string) Renderer {
	return mr.renders[mr.fmtSuffix(suffix)]
}

// Del removes the corresponding renderer by the suffix.
func (mr *MuxRenderer) Del(suffix string) {
	delete(mr.renders, mr.fmtSuffix(suffix))
}

// Render implements the interface Renderer, which will get the renderer
// the name suffix then render the content.
func (mr *MuxRenderer) Render(w http.ResponseWriter, name string, code int, data interface{}) error {
	if renderer := mr.Get(name); renderer != nil {
		return renderer.Render(w, name, code, data)
	}
	return fmt.Errorf("not support the renderer named '%s'", name)
}

// Marshaler is used to marshal a value to []byte.
type Marshaler func(w io.Writer, data interface{}) error

// SimpleRenderer returns a simple renderer, which is the same as follow:
//
//     resp.WriteHeader(code)
//     return marshaler(w, v)
//
func SimpleRenderer(name string, contentType string, marshaler Marshaler) Renderer {
	return RendererFunc(func(w http.ResponseWriter, n string, c int, v interface{}) error {
		if name != n {
			panic(fmt.Errorf("the renderer name '%s' is not '%s'", n, name))
		}
		w.WriteHeader(c)
		return marshaler(w, v)
	})
}

// JSONRenderer returns a JSON renderer.
//
// Example
//
//     renderer := JSONRenderer()
//     renderer.Render(ctx, "json", code, data)
//
// Notice: the renderer name must be "json".
func JSONRenderer(marshal ...Marshaler) Renderer {
	e := func(w io.Writer, v interface{}) error { return json.NewEncoder(w).Encode(v) }
	if len(marshal) > 0 && marshal[0] != nil {
		e = marshal[0]
	}
	return SimpleRenderer("json", "application/json; charset=UTF-8", e)
}

// JSONPrettyRenderer returns a pretty JSON renderer.
//
// Example
//
//     renderer := JSONPrettyRenderer()
//     renderer.Render(ctx, "jsonpretty", code, data)
//
// Notice: the renderer name must be "jsonpretty".
func JSONPrettyRenderer(marshal ...Marshaler) Renderer {
	e := func(w io.Writer, v interface{}) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		return enc.Encode(v)
	}
	if len(marshal) > 0 && marshal[0] != nil {
		e = marshal[0]
	}
	return SimpleRenderer("jsonpretty", "application/json; charset=UTF-8", e)
}

// XMLRenderer returns a XML renderer.
//
// Example
//
//     renderer := XMLRenderer()
//     renderer.Render(ctx, "xml", code, data)
//
// Notice: the default marshaler won't add the XML header. and the renderer
// name must be "xml".
func XMLRenderer(marshal ...Marshaler) Renderer {
	e := func(w io.Writer, v interface{}) error { return xml.NewEncoder(w).Encode(v) }
	if len(marshal) > 0 && marshal[0] != nil {
		e = marshal[0]
	}
	return SimpleRenderer("xml", "application/xml; charset=UTF-8", e)
}

// XMLPrettyRenderer returns a pretty XML renderer.
//
// Example
//
//     renderer := XMLPrettyRenderer()
//     renderer.Render(ctx, "xmlpretty", code, data)
//
// Notice: the default marshaler won't add the XML header, and the renderer
// name must be "xmlpretty".
func XMLPrettyRenderer(marshal ...Marshaler) Renderer {
	e := func(w io.Writer, v interface{}) error {
		enc := xml.NewEncoder(w)
		enc.Indent("", "    ")
		return enc.Encode(v)
	}
	if len(marshal) > 0 && marshal[0] != nil {
		e = marshal[0]
	}
	return SimpleRenderer("xmlpretty", "application/xml; charset=UTF-8", e)
}
