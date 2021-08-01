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
	"net/http"
	"strings"

	"github.com/xgfone/ship/v5/binder"
)

// Binder is the interface to bind the value dst to req.
type Binder interface {
	// Bind parses the data from http.Request to dst.
	//
	// Notice: dst must be a non-nil pointer.
	Bind(dst interface{}, req *http.Request) error
}

// BinderFunc is a function type implementing the interface Binder.
type BinderFunc func(dst interface{}, req *http.Request) error

// Bind implements the interface Binder.
func (f BinderFunc) Bind(dst interface{}, req *http.Request) error {
	return f(dst, req)
}

// MuxBinder is a multiplexer for kinds of Binders based on the request header
// "Content-Type".
type MuxBinder struct {
	binders map[string]Binder
}

// NewMuxBinder returns a new MuxBinder.
func NewMuxBinder() *MuxBinder {
	return &MuxBinder{binders: make(map[string]Binder, 8)}
}

// Add adds a binder to bind the content for the header "Content-Type".
func (mb *MuxBinder) Add(contentType string, binder Binder) {
	mb.binders[contentType] = binder
}

// Get returns the corresponding binder by the header "Content-Type".
//
// Return nil if not found.
func (mb *MuxBinder) Get(contentType string) Binder {
	return mb.binders[contentType]
}

// Del removes the corresponding binder by the header "Content-Type".
func (mb *MuxBinder) Del(contentType string) {
	delete(mb.binders, contentType)
}

// Bind implements the interface Binder, which looks up the registered binder
// by the request header "Content-Type" and calls it to bind the value dst
// to req.
func (mb *MuxBinder) Bind(dst interface{}, req *http.Request) error {
	ct := req.Header.Get("Content-Type")
	if index := strings.IndexAny(ct, ";"); index > 0 {
		ct = strings.TrimSpace(ct[:index])
	}

	if ct == "" {
		return ErrMissingContentType
	}

	if binder := mb.Get(ct); binder != nil {
		return binder.Bind(dst, req)
	}

	return ErrUnsupportedMediaType.Newf("not support Content-Type '%s'", ct)
}

// JSONBinder returns a binder to bind the data to the request body as JSON.
func JSONBinder() Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		if r.ContentLength > 0 {
			err = json.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// XMLBinder returns a binder to bind the data to the request body as XML.
func XMLBinder() Binder {
	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		if r.ContentLength > 0 {
			err = xml.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// FormBinder returns a binder to bind the data to the request body as Form.
//
// Notice: The bound value must be a pointer to a struct with the tag
// named tag, which is "form" by default.
func FormBinder(maxMemory int64, tag ...string) Binder {
	_tag := "form"
	if len(tag) > 0 && tag[0] != "" {
		_tag = tag[0]
	}

	return BinderFunc(func(v interface{}, r *http.Request) (err error) {
		ct := r.Header.Get("Content-Type")

		if strings.HasPrefix(ct, MIMEMultipartForm) {
			if err = r.ParseMultipartForm(maxMemory); err != nil {
				return
			}
		} else if err = r.ParseForm(); err != nil {
			return err
		}

		return binder.BindURLValues(v, r.Form, _tag)
	})
}
