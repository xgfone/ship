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
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
)

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
		buf = make([]byte, 2048)
	}

	// For like byte.Buffer, we maybe grow its capacity to avoid allocating
	// the memory more times.
	if b, ok := dst.(interface{ Grow(int) }); ok {
		if n < 32768 { // 32KB
			b.Grow(int(n))
		} else {
			b.Grow(32768)
		}
	}

	written, err = io.CopyBuffer(dst, io.LimitReader(src, n), buf)
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

// SetStructFieldToDefault sets the default value of the fields of the struct v
// to the value of the tag "default" of the fields when the field value is ZERO.
//
// If v is not a struct, it does nothing; and not a pointer to struct, panic.
//
// For the type of the field, it only support:
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
//
func SetStructFieldToDefault(v interface{}) (err error) {
	vf := reflect.ValueOf(v)
	switch kind := vf.Kind(); kind {
	case reflect.Ptr:
		vf = vf.Elem()
		if vf.Kind() != reflect.Struct {
			return errors.New("the argument must be a pointer to struct")
		}
		err = setDefault(vf)
	case reflect.Struct:
		return errors.New("the argument must be a pointer to struct")
	}

	return
}

func setDefault(vf reflect.Value) (err error) {
	vt := vf.Type()
	for i, _len := 0, vt.NumField(); i < _len; i++ {
		tag := strings.TrimSpace(vt.Field(i).Tag.Get("default"))
		if tag == "" {
			continue
		}

		fieldv := vf.Field(i)
		switch v := fieldv.Interface().(type) {
		case string:
			if v == "" {
				fieldv.SetString(tag)
			}
		case int:
			err = setFieldInt(fieldv, int64(v), tag)
		case int8:
			err = setFieldInt(fieldv, int64(v), tag)
		case int16:
			err = setFieldInt(fieldv, int64(v), tag)
		case int32:
			err = setFieldInt(fieldv, int64(v), tag)
		case int64:
			err = setFieldInt(fieldv, v, tag)
		case uint:
			err = setFieldUint(fieldv, uint64(v), tag)
		case uint8:
			err = setFieldUint(fieldv, uint64(v), tag)
		case uint16:
			err = setFieldUint(fieldv, uint64(v), tag)
		case uint32:
			err = setFieldUint(fieldv, uint64(v), tag)
		case uint64:
			err = setFieldUint(fieldv, v, tag)
		case uintptr:
			err = setFieldUint(fieldv, uint64(v), tag)
		case float32:
			err = setFieldFloat(fieldv, float64(v), tag)
		case float64:
			err = setFieldFloat(fieldv, v, tag)
		}

		if err != nil {
			return
		}
	}

	return
}

func setFieldInt(fieldv reflect.Value, v int64, tag string) (err error) {
	if v == 0 {
		if v, err = strconv.ParseInt(tag, 10, 64); err == nil {
			fieldv.SetInt(v)
		}
	}
	return
}

func setFieldUint(fieldv reflect.Value, v uint64, tag string) (err error) {
	if v == 0 {
		if v, err = strconv.ParseUint(tag, 10, 64); err == nil {
			fieldv.SetUint(v)
		}
	}
	return
}

func setFieldFloat(fieldv reflect.Value, v float64, tag string) (err error) {
	if v == 0 {
		if v, err = strconv.ParseFloat(tag, 64); err == nil {
			fieldv.SetFloat(v)
		}
	}
	return
}
