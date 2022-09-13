// Copyright 2022 xgfone
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

package binder

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method
// to unmarshal itself from the string parameter.
type BindUnmarshaler interface {
	// Unmarshal decodes the argument param and assigns to itself.
	UnmarshalBind(param string) error
}

// BindURLValuesAndFiles parses the data and assign to the pointer ptr to a struct.
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
// And any pointer to the type above, and
//   - *multipart.FileHeader
//   - []*multipart.FileHeader
//   - interface { UnmarshalBind(param string) error }
//
func BindURLValuesAndFiles(ptr interface{}, data url.Values,
	files map[string][]*multipart.FileHeader, tag string) error {
	value := reflect.ValueOf(ptr)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", ptr)
	}
	return bindURLValues(value.Elem(), files, data, tag)
}

// BindURLValues is equal to BindURLValuesAndFiles(ptr, data, nil, tag).
func BindURLValues(ptr interface{}, data url.Values, tag string) error {
	return BindURLValuesAndFiles(ptr, data, nil, tag)
}

func bindURLValues(val reflect.Value, files map[string][]*multipart.FileHeader,
	data url.Values, tag string) (err error) {
	valType := val.Type()
	if valType.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i, num := 0, valType.NumField(); i < num; i++ {
		field := valType.Field(i)
		fieldName := field.Tag.Get(tag)
		switch fieldName = strings.TrimSpace(fieldName); fieldName {
		case "":
			fieldName = field.Name
		case "-":
			continue
		}

		fieldValue := val.Field(i)
		fieldKind := fieldValue.Kind()
		if field.Anonymous && fieldKind == reflect.Struct {
			if err = bindURLValues(fieldValue, files, data, tag); err != nil {
				return err
			}
			continue
		} else if !fieldValue.CanSet() {
			continue
		}

		inputValue, exists := data[fieldName]
		if !exists {
			if fhs := files[fieldName]; len(fhs) > 0 {
				switch fieldValue.Interface().(type) {
				case *multipart.FileHeader:
					fieldValue.Set(reflect.ValueOf(fhs[0]))
				case []*multipart.FileHeader:
					fieldValue.Set(reflect.ValueOf(fhs))
				}
			}
			continue
		} else if len(inputValue) == 0 {
			continue
		}

		if fieldKind == reflect.Slice {
			num := len(inputValue)
			kind := field.Type.Elem().Kind()
			slice := reflect.MakeSlice(field.Type, num, num)
			for j := 0; j < num; j++ {
				err = setWithProperType(kind, slice.Index(j), inputValue[j])
				if err != nil {
					return
				}
			}
			fieldValue.Set(slice)
		} else {
			err = setWithProperType(fieldKind, fieldValue, inputValue[0])
			if err != nil {
				return
			}
		}
	}

	return
}

var binderType = reflect.TypeOf((*BindUnmarshaler)(nil)).Elem()

func bindUnmarshaler(kind reflect.Kind, val reflect.Value, value string) (ok bool, err error) {
	if kind != reflect.Ptr && kind != reflect.Interface {
		val = val.Addr()
	}

	if val.Type().Implements(binderType) {
		if unmarshaler, ok := val.Interface().(BindUnmarshaler); ok {
			return true, unmarshaler.UnmarshalBind(value)
		}
	}

	return false, nil
}

func setWithProperType(kind reflect.Kind, value reflect.Value, input string) error {
	if kind == reflect.Ptr && value.IsNil() {
		value.Set(reflect.New(value.Type().Elem()))
	} else if kind == reflect.Interface && value.IsNil() {
		panic("the bind struct field interface value must not be nil")
	}

	if ok, err := bindUnmarshaler(kind, value, input); ok {
		return err
	}

	switch kind {
	case reflect.Ptr:
		value = value.Elem()
		return setWithProperType(value.Kind(), value, input)
	case reflect.Int:
		return setIntField(input, 0, value)
	case reflect.Int8:
		return setIntField(input, 8, value)
	case reflect.Int16:
		return setIntField(input, 16, value)
	case reflect.Int32:
		return setIntField(input, 32, value)
	case reflect.Int64:
		if _, ok := value.Interface().(time.Duration); ok {
			v, err := time.ParseDuration(input)
			if err == nil {
				value.SetInt(int64(v))
			}
			return err
		}
		return setIntField(input, 64, value)
	case reflect.Uint:
		return setUintField(input, 0, value)
	case reflect.Uint8:
		return setUintField(input, 8, value)
	case reflect.Uint16:
		return setUintField(input, 16, value)
	case reflect.Uint32:
		return setUintField(input, 32, value)
	case reflect.Uint64:
		return setUintField(input, 64, value)
	case reflect.Bool:
		return setBoolField(input, value)
	case reflect.Float32:
		return setFloatField(input, 32, value)
	case reflect.Float64:
		return setFloatField(input, 64, value)
	case reflect.String:
		value.SetString(input)
	default:
		if _, ok := value.Interface().(time.Time); ok {
			if input == "" {
				return nil
			}
			return value.Addr().Interface().(*time.Time).UnmarshalText([]byte(input))
		}
		return fmt.Errorf("unknown field type '%T'", value.Interface())
	}
	return nil
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
