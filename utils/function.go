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
	"reflect"

	"github.com/xgfone/go-tools/v6/function"
)

// SetDefaultForStruct sets the default value by the tag "default".
func SetDefaultForStruct(v interface{}) (err error) {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		panic("the value is not a pointer")
	} else if value = value.Elem(); value.Kind() != reflect.Struct {
		panic("the value is not a pointer to struct")
	}

	vtype := value.Type()
	for i, num := 0, value.NumField(); i < num; i++ {
		if v := vtype.Field(i).Tag.Get("default"); v != "" {
			err = function.SetValue(value.Field(i).Addr().Interface(), v)
			if err != nil {
				return
			}
		}
	}

	return nil
}
