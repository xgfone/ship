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
	"encoding/json"
	"fmt"
)

func ExampleSetDefaultForStruct() {
	type S struct {
		Name string `json:"name" default:"Aaron"`
		Age  int    `json:"age" default:"18"`
	}

	var err error
	var s1, s2, s3 S

	SetDefaultForStruct(&s1)
	SetDefaultForStruct(&s2)
	SetDefaultForStruct(&s3)

	if err = json.Unmarshal([]byte(`{}`), &s1); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("name=%s, age=%d\n", s1.Name, s1.Age)
	}

	if err = json.Unmarshal([]byte(`{"name":"abc"}`), &s2); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("name=%s, age=%d\n", s2.Name, s2.Age)
	}

	if err = json.Unmarshal([]byte(`{"name":"abc","age":100}`), &s3); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("name=%s, age=%d\n", s3.Name, s3.Age)
	}

	// Output:
	// name=Aaron, age=18
	// name=abc, age=18
	// name=abc, age=100
}
