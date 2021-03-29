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
	"errors"
	"fmt"
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
