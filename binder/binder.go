// The MIT License (MIT)
//
// Copyright (c) 2017 LabStack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package binder

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/xgfone/ship/core"
)

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
type BindUnmarshaler interface {
	// Unmarshal decodes and assigns a value from the
	UnmarshalBind(param string) error
}

// BindQuery binds the request query to v.
func BindQuery(queries url.Values, v interface{}) error {
	if err := globalBinder.bindData(v, queries, "query"); err != nil {
		return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
	}
	return nil
}

// NewBinder returns a new default Binder.
func NewBinder() core.Binder {
	return globalBinder
}

// defaultBinder is the default implementation of the Binder interface.
type defaultBinder struct{}

var globalBinder = &defaultBinder{}

// Bind implements the `Binder#Bind` function.
func (b *defaultBinder) Bind(ctx core.Context, v interface{}) (err error) {
	req := ctx.Request()
	if req.ContentLength == 0 {
		if req.Method == http.MethodGet || req.Method == http.MethodDelete {
			if err = b.bindData(v, ctx.QueryParams(), "query"); err != nil {
				return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
			}
			return
		}
		return core.NewHTTPError(http.StatusBadRequest, "Request body can't be empty")
	}
	ctype := req.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ctype, "application/json"):
		if err = json.NewDecoder(req.Body).Decode(v); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				return core.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, offset=%v",
						ute.Type, ute.Value, ute.Offset)).SetInnerError(err)
			} else if se, ok := err.(*json.SyntaxError); ok {
				return core.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Syntax error: offset=%v, error=%v",
						se.Offset, se.Error())).SetInnerError(err)
			}
			return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
		}
	case strings.HasPrefix(ctype, "application/xml"), strings.HasPrefix(ctype, "text/xml"):
		if err = xml.NewDecoder(req.Body).Decode(v); err != nil {
			if ute, ok := err.(*xml.UnsupportedTypeError); ok {
				return core.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Unsupported type error: type=%v, error=%v",
						ute.Type, ute.Error())).SetInnerError(err)
			} else if se, ok := err.(*xml.SyntaxError); ok {
				return core.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("Syntax error: line=%v, error=%v",
						se.Line, se.Error())).SetInnerError(err)
			}
			return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
		}
	case strings.HasPrefix(ctype, "application/x-www-form-urlencoded"),
		strings.HasPrefix(ctype, "multipart/form-data"):
		params, err := ctx.FormParams()
		if err != nil {
			return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
		}
		if err = b.bindData(v, params, "form"); err != nil {
			return core.NewHTTPError(http.StatusBadRequest, err.Error()).SetInnerError(err)
		}
	default:
		return core.ErrUnsupportedMediaType
	}
	return
}

func (b *defaultBinder) bindData(ptr interface{}, data map[string][]string, tag string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)

		if inputFieldName == "" {
			inputFieldName = typeField.Name
			// If tag is nil, we inspect if the field is a struct.
			if _, ok := bindUnmarshaler(structField); !ok && structFieldKind == reflect.Struct {
				if err := b.bindData(structField.Addr().Interface(), data, tag); err != nil {
					return err
				}
				continue
			}
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			// Go json.Unmarshal supports case insensitive binding.  However the
			// url params are bound case sensitive which is inconsistent.  To
			// fix this we must check all of the map values in a
			// case-insensitive search.
			inputFieldName = strings.ToLower(inputFieldName)
			for k, v := range data {
				if strings.ToLower(k) == inputFieldName {
					inputValue = v
					exists = true
					break
				}
			}
		}

		if !exists {
			continue
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(typeField.Type.Kind(), inputValue[0], structField); ok {
			if err != nil {
				return err
			}
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j], slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
			return err
		}
	}

	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	// But also call it here, in case we're dealing with an array of BindUnmarshalers
	if ok, err := unmarshalField(valueKind, val, structField); ok {
		return err
	}

	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

// bindUnmarshaler attempts to unmarshal a reflect.Value into a BindUnmarshaler
func bindUnmarshaler(field reflect.Value) (BindUnmarshaler, bool) {
	ptr := reflect.New(field.Type())
	if ptr.CanInterface() {
		iface := ptr.Interface()
		if unmarshaler, ok := iface.(BindUnmarshaler); ok {
			return unmarshaler, ok
		}
	}
	return nil, false
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	if unmarshaler, ok := bindUnmarshaler(field); ok {
		err := unmarshaler.UnmarshalBind(value)
		field.Set(reflect.ValueOf(unmarshaler).Elem())
		return true, err
	}
	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		// Initialize the pointer to a nil value
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
