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
)

// File represents a template file.
type File interface {
	Name() string
	Data() []byte
	Ext() string
}

type file struct {
	name string
	data []byte
	ext  string
}

// NewFile returns a File based on the given information.
func NewFile(name, ext string, data []byte) File { return file{name, data, ext} }
func (f file) Name() string                      { return f.name }
func (f file) Data() []byte                      { return f.data }
func (f file) Ext() string                       { return f.ext }

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

	return loader{dirs: _dirs, filter: filter}
}

type loader struct {
	dirs   []string
	filter FileFilter
}

func (l loader) Load(name string) (File, error) {
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

func (l loader) LoadAll() (files []File, err error) {
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

func (l loader) loadFile(prefix, filename string) (File, error) {
	if l.filter(filename) {
		return nil, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	name := strings.TrimPrefix(filename, prefix)
	name = strings.Replace(name, "\\", "/", -1)
	ext := filepath.Ext(name)
	// name = strings.TrimSuffix(name, ext)
	return NewFile(name, ext, data), nil
}

// NewHTMLTemplateRender returns a new Renderer to render the html template.
func NewHTMLTemplateRender(loader Loader) *HTMLTemplateRender {
	r := &HTMLTemplateRender{
		loader: loader,
		bufs:   sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
	}
	return r
}

// HTMLTemplateRender is used to render a html/template.
type HTMLTemplateRender struct {
	loader Loader
	funcs  []template.FuncMap
	debug  bool
	right  string
	left   string

	load sync.Once
	lock *sync.RWMutex
	tmpl *template.Template
	bufs sync.Pool
}

// Debug sets the debug model and returns itself.
//
// If debug is true, it will reload all the templates automatically each time
// the template is rendered.
func (r *HTMLTemplateRender) Debug(debug bool) *HTMLTemplateRender {
	r.debug = debug
	return r
}

// Lock enables the lock to reload the templates safely and concurrently.
//
// Notice: There is no need to enable the lock when no reloading the templates
// during rendering the templates.
func (r *HTMLTemplateRender) Lock(lock bool) *HTMLTemplateRender {
	if lock && r.lock == nil {
		r.lock = new(sync.RWMutex)
	} else if !lock && r.lock != nil {
		r.lock = nil
	}
	return r
}

// Delims resets the left and right delimiter.
//
// The default delimiters are "{{" and "}}".
//
// Notice: it must be set before rendering the html template.
func (r *HTMLTemplateRender) Delims(left, right string) *HTMLTemplateRender {
	r.left = left
	r.right = right
	return r
}

// Funcs appends the FuncMap.
//
// Notice: it must be set before rendering the html template.
func (r *HTMLTemplateRender) Funcs(funcs template.FuncMap) *HTMLTemplateRender {
	r.funcs = append(r.funcs, funcs)
	return r
}

// Reload reloads all the templates.
func (r *HTMLTemplateRender) Reload() error {
	return r.reload()
}

func (r *HTMLTemplateRender) reload() error {
	files, err := r.loader.LoadAll()
	if err != nil {
		return err
	}

	tmpl := template.New("__DEFAULT_HTML_TEMPLATE__")
	tmpl.Delims(r.left, r.right)
	for _, file := range files {
		t := tmpl.New(file.Name())
		for _, funcs := range r.funcs {
			t.Funcs(funcs)
		}
		if _, err = t.Parse(string(file.Data())); err != nil {
			return err
		}
	}

	if r.lock == nil {
		r.tmpl = tmpl
	} else {
		r.lock.Lock()
		r.tmpl = tmpl
		r.lock.Unlock()
	}

	return nil
}

func (r *HTMLTemplateRender) execute(w io.Writer, name string, data interface{}) error {
	var tmpl *template.Template
	if r.lock == nil {
		tmpl = r.tmpl
	} else {
		r.lock.RLock()
		tmpl = r.tmpl
		r.lock.RUnlock()
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// Render implements the interface render.Renderer.
func (r *HTMLTemplateRender) Render(w http.ResponseWriter, name string, code int,
	data interface{}) (err error) {
	if r.debug {
		if err = r.reload(); err != nil {
			return
		}
	} else {
		r.load.Do(func() { err = r.reload() })
		if err != nil {
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
