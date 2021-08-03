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
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errInvalidTagValue    = errors.New("invalid tag value")
	errNotPointerToStruct = errors.New("the argument must be a pointer to struct")
)

// Defaulter is used to set the default value if the data or the field of data
// is ZERO.
type Defaulter interface {
	SetDefault(data interface{}) error
}

// DefaulterFunc is the function type implementing the interface Defaulter.
type DefaulterFunc func(interface{}) error

// SetDefault implements the interface Defaulter.
func (d DefaulterFunc) SetDefault(data interface{}) error { return d(data) }

// NothingDefaulter returns a Defaulter that does nothing.
func NothingDefaulter() Defaulter { return DefaulterFunc(nothingDefaulter) }

func nothingDefaulter(interface{}) error { return nil }

// SetStructFieldToDefault sets the default value of the fields of the pointer
// to struct v to the value of the tag "default" of the fields when the field
// value is ZERO.
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
//   time.Time      // Format: A. Integer(UTC); B. String(RFC3339)
//   time.Duration  // Format: A. Integer(ms);  B. String(time.ParseDuration)
//   pointer to the types above
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
		fieldv := vf.Field(i)

		tag := strings.TrimSpace(vt.Field(i).Tag.Get("default"))
		if fieldv.Kind() == reflect.Ptr {
			if !fieldv.IsNil() {
				fieldv = fieldv.Elem()
			} else if tag != "" {
				fieldv.Set(reflect.New(fieldv.Type().Elem()))
				fieldv = fieldv.Elem()
			}
		}

		if !fieldv.CanSet() {
			continue
		}

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
		case time.Duration:
			err = setFieldDuration(vf, fieldv, v, tag)
		case time.Time:
			err = setFieldTime(vf, fieldv, v, tag)
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
					if f := fieldv.Index(i); f.Kind() == reflect.Struct {
						if err = setDefault(f); err != nil {
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

func setFieldTime(structv, fieldv reflect.Value, v time.Time, tag string) (err error) {
	if v.IsZero() && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if IsInteger(tag) {
			var i int64
			if i, err = strconv.ParseInt(tag, 10, 64); err == nil {
				fieldv.Set(reflect.ValueOf(time.Unix(i, 0)))
			}
		} else if v, err = time.Parse(time.RFC3339, tag); err == nil {
			fieldv.Set(reflect.ValueOf(v))
		}
	}
	return
}

func setFieldDuration(structv, fieldv reflect.Value, v time.Duration, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if IsInteger(tag) {
			var i int64
			if i, err = strconv.ParseInt(tag, 10, 64); err == nil {
				fieldv.SetInt(i * int64(time.Millisecond))
			}
		} else if v, err = time.ParseDuration(tag); err == nil {
			fieldv.SetInt(int64(v))
		}
	}
	return
}
