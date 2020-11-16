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

package ship

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var bufpool = sync.Pool{
	New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 1024)) },
}

func getBuffer() *bytes.Buffer    { return bufpool.Get().(*bytes.Buffer) }
func putBuffer(buf *bytes.Buffer) { buf.Reset(); bufpool.Put(buf) }

// OnceRunner is used to run the task only once, which is different from
// sync.Once, the second calling does not wait until the first calling finishes.
type OnceRunner struct {
	done uint32
	task func()
}

// NewOnceRunner returns a new OnceRunner.
func NewOnceRunner(task func()) *OnceRunner { return &OnceRunner{task: task} }

// Run runs the task.
func (r *OnceRunner) Run() {
	if atomic.CompareAndSwapUint32(&r.done, 0, 1) {
		r.task()
	}
}

// CopyNBuffer is the same as io.CopyN, but uses the given buf as the buffer.
//
// If buf is nil or empty, it will make a new one with 2048.
func CopyNBuffer(dst io.Writer, src io.Reader, n int64, buf []byte) (written int64, err error) {
	if len(buf) == 0 {
		buf = make([]byte, 1024)
	}

	// For like byte.Buffer, we maybe grow its capacity to avoid allocating
	// the memory more times.
	if b, ok := dst.(interface{ Grow(int) }); ok && n > 0 {
		if n < 32768 { // 32KB
			b.Grow(int(n))
		} else {
			b.Grow(32768)
		}
	}

	// (xgfone): Fix for compression, such as gzip or deflate.
	if n > 0 {
		src = io.LimitReader(src, n)
	}

	written, err = io.CopyBuffer(dst, src, buf)
	if written == n {
		return n, nil
	} else if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}

	return
}

// DisalbeRedirect is used to disalbe the default redirect behavior
// of http.Client, that's, http.Client won't handle the redirect response
// and just return it to the caller.
func DisalbeRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

var (
	errInvalidTagValue    = errors.New("invalid tag value")
	errNotPointerToStruct = errors.New("the argument must be a pointer to struct")
)

// SetStructFieldToDefault sets the default value of the fields of the struct v
// to the value of the tag "default" of the fields when the field value is ZERO.
//
// If v is not a struct, it does nothing; and not a pointer to struct, panic.
//
// For the type of the field, it only supports some base types as follow:
//   string
//   float32
//   float64
//   int
//   int8
//   int16
//   int32
//   int64
//   uint
//   uint8
//   uint16
//   uint32
//   uint64
//   struct
//   struct slice
//   interface{ SetDefault(_default interface{}) error }
//
// Notice: If the tag value starts with ".", it represents a field name and
// the default value of current field is set to the value of that field.
// But their types must be consistent, or panic.
func SetStructFieldToDefault(v interface{}) (err error) {
	vf := reflect.ValueOf(v)
	switch kind := vf.Kind(); kind {
	case reflect.Ptr:
		vf = vf.Elem()
		if vf.Kind() != reflect.Struct {
			return errNotPointerToStruct
		}
		err = setDefault(vf)
	case reflect.Struct:
		return errNotPointerToStruct
	}

	return
}

type setDefaulter interface {
	SetDefault(_default interface{}) error
}

func setDefault(vf reflect.Value) (err error) {
	vt := vf.Type()
	for i, _len := 0, vt.NumField(); i < _len; i++ {
		tag := strings.TrimSpace(vt.Field(i).Tag.Get("default"))

		fieldv := vf.Field(i)
		switch v := fieldv.Interface().(type) {
		case string:
			if v == "" && tag != "" {
				fieldv.SetString(tag)
			}
		case int:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int8:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int16:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int32:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int64:
			err = setFieldInt(vf, fieldv, v, tag)
		case uint:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint8:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint16:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint32:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint64:
			err = setFieldUint(vf, fieldv, v, tag)
		case uintptr:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case float32:
			err = setFieldFloat(vf, fieldv, float64(v), tag)
		case float64:
			err = setFieldFloat(vf, fieldv, v, tag)
		case setDefaulter:
			if tag != "" {
				err = v.SetDefault(tag)
			}

		default:
			switch fieldv.Kind() {
			case reflect.Struct:
				err = setDefault(fieldv)
			case reflect.Slice:
				for i, _len := 0, fieldv.Len(); i < _len; i++ {
					if _fieldv := fieldv.Index(i); _fieldv.Kind() == reflect.Struct {
						if err = setDefault(_fieldv); err != nil {
							return
						}
					}
				}
			}
		}

		if err != nil {
			return
		}
	}

	return
}

func setFieldInt(structv, fieldv reflect.Value, v int64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseInt(tag, 10, 64); err == nil {
			fieldv.SetInt(v)
		}
	}
	return
}

func setFieldUint(structv, fieldv reflect.Value, v uint64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseUint(tag, 10, 64); err == nil {
			fieldv.SetUint(v)
		}
	}
	return
}

func setFieldFloat(structv, fieldv reflect.Value, v float64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseFloat(tag, 64); err == nil {
			fieldv.SetFloat(v)
		}
	}
	return
}

// GetJSON is equal to RequestJSON(http.MethodGet, url, nil, resp...).
func GetJSON(url string, resp ...interface{}) (err error) {
	return RequestJSON(http.MethodGet, url, nil, resp...)
}

// PostJSON is equal to RequestJSON(http.MethodPost, url, req, resp...).
func PostJSON(url string, req interface{}, resp ...interface{}) (err error) {
	return RequestJSON(http.MethodPost, url, req, resp...)
}

// PutJSON is equal to RequestJSON(http.MethodPut, url, req, resp...).
func PutJSON(url string, req interface{}, resp ...interface{}) (err error) {
	return RequestJSON(http.MethodPut, url, req, resp...)
}

// DeleteJSON is equal to RequestJSON(http.MethodDelete, url, req, resp...).
func DeleteJSON(url string, req interface{}, resp ...interface{}) (err error) {
	return RequestJSON(http.MethodDelete, url, req, resp...)
}

// RequestJSON is equal to RequestJSONWithContext(context.Background(), ...).
func RequestJSON(method, url string, req interface{}, resp ...interface{}) (err error) {
	return RequestJSONWithContext(context.Background(), method, url, req, resp...)
}

// RequestJSONWithContext sends the http request with JSON and puts the response
// body into respBody as JSON.
//
// reqBody may be one of types: nil, []byte, string, io.Reader, and otehr types.
// For other types, it will be serialized by json.NewEncoder.
//
// If respBody is nil, it will ignore the response body.
//
// If respBody[1] is a function and its type is func(*http.Request)*http.Request,
// it will call it to fix the new request and use the returned request.
func RequestJSONWithContext(ctx context.Context, method, url string,
	reqBody interface{}, respBody ...interface{}) (err error) {
	var body io.Reader
	var buf *bytes.Buffer
	switch data := reqBody.(type) {
	case nil:
	case []byte:
		body = bytes.NewBuffer(data)
	case string:
		body = bytes.NewBufferString(data)
	case io.Reader:
		body = data
	default:
		buf = getBuffer()
		defer putBuffer(buf)
		if err = json.NewEncoder(buf).Encode(reqBody); err != nil {
			return NewHTTPClientError(method, url, 0, err)
		}
		body = buf
	}

	req, err := NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return NewHTTPClientError(method, url, 0, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if len(respBody) > 1 {
		if fix, ok := respBody[1].(func(*http.Request) *http.Request); ok {
			req = fix(req)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NewHTTPClientError(method, url, 0, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if buf == nil {
			buf = getBuffer()
			defer putBuffer(buf)
		}
		CopyNBuffer(buf, resp.Body, resp.ContentLength, nil)
		return NewHTTPClientError(method, url, resp.StatusCode, nil, buf.String())
	}

	if len(respBody) != 0 && respBody[0] != nil {
		if buf == nil {
			buf = getBuffer()
			defer putBuffer(buf)
		}

		if _, err = CopyNBuffer(buf, resp.Body, resp.ContentLength, nil); err == nil {
			err = json.Unmarshal(buf.Bytes(), respBody[0])
		}

		if err != nil {
			err = NewHTTPClientError(method, url, resp.StatusCode, err, buf.String())
		}
	}

	return
}
