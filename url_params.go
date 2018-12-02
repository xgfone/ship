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
	"context"
	"net/http"
)

// NewURLParam returns a new URLParam.
func NewURLParam(cap int) URLParam {
	return &urlParams{kvs: make([]paramT, 0, cap)}
}

type paramT struct {
	key   string
	value string
}

type urlParams struct {
	kvs []paramT
}

// Get returns the current routes Params
func (u *urlParams) Get(pname string) string {
	for i := len(u.kvs) - 1; i >= 0; i-- {
		if u.kvs[i].key == pname {
			return u.kvs[i].value
		}
	}
	return ""
}

// Set sets the name in URL to the value.
func (u *urlParams) Set(pname, pvalue string) {
	for i := len(u.kvs) - 1; i >= 0; i-- {
		if u.kvs[i].key == pname {
			u.kvs[i].value = pvalue
			return
		}
	}
	u.kvs = append(u.kvs, paramT{key: pname, value: pvalue})
}

func (u *urlParams) Each(f func(string, string)) {
	for i := range u.kvs {
		f(u.kvs[i].key, u.kvs[i].value)
	}
}

// Reset clears all the key-values.
func (u *urlParams) Reset() {
	u.kvs = u.kvs[:0]
}

type contextkey int

const urlParamCtxKey contextkey = 0

// GetURLParam returns the URLParam from the request.
//
// Return nil if the request does not have the url param.
func GetURLParam(r *http.Request) URLParam {
	if up := r.Context().Value(urlParamCtxKey); up != nil {
		return up.(URLParam)
	}
	return nil
}

// SetURLParam sets the URLParam into http.Request.
func SetURLParam(r *http.Request, up URLParam) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), urlParamCtxKey, up))
}
