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

	"github.com/xgfone/ship/v4/binder"
)

// Binder is the interface to bind the value to v from ctx.
type Binder interface {
	// Bind parses the data from http.Request to v.
	//
	// Notice: v must be a non-nil pointer.
	Bind(req *http.Request, v interface{}) error
}

type binderFunc func(*http.Request, interface{}) error

func (f binderFunc) Bind(r *http.Request, v interface{}) error { return f(r, v) }

// BinderFunc converts a function to Binder.
func BinderFunc(f func(*http.Request, interface{}) error) Binder { return binderFunc(f) }

// MuxBinder is a multiplexer for kinds of Binders based on the Content-Type.
type MuxBinder struct {
	binders map[string]Binder
}

// NewMuxBinder returns a new MuxBinder.
func NewMuxBinder() *MuxBinder { return &MuxBinder{binders: make(map[string]Binder, 8)} }

// Add adds a binder to bind the content for the header Content-Type.
func (mb *MuxBinder) Add(contentType string, binder Binder) { mb.binders[contentType] = binder }

// Get returns the corresponding binder by the header Content-Type.
//
// Return nil if not found.
func (mb *MuxBinder) Get(contentType string) Binder { return mb.binders[contentType] }

// Del removes the corresponding binder by the header Content-Type.
func (mb *MuxBinder) Del(contentType string) { delete(mb.binders, contentType) }

// Bind implements the interface Binder, which will call the registered binder
// to bind the request to v by the request header Content-Type.
func (mb *MuxBinder) Bind(req *http.Request, v interface{}) error {
	ct := req.Header.Get("Content-Type")
	if index := strings.IndexAny(ct, ";"); index > 0 {
		ct = strings.TrimSpace(ct[:index])
	}

	if ct == "" {
		return ErrMissingContentType
	}
	if binder := mb.Get(ct); binder != nil {
		return binder.Bind(req, v)
	}
	return ErrUnsupportedMediaType.Newf("not support Content-Type '%s'", ct)
}

// JSONBinder returns a JSON binder to bind the JSON request.
func JSONBinder() Binder {
	return BinderFunc(func(r *http.Request, v interface{}) (err error) {
		if r.ContentLength > 0 {
			err = json.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// XMLBinder returns a XML binder to bind the XML request.
func XMLBinder() Binder {
	return BinderFunc(func(r *http.Request, v interface{}) (err error) {
		if r.ContentLength > 0 {
			err = xml.NewDecoder(r.Body).Decode(v)
		}
		return
	})
}

// FormBinder returns a Form binder to bind the Form request.
//
// Notice: The bound value must be a pointer to a struct with the tag
// named tag, which is "form" by default.
func FormBinder(maxMemory int64, tag ...string) Binder {
	_tag := "form"
	if len(tag) > 0 && tag[0] != "" {
		_tag = tag[0]
	}

	return BinderFunc(func(r *http.Request, v interface{}) (err error) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if err = r.ParseMultipartForm(maxMemory); err != nil {
				return
			}
		} else if err = r.ParseForm(); err != nil {
			return err
		}

		return binder.BindURLValues(v, r.Form, _tag)
	})
}

// QueryBinder returns a query binder to bind the query parameters..
//
// Notice: The bound value must be a pointer to a struct with the tag
// named tag, which is "query" by default.
func QueryBinder(tag ...string) Binder {
	_tag := "query"
	if len(tag) > 0 && tag[0] != "" {
		_tag = tag[0]
	}

	return BinderFunc(func(r *http.Request, v interface{}) error {
		return binder.BindURLValues(v, r.URL.Query(), _tag)
	})
}
