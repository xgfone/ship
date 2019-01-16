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
	"strconv"
	"time"
)

// SetValue binds data to v.
//
// v must be a pointer to the type as follow:
//
//     bool
//     string
//     []byte
//     float32
//     float64
//     int
//     int8
//     int16
//     int32
//     int64
//     uint
//     uint8
//     uint16
//     uint32
//     uint64
//     time.Time
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
func SetValue(v interface{}, data string) (err error) {
	var f64 float64
	var u64 uint64
	var i64 int64

	switch p := v.(type) {
	case *bool:
		switch data {
		case "t", "T", "1", "on", "On", "ON", "true", "True", "TRUE":
			*p = true
		case "f", "F", "0", "off", "Off", "OFF", "false", "False", "FALSE":
			*p = false
		default:
			return fmt.Errorf("invalid bool value '%s'", data)
		}
	case *string:
		*p = data
	case *[]byte:
		*p = []byte(data)
	case *float32:
		f64, err = strconv.ParseFloat(data, 32)
		*p = float32(f64)
	case *float64:
		*p, err = strconv.ParseFloat(data, 64)
	case *int:
		i64, err = strconv.ParseInt(data, 10, 0)
		*p = int(i64)
	case *int8:
		i64, err = strconv.ParseInt(data, 10, 8)
		*p = int8(i64)
	case *int16:
		i64, err = strconv.ParseInt(data, 10, 16)
		*p = int16(i64)
	case *int32:
		i64, err = strconv.ParseInt(data, 10, 32)
		*p = int32(i64)
	case *int64:
		*p, err = strconv.ParseInt(data, 10, 64)
	case *uint:
		u64, err = strconv.ParseUint(data, 10, 0)
		*p = uint(u64)
	case *uint8:
		u64, err = strconv.ParseUint(data, 10, 8)
		*p = uint8(u64)
	case *uint16:
		u64, err = strconv.ParseUint(data, 10, 16)
		*p = uint16(u64)
	case *uint32:
		u64, err = strconv.ParseUint(data, 10, 32)
		*p = uint32(u64)
	case *uint64:
		*p, err = strconv.ParseUint(data, 10, 64)
	case *time.Time:
		_len := len(data)
		if _len == 0 {
			return errors.New("the data is empty")
		}
		if data[_len-1] == 'Z' {
			*p, err = time.ParseInLocation("2006-01-02T15:04:05Z", data, time.UTC)
		} else {
			*p, err = time.Parse(time.RFC3339, data)
		}
	default:
		return errors.New("type is not supported")
	}
	return
}
