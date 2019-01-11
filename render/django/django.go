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

package django

import (
	"io"

	"github.com/flosch/pongo2"
)

// Type Aliases from pongo2.
type (
	// A Context type provides constants, variables, instances or functions
	// to a template.
	Context = pongo2.Context

	// The Error type is being used to address an error during lexing,
	// parsing or execution.
	Error = pongo2.Error

	// FilterFunction is the type filter functions must fulfil.
	FilterFunction = pongo2.FilterFunction

	// TagParser is the function signature of the tag's parser you will have to
	// implement in order to create a new tag.
	TagParser = pongo2.TagParser

	// Template is a template type.
	Template = pongo2.Template
)

// Some functions from pongo2.
var (
	// FilterExists returns true if the given filter is already registered.
	FilterExists = pongo2.FilterExists

	// RegisterFilter registers a new filter. If there's already a filter
	// with the same name, RegisterFilter will panic. You usually want
	// to call this function in the filter's init() function.
	RegisterFilter = pongo2.RegisterFilter

	// RegisterTag registers a new tag. You usually want to call this function
	// in the tag's init() function.
	RegisterTag = pongo2.RegisterTag

	// ReplaceFilter replaces an already registered filter with a new
	// implementation. Use this function with caution since it allows you
	// to change existing filter behaviour.
	ReplaceFilter = pongo2.ReplaceFilter

	// ReplaceTag replaces an already registered tag with a new implementation.
	// Use this function with caution since it allows you to change existing
	// tag behaviour.
	ReplaceTag = pongo2.ReplaceTag

	// SetAutoescape sets whether or not to escape automatically.
	SetAutoescape = pongo2.SetAutoescape
)

// Engine adapts the pongo2 engine.
type Engine struct {
	*pongo2.TemplateSet
	directory string
	extension string
}

// New returns a new django engine.
func New(dir string, extension ...string) *Engine {
	ext := ".html"
	if len(extension) > 0 {
		ext = extension[0]
	}

	tplset := pongo2.NewSet("django", pongo2.MustNewLocalFileSystemLoader(dir))
	return &Engine{
		TemplateSet: tplset,
		directory:   dir,
		extension:   ext,
	}
}

// Ext returns the file extension which this django engine is responsible to render.
func (e *Engine) Ext() string {
	return e.extension
}

// Execute renders a django template.
func (e *Engine) Execute(w io.Writer, filename string, data interface{}, metadata map[string]interface{}) error {
	tpl, err := e.FromCache(filename)
	if err != nil {
		return err
	}
	return tpl.ExecuteWriterUnbuffered(data.(map[string]interface{}), w)
}

// Load reloads all the django templates.
func (e Engine) Load() error {
	e.CleanCache()
	return nil
}
