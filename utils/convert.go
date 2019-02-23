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

package utils

import (
	"fmt"
	"strconv"
	"text/template"
)

// bool2Int converts the bool to int64.
func bool2Int64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// ToBool does the best to convert a certain value to bool
//
// For "t", "T", "1", "on", "On", "ON", "true", "True", "TRUE", it's true.
// For "f", "F", "0", "off", "Off", "OFF", "false", "False", "FALSE", it's false.
func ToBool(v interface{}) (bool, error) {
	switch _v := v.(type) {
	case nil:
		return false, nil
	case bool:
		return _v, nil
	case string:
		switch _v {
		case "t", "T", "1", "on", "On", "ON", "true", "True", "TRUE":
			return true, nil
		case "f", "F", "0", "off", "Off", "OFF", "false", "False", "FALSE", "":
			return false, nil
		default:
			return false, fmt.Errorf("unrecognized bool string: %s", _v)
		}
	}

	ok, _ := template.IsTrue(v)
	return ok, nil
}

// ToInt64 does the best to convert a certain value to int64.
func ToInt64(_v interface{}) (v int64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = bool2Int64(t)
	case string:
		v, err = strconv.ParseInt(t, 10, 64)
	case int:
		v = int64(t)
	case int8:
		v = int64(t)
	case int16:
		v = int64(t)
	case int32:
		v = int64(t)
	case int64:
		v = t
	case uint:
		v = int64(t)
	case uint8:
		v = int64(t)
	case uint16:
		v = int64(t)
	case uint32:
		v = int64(t)
	case uint64:
		v = int64(t)
	case float32:
		v = int64(t)
	case float64:
		v = int64(t)
	case complex64:
		v = int64(real(t))
	case complex128:
		v = int64(real(t))
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToUint64 does the best to convert a certain value to uint64.
func ToUint64(_v interface{}) (v uint64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = uint64(bool2Int64(t))
	case string:
		v, err = strconv.ParseUint(t, 10, 64)
	case int:
		v = uint64(t)
	case int8:
		v = uint64(t)
	case int16:
		v = uint64(t)
	case int32:
		v = uint64(t)
	case int64:
		v = uint64(t)
	case uint:
		v = uint64(t)
	case uint8:
		v = uint64(t)
	case uint16:
		v = uint64(t)
	case uint32:
		v = uint64(t)
	case uint64:
		v = t
	case float32:
		v = uint64(t)
	case float64:
		v = uint64(t)
	case complex64:
		v = uint64(real(t))
	case complex128:
		v = uint64(real(t))
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToFloat64 does the best to convert a certain value to float64.
func ToFloat64(_v interface{}) (v float64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = float64(bool2Int64(t))
	case string:
		v, err = strconv.ParseFloat(t, 64)
	case int:
		v = float64(t)
	case int8:
		v = float64(t)
	case int16:
		v = float64(t)
	case int32:
		v = float64(t)
	case int64:
		v = float64(t)
	case uint:
		v = float64(t)
	case uint8:
		v = float64(t)
	case uint16:
		v = float64(t)
	case uint32:
		v = float64(t)
	case uint64:
		v = float64(t)
	case float32:
		v = float64(t)
	case float64:
		v = t
	case complex64:
		v = float64(real(t))
	case complex128:
		v = real(t)
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToString does the best to convert a certain value to string.
func ToString(_v interface{}) (v string, err error) {
	switch t := _v.(type) {
	case nil:
	case string:
		v = t
	case []byte:
		v = string(t)
	case bool:
		if t {
			v = "true"
		} else {
			v = "false"
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		v = fmt.Sprintf("%d", t)
	case float32:
		v = strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		v = strconv.FormatFloat(t, 'f', -1, 64)
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}
