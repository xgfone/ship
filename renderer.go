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
	"fmt"
	"path"
)

// MuxRender is a multiplexer for kinds of Renderer.
type MuxRender struct {
	renders map[string]Renderer
}

// NewMuxRender returns a new MuxRender.
func NewMuxRender() *MuxRender {
	return &MuxRender{renders: make(map[string]Renderer, 8)}
}

func (mr *MuxRender) fmtSuffix(suffix string) string {
	if s := path.Ext(suffix); s != "" {
		suffix = s[1:]
	}
	if suffix == "" || suffix == "." {
		panic(errors.New("the suffix is empty"))
	}
	return suffix
}

// Add adds a renderer with a suffix identifier.
func (mr *MuxRender) Add(suffix string, renderer Renderer) {
	if renderer == nil {
		panic(errors.New("the renderer is nil"))
	}
	suffix = mr.fmtSuffix(suffix)

	if _, ok := mr.renders[suffix]; ok {
		panic(fmt.Errorf("the renderer '%s' has been added", suffix))
	}

	mr.renders[suffix] = renderer
}

// Get returns the corresponding renderer by the suffix.
//
// Return nil if not found.
func (mr *MuxRender) Get(suffix string) Renderer {
	suffix = mr.fmtSuffix(suffix)
	return mr.renders[suffix]
}

// Del removes the corresponding renderer by the suffix.
func (mr *MuxRender) Del(suffix string) {
	suffix = mr.fmtSuffix(suffix)
	delete(mr.renders, suffix)
}

// Render implements the interface Renderer, which will get the renderer
// the name suffix then render the content.
func (mr *MuxRender) Render(ctx Context, name string, code int, data interface{}) error {
	if renderer := mr.Get(name); renderer != nil {
		return renderer.Render(ctx, name, code, data)
	}
	return fmt.Errorf("not support the renderer named '%s'", name)
}
