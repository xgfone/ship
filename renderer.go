// Copyright 2018 xgfone
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
	"fmt"
	"io"
	"path"

	"github.com/xgfone/ship/utils"
)

// Renderer is the interface to render the response.
type Renderer interface {
	Render(ctx *Context, name string, code int, data interface{}) error
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
		panic(errors.New("the suffix is empty"))
	}
	return suffix
}

// Add adds a renderer with a suffix identifier.
func (mr *MuxRenderer) Add(suffix string, renderer Renderer) {
	if renderer == nil {
		panic(errors.New("the renderer is nil"))
	}
	suffix = mr.fmtSuffix(suffix)
	mr.renders[suffix] = renderer
}

// Get returns the corresponding renderer by the suffix.
//
// Return nil if not found.
func (mr *MuxRenderer) Get(suffix string) Renderer {
	suffix = mr.fmtSuffix(suffix)
	return mr.renders[suffix]
}

// Del removes the corresponding renderer by the suffix.
func (mr *MuxRenderer) Del(suffix string) {
	suffix = mr.fmtSuffix(suffix)
	delete(mr.renders, suffix)
}

// Render implements the interface Renderer, which will get the renderer
// the name suffix then render the content.
func (mr *MuxRenderer) Render(ctx *Context, name string, code int, data interface{}) error {
	if renderer := mr.Get(name); renderer != nil {
		return renderer.Render(ctx, name, code, data)
	}
	return fmt.Errorf("not support the renderer named '%s'", name)
}

type rendererFunc func(*Context, string, int, interface{}) error

func (f rendererFunc) Render(ctx *Context, name string, code int, data interface{}) error {
	return f(ctx, name, code, data)
}

// RendererFunc converts a function to Renderer.
func RendererFunc(f func(ctx *Context, name string, code int, v interface{}) error) Renderer {
	return rendererFunc(f)
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
func SimpleRenderer(name string, contentType string, marshaler Marshaler) Renderer {
	return RendererFunc(func(ctx *Context, _name string, code int, v interface{}) error {
		if name != _name {
			panic(fmt.Errorf("the renderer name '%s' is not '%s'", _name, name))
		}
		b, err := marshaler(v)
		if err != nil {
			return err
		}
		return ctx.Blob(code, contentType, b)
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
	encode := json.Marshal
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("json", MIMEApplicationJSONCharsetUTF8, encode)
}

// JSONPrettyRenderer returns a pretty JSON renderer.
//
// Example
//
//     renderer := JSONPrettyRenderer("    ")
//     renderer.Render(ctx, "jsonpretty", code, data)
//
//     # Or appoint a specific Marshaler.
//     renderer := JSONPrettyRenderer("", func(v interface{}) ([]byte, error) {
//         return json.MarshalIndent(v, "", "  ")
//     })
//     renderer.Render(ctx, "jsonpretty", code, data)
//
// Notice: the renderer name must be "jsonpretty".
func JSONPrettyRenderer(indent string, marshal ...Marshaler) Renderer {
	encode := func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", indent) }
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("jsonpretty", MIMEApplicationJSONCharsetUTF8, encode)
}

// XMLRenderer returns a XML renderer.
//
// Example
//
//     renderer := XMLRenderer()
//     renderer.Render(ctx, "xml", code, data)
//
// Notice: the default marshaler won't add the XML header.
// and the renderer name must be "xml".
func XMLRenderer(marshal ...Marshaler) Renderer {
	encode := xml.Marshal
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("xml", MIMEApplicationXMLCharsetUTF8, encode)
}

// XMLPrettyRenderer returns a pretty XML renderer.
//
// Example
//
//     renderer := XMLPrettyRenderer("    ")
//     renderer.Render(ctx, "xmlpretty", code, data)
//
//     # Or appoint a specific Marshaler.
//     renderer := XMLPrettyRenderer("", func(v interface{}) ([]byte, error) {
//         return xml.MarshalIndent(v, "", "  ")
//     })
//     renderer.Render(ctx, "xmlpretty", code, data)
//
// Notice: the default marshaler won't add the XML header,
// and the renderer name must be "xmlpretty".
func XMLPrettyRenderer(indent string, marshal ...Marshaler) Renderer {
	encode := func(v interface{}) ([]byte, error) { return xml.MarshalIndent(v, "", indent) }
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("xmlpretty", MIMEApplicationXMLCharsetUTF8, encode)
}

// HTMLTemplateEngine is the interface which all html template engines should
// be implemented.
type HTMLTemplateEngine interface {
	// Ext should return the final file extension which this template engine
	// is responsible to render.
	Ext() string

	// Load or reload all the templates.
	Load() error

	// Eexecute and render a template by its filename.
	Execute(w io.Writer, filename string, data interface{}, metadata map[string]interface{}) error
}

var htmlTemplatePool = utils.NewBufferPool(1024 * 32)

// HTMLTemplateRenderer returns HTML template renderer.
func HTMLTemplateRenderer(engine HTMLTemplateEngine) Renderer {
	if err := engine.Load(); err != nil {
		panic(err)
	}

	return RendererFunc(func(ctx *Context, name string, code int, v interface{}) (err error) {
		buf := htmlTemplatePool.Get()
		if err = engine.Execute(buf, name, v, ctx.Data); err == nil {
			err = ctx.HTMLBlob(code, buf.Bytes())
		}
		htmlTemplatePool.Put(buf)
		return
	})
}
