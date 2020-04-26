// Copyright 2020 xgfone
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

package template

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xgfone/ship/v2/render"
)

// File represents a template file.
type File interface {
	Name() string
	Data() []byte
	Ext() string
}

type tmplFile struct {
	name string
	data []byte
	ext  string
}

// NewFile returns a File based on the given information.
func NewFile(name, ext string, data []byte) File { return tmplFile{name, data, ext} }
func (f tmplFile) Name() string                  { return f.name }
func (f tmplFile) Data() []byte                  { return f.data }
func (f tmplFile) Ext() string                   { return f.ext }

// Loader is used to load the template file from the disk.
type Loader interface {
	// Load reloads and returns the information and content of the file
	// identified by the name.
	//
	// If the name does not exist, return (nil, nil).
	Load(name string) (File, error)

	// LoadAll reloads returns the information and content of all the files.
	LoadAll() ([]File, error)
}

// FileFilter is used to filter the filepath if it returns true.
type FileFilter func(filepath string) bool

// NewDirLoader is the same as NewDirLoaderWithFilter, not filter any files.
func NewDirLoader(dirs ...string) Loader {
	return NewDirLoaderWithFilter(func(s string) bool { return false }, dirs...)
}

// NewDirLoaderWithFilter returns a new Loader to load the files below the dirs.
//
// Notice: the name of the template file is stripped with the prefix dir.
func NewDirLoaderWithFilter(filter FileFilter, dirs ...string) Loader {
	if filter == nil {
		panic("NewDirLoaderWithFilter: filter must not be nil")
	} else if len(dirs) == 0 {
		panic("NewDirLoaderWithFilter: no dirs")
	}

	_dirs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		if len(dir) > 0 && dir[len(dir)-1] != os.PathSeparator {
			_dirs = append(_dirs, dir+string(os.PathSeparator))
		}
	}

	return tmplLoader{dirs: _dirs, filter: filter}
}

type tmplLoader struct {
	dirs   []string
	filter FileFilter
}

func (l tmplLoader) Load(name string) (File, error) {
	for _, dir := range l.dirs {
		filename := filepath.Join(dir, name)
		if _, err := os.Stat(filename); err != nil {
			return l.loadFile(dir, filename)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, nil
}

func (l tmplLoader) LoadAll() (files []File, err error) {
	for _, dir := range l.dirs {
		err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if fi.IsDir() {
				return nil
			} else if file, err := l.loadFile(dir, path); err != nil {
				return err
			} else if file != nil {
				files = append(files, file)
			}
			return nil
		})

		if err != nil {
			return
		}
	}
	return
}

func (l tmplLoader) loadFile(prefix, filename string) (File, error) {
	if l.filter(filename) {
		return nil, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	name := strings.TrimPrefix(filename, prefix)
	ext := filepath.Ext(name)
	// name = strings.TrimSuffix(name, ext)
	return NewFile(name, ext, data), nil
}

// NewHTMLRender returns a new Renderer to render the html template.
//
// If debug is true, it will reload all the templates automatically
// each time the template is rendered.
//
// The returned Renderer has a method `Reload() error` to reload
// all the templates, and you can use as follow:
//
//    htmlRender, _ := NewHTMLRender(loader, false)
//    err := htmlRender.(interface{ Reload() error }).Reload()
//    // ...
//
func NewHTMLRender(loader Loader, debug bool) (render.Renderer, error) {
	return newHTMLRender(loader, debug, false)
}

// NewSafeHTMLRender is the same as NewHTMLRender, but you can reload
// all the templates safely and concurrently.
func NewSafeHTMLRender(loader Loader, debug bool) (render.Renderer, error) {
	return newHTMLRender(loader, debug, true)
}

func newHTMLRender(loader Loader, debug, safe bool) (render.Renderer, error) {
	r := &htmlRender{
		loader: loader,
		bufs:   sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}
	if safe {
		r.lock = new(sync.RWMutex)
	}

	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

// HTMLRender is used to render a html/template.
type htmlRender struct {
	loader Loader
	debug  bool
	lock   *sync.RWMutex
	tmpl   *template.Template
	bufs   sync.Pool
}

// Reload reloads all the templates.
func (r *htmlRender) Reload() error {
	return r.reload()
}

func (r *htmlRender) reload() error {
	files, err := r.loader.LoadAll()
	if err != nil {
		return err
	}

	if r.lock != nil {
		r.lock.Lock()
		defer r.lock.Unlock()
	}

	tmpl := template.New("__DEFAULT_HTML_TEMPLATE__")
	for _, file := range files {
		t := tmpl.New(file.Name())
		if _, err = t.Parse(string(file.Data())); err != nil {
			return err
		}
	}
	r.tmpl = tmpl
	return nil
}

func (r *htmlRender) execute(w io.Writer, name string, data interface{}) error {
	if r.lock != nil {
		r.lock.RLock()
		defer r.lock.RUnlock()
	}
	return r.tmpl.ExecuteTemplate(w, name, data)
}

func (r *htmlRender) Render(w http.ResponseWriter, name string, code int,
	data interface{}) (err error) {
	if r.debug {
		if err = r.reload(); err != nil {
			return
		}
	}

	buf := r.bufs.Get().(*bytes.Buffer)
	if err = r.execute(buf, name, data); err == nil {
		if b, ok := w.(interface{ HTMLBlob(int, []byte) error }); ok {
			err = b.HTMLBlob(code, buf.Bytes())
		} else {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.WriteHeader(code)
			_, err = w.Write(buf.Bytes())
		}

	}
	buf.Reset()
	r.bufs.Put(buf)

	return
}
