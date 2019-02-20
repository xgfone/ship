// Copyright 2019 xgfone <xgfone@126.com>
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
	"reflect"
	"time"

	"github.com/xgfone/ship/utils"
)

// SetValue binds data to v which must be a pointer.
//
// The converting rule between the types of data and v:
//
//    bool, string, number           ->  *bool
//    bool, string, number, []byte,  ->  *string
//    bool, string, number, []byte,  ->  *[]byte
//    bool, string, number           ->  *float32
//    bool, string, number           ->  *float64
//    bool, string, number           ->  *int
//    bool, string, number           ->  *int8
//    bool, string, number           ->  *int16
//    bool, string, number           ->  *int32
//    bool, string, number           ->  *int64
//    bool, string, number           ->  *uint
//    bool, string, number           ->  *uint8
//    bool, string, number           ->  *uint16
//    bool, string, number           ->  *uint32
//    bool, string, number           ->  *uint64
//    string, time.Time              ->  *time.Time
//
// Notice: number stands for all the integer and float types.
//
// For bool, "t", "T", "1", "on", "On", "ON", "true", "True", "TRUE" are true,
// and "f", "F", "0", "off", "Off", "OFF", "false", "False", "FALSE" are false.
// Others is invalid.
//
// For time.Time, it supports the layout ISO8601 and RFC3339. If it's ISO8601,
// the time must be UTC. So you can parse the time as follow:
//
//     var t1, t2 time.Time
//     SetValue(&t1, "2019-01-16T15:39:40Z")
//     SetValue(&t2, "2019-01-16T15:39:40+08:00")
//
func SetValue(v interface{}, data interface{}) (err error) {
	var u64 uint64
	var i64 int64

	switch p := v.(type) {
	case *bool:
		switch data.(type) {
		case bool:
			*p = data.(bool)
		case string, float32, float64,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			*p, err = utils.ToBool(data)
		default:
			return fmt.Errorf("the unknown type '%T'", data)
		}
	case *string:
		*p, err = utils.ToString(data)
	case *[]byte:
		switch d := data.(type) {
		case string:
			*p = []byte(d)
		case []byte:
			*p = d
		default:
			s, e := utils.ToString(data)
			*p = []byte(s)
			err = e
		}
	case *float32:
		f64, e := utils.ToFloat64(data)
		*p = float32(f64)
		err = e
	case *float64:
		*p, err = utils.ToFloat64(data)
	case *int:
		i64, err = utils.ToInt64(data)
		*p = int(i64)
	case *int8:
		i64, err = utils.ToInt64(data)
		*p = int8(i64)
	case *int16:
		i64, err = utils.ToInt64(data)
		*p = int16(i64)
	case *int32:
		i64, err = utils.ToInt64(data)
		*p = int32(i64)
	case *int64:
		*p, err = utils.ToInt64(data)
	case *uint:
		u64, err = utils.ToUint64(data)
		*p = uint(u64)
	case *uint8:
		u64, err = utils.ToUint64(data)
		*p = uint8(u64)
	case *uint16:
		u64, err = utils.ToUint64(data)
		*p = uint16(u64)
	case *uint32:
		u64, err = utils.ToUint64(data)
		*p = uint32(u64)
	case *uint64:
		*p, err = utils.ToUint64(data)
	case *time.Time:

		switch d := data.(type) {
		case time.Time:
			*p = d
		case string:
			_len := len(d)
			if _len == 0 {
				return errors.New("the data is empty")
			}
			if d[_len-1] == 'Z' {
				*p, err = time.ParseInLocation("2006-01-02T15:04:05Z", d, time.UTC)
			} else {
				*p, err = time.Parse(time.RFC3339, d)
			}
		default:
			return fmt.Errorf("the unknown type '%T'", data)
		}
	default:
		return fmt.Errorf("the unknown type '%T'", data)
	}
	return
}

// SetStructValue is equal to SetStructValueOf(reflect.ValueOf(s), attr, v)
func SetStructValue(s interface{}, attr string, v interface{}) error {
	if s == nil {
		return errors.New("the struct value is nil")
	}
	return SetStructValueOf(reflect.ValueOf(s), attr, v)
}

// SetStructValueOf is the same as SetValue, but binds the attribute attr of
// the struct s to v.
func SetStructValueOf(s reflect.Value, attr string, v interface{}) error {
	if attr == "" {
		return errors.New("the name of the struct attribute is empty")
	}
	if v == nil {
		return errors.New("the value is nil")
	}

	if s.Kind() != reflect.Ptr {
		return errors.New("the struce value is not a pointer")
	}
	if s = s.Elem(); s.Kind() != reflect.Struct {
		return errors.New("the struct value is not a pointer to a struct")
	}

	st := s.Type()
	for i := st.NumField() - 1; i >= 0; i-- {
		if st.Field(i).Name == attr {
			return SetValue(s.Field(i).Addr().Interface(), v)
		}
	}

	return fmt.Errorf("the struct has no field '%s'", attr)
}

// BindMapToStruct binds a map to struct.
//
// Notice: it supports the json tag.
func BindMapToStruct(value interface{}, m map[string]interface{}) (err error) {
	if value == nil {
		return errors.New("the value is nil")
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		return errors.New("the value is not a pointer")
	} else if v = v.Elem(); v.Kind() != reflect.Struct {
		return errors.New("the value is not a pointer to struct")
	}
	return bindMapToStruct(v, m)
}

func bindMapToStruct(v reflect.Value, m map[string]interface{}) (err error) {
	vtype := v.Type()
	for i, num := 0, v.NumField(); i < num; i++ {
		fieldv := v.Field(i)
		fieldt := vtype.Field(i)

		// Check whether the field can be set.
		if !fieldv.CanSet() {
			return fmt.Errorf("the field '%s' can't be set", fieldt.Name)
		}

		name := fieldt.Name
		if n := fieldt.Tag.Get("json"); n != "" {
			if n == "-" {
				continue
			}
			name = n
		}

		if fieldv.Kind() == reflect.Ptr {
			switch subfieldv := fieldv.Elem(); subfieldv.Kind() {
			case reflect.Invalid:
				continue
			case reflect.Struct:
				if mvalue, ok := m[name].(map[string]interface{}); ok {
					if err = bindMapToStruct(subfieldv, mvalue); err != nil {
						return err
					}
					continue
				}
				return fmt.Errorf("the value of '%s' is not a map", name)
			}
		} else if fieldv.Kind() == reflect.Struct {
			if mvalue, ok := m[name].(map[string]interface{}); ok {
				if err = bindMapToStruct(fieldv, mvalue); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("the value of '%s' is not a map", name)
		} else {
			fieldv = fieldv.Addr()
		}

		if err = SetValue(fieldv.Interface(), m[name]); err != nil {
			return err
		}
	}

	return nil
}
