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
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/xgfone/ship/core"
	"github.com/xgfone/ship/utils"
)

// ErrNoHTMLTemplateEngine represents that it cannot find the HTML template engine by the extension.
var ErrNoHTMLTemplateEngine = errors.New("not found html template engine")

// HTMLTemplateEngine is the interface which all html template engines should be implemented.
type HTMLTemplateEngine interface {
	// Ext should return the final file extension which this template engine is responsible to render.
	Ext() string

	// Load or reload all the templates.
	Load() error

	// Eexecute and render a template by its filename.
	Execute(w io.Writer, filename string, data interface{}, metadata map[string]interface{}) error
}

// HTMLTemplateManager is used to manage the html template engine.
type HTMLTemplateManager struct {
	bufpool utils.BufferPool
	engines map[string]HTMLTemplateEngine
}

// NewHTMLTemplateManager returns a new NewHTMLTemplateManager.
func NewHTMLTemplateManager(cacheSize ...int) *HTMLTemplateManager {
	size := 8192
	if len(cacheSize) > 0 && cacheSize[0] > 0 {
		size = cacheSize[0]
	}

	return &HTMLTemplateManager{
		bufpool: utils.NewBufferPool(size),
		engines: make(map[string]HTMLTemplateEngine, 8),
	}
}

// Register registers a html template engine.
func (tm *HTMLTemplateManager) Register(e HTMLTemplateEngine) {
	if _, ok := tm.engines[e.Ext()]; ok {
		panic(fmt.Errorf("the html template engine '%s' has been registered", e.Ext()))
	}
	tm.engines[e.Ext()] = e
}

// Delete removes the corresponding html template engine by ext.
func (tm *HTMLTemplateManager) Delete(ext string) {
	delete(tm.engines, ext)
}

// GetAllTemplateEninges returns all the html template engines.
func (tm *HTMLTemplateManager) GetAllTemplateEninges() []HTMLTemplateEngine {
	tes := make([]HTMLTemplateEngine, 0, len(tm.engines))
	for _, e := range tm.engines {
		tes = append(tes, e)
	}
	return tes
}

// Render implements the interface core.Renderer.
func (tm *HTMLTemplateManager) Render(ctx core.Context, name string, code int, data interface{}) (err error) {
	buf := tm.bufpool.Get()
	if err = tm.Execute(buf, name, data, ctx.Store()); err == nil {
		err = ctx.HTMLBlob(code, buf.Bytes())
	}
	tm.bufpool.Put(buf)
	return
}

// Find returns the html template engine by the a filename extension.
func (tm *HTMLTemplateManager) Find(filename string) HTMLTemplateEngine {
	return tm.engines[filepath.Ext(filename)]
}

// Execute renders the html template and returns the result.
func (tm *HTMLTemplateManager) Execute(w io.Writer, filename string, data interface{}, metadata map[string]interface{}) error {
	if engine := tm.Find(filename); engine != nil {
		return engine.Execute(w, filename, data, metadata)
	}
	return ErrNoHTMLTemplateEngine
}

// Load (re)loads all the templates.
func (tm *HTMLTemplateManager) Load() error {
	for ext, engine := range tm.engines {
		if err := engine.Load(); err != nil {
			return fmt.Errorf("the %s template engine failed to load: %s", ext, err)
		}
	}
	return nil
}
