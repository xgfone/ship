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
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
type BindUnmarshaler interface {
	// Unmarshal decodes the argument param and assigns to itself.
	UnmarshalBind(param string) error
}

// BindURLValues parses the data and assign to the pointer ptr to a struct.
//
// Notice: tag is the name of the struct tag. such as "form", "query", etc.
// If the tag value is equal to "-", ignore this field.
//
// Support the types of the struct fields as follow:
//   - bool
//   - int
//   - int8
//   - int16
//   - int32
//   - int64
//   - uint
//   - uint8
//   - uint16
//   - uint32
//   - uint64
//   - string
//   - float32
//   - float64
//   - time.Time     // use time.Time.UnmarshalText(), so only support RFC3339 format
//   - time.Duration // use time.ParseDuration()
//   - interface { UnmarshalBind(param string) error }
// And any pointer to the type above.
//
func BindURLValues(ptr interface{}, data url.Values, tag string) error {
	value := reflect.ValueOf(ptr)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", ptr)
	}
	return bindURLValues(value.Elem(), data, tag)
}

func bindURLValues(val reflect.Value, data url.Values, tag string) (err error) {
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		inputFieldName := typeField.Tag.Get(tag)
		if inputFieldName = strings.TrimSpace(inputFieldName); inputFieldName == "-" {
			continue
		}

		structField := val.Field(i)
		structFieldKind := structField.Kind()
		if !structField.CanSet() {
			if typeField.Anonymous && structFieldKind == reflect.Struct {
				err := bindURLValues(structField, data, tag)
				if err != nil {
					return err
				}
			}
			continue
		}

		if inputFieldName == "" {
			inputFieldName = typeField.Name
			// If tag is nil, we inspect if the field is a struct.
			if _, ok := bindUnmarshaler(structField); !ok && structFieldKind == reflect.Struct {
				if err := bindURLValues(structField, data, tag); err != nil {
					return err
				}
				continue
			}
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			// Go json.Unmarshal supports case insensitive binding.
			// However the url params are bound case sensitive which is inconsistent.
			// To fix this we must check all of the map values in a case-insensitive search.
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

	return
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
		if _, ok := structField.Interface().(time.Duration); ok {
			v, err := time.ParseDuration(val)
			if err == nil {
				structField.SetInt(int64(v))
			}
			return err
		}
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
		if _, ok := structField.Interface().(time.Time); ok {
			if val == "" {
				return nil
			}
			return structField.Addr().Interface().(*time.Time).UnmarshalText([]byte(val))
		}
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
		return nil
	}

	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		return nil
	}

	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		return nil
	}

	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		return nil
	}

	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
