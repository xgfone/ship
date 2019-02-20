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
	"fmt"
	"testing"
	"time"
)

type stringSetValuer struct {
	name string
}

func (s *stringSetValuer) SetValue(v interface{}) error {
	return SetValue(&s.name, v)
}

type mapSetValuer struct {
	name string
	age  int
}

func (m *mapSetValuer) SetValue(v interface{}) error {
	if ms, ok := v.(map[string]interface{}); ok {
		if err := SetValue(&m.name, ms["name"]); err != nil {
			return err
		}
		if err := SetValue(&m.age, ms["age"]); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("the value is not a map")
}

func TestSetValue(t *testing.T) {
	var b bool
	var bs []byte
	var s string
	var f32 float32
	var f64 float64
	var i int
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var u uint
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var u64 uint64
	var tt1 time.Time
	var tt2 time.Time
	var svaluer stringSetValuer
	var mvaluer mapSetValuer

	if err := SetValue(&svaluer, "abc"); err != nil || svaluer.name == "" {
		t.Error(svaluer)
	}
	if err := SetValue(&mvaluer, map[string]interface{}{"name": "abc", "age": 123}); err != nil || mvaluer.name != "abc" || mvaluer.age != 123 {
		t.Error(mvaluer)
	}

	if err := SetValue(&b, "on"); err != nil || !b {
		t.Fail()
	}
	if err := SetValue(&bs, "bytes"); err != nil || string(bs) != "bytes" {
		t.Fail()
	}
	if err := SetValue(&s, "string"); err != nil || s != "string" {
		t.Fail()
	}
	if err := SetValue(&f32, "1.0"); err != nil || f32 != 1.0 {
		t.Fail()
	}
	if err := SetValue(&f64, "1.0"); err != nil || f64 != 1.0 {
		t.Fail()
	}
	if err := SetValue(&i, "123"); err != nil || i != 123 {
		t.Fail()
	}
	if err := SetValue(&i8, "123"); err != nil || i8 != 123 {
		t.Fail()
	}
	if err := SetValue(&i16, "123"); err != nil || i16 != 123 {
		t.Fail()
	}
	if err := SetValue(&i32, "123"); err != nil || i32 != 123 {
		t.Fail()
	}
	if err := SetValue(&i64, "123"); err != nil || i64 != 123 {
		t.Fail()
	}
	if err := SetValue(&u, "123"); err != nil || u != 123 {
		t.Fail()
	}
	if err := SetValue(&u8, "123"); err != nil || u8 != 123 {
		t.Fail()
	}
	if err := SetValue(&u16, "123"); err != nil || u16 != 123 {
		t.Fail()
	}
	if err := SetValue(&u32, "123"); err != nil || u32 != 123 {
		t.Fail()
	}
	if err := SetValue(&u64, "123"); err != nil || u64 != 123 {
		t.Fail()
	}
	if err := SetValue(&tt1, "2019-01-16T15:39:40Z"); err != nil || tt1.String() != "2019-01-16 15:39:40 +0000 UTC" {
		t.Error(tt1)
	}
	if err := SetValue(&tt2, "2019-01-16T15:39:40+08:00"); err != nil ||
		(tt2.String() != "2019-01-16 15:39:40 +0800 CST" && tt2.String() != "2019-01-16 15:39:40 +0800 +0800") {
		t.Error(tt2)
	}

	tt2 = tt2.UTC()

	if tt1.Year() != tt2.Year() {
		t.Error(tt1.Year(), tt2.Year())
	}
	if tt1.Month() != tt2.Month() {
		t.Error(tt1.Month(), tt2.Month())
	}
	if tt1.Day() != tt2.Day() {
		t.Error(tt1.Day(), tt2.Day())
	}
	if tt1.Hour()-tt2.Hour() != 8 {
		t.Error(tt1.Hour(), tt2.Hour())
	}
	if tt1.Minute() != tt2.Minute() {
		t.Error(tt1.Minute(), tt2.Minute())
	}
	if tt1.Second() != tt2.Second() {
		t.Error(tt1.Second(), tt2.Second())
	}
}

func TestSetStructValue(t *testing.T) {
	type S struct {
		Name string
		Age  int
	}
	s := S{}
	if err := SetStructValue(&s, "Name", "abc"); err != nil {
		t.Error(err)
	}
	if err := SetStructValue(&s, "Age", "123"); err != nil {
		t.Error(err)
	}
	if s.Name != "abc" || s.Age != 123 {
		t.Error(s.Name, s.Age)
	}
}

func TestBindMapToStruct(t *testing.T) {
	type subTypeT struct {
		Float float32 `json:"float"`
	}
	type typeT struct {
		Int  int
		Bool bool      `json:"bool"`
		Str  string    `json:"-"`
		Sub  subTypeT  `json:"sub"`
		PSub *subTypeT `json:"psub"`
		NPub *subTypeT `json:"npub"`
	}

	ms := map[string]interface{}{
		"Int":  123,
		"bool": "on",
		"Str":  "abc",
		"str":  "xyz",
		"sub":  map[string]interface{}{"float": 1.2, "Float": 1.4},
		"psub": map[string]interface{}{"float": 2.2, "Float": 2.4},
		"nsub": map[string]interface{}{"float": 3.2, "Float": 3.4},
	}

	var t1 typeT
	if err := BindMapToStruct(t1, ms); err == nil {
		t.Fail()
	}

	var t2 typeT
	t2.PSub = &subTypeT{}
	if err := BindMapToStruct(&t2, ms); err != nil {
		t.Error(err)
	} else if t2.Int == 0 || !t2.Bool || t2.Str != "" || t2.Sub.Float != 1.2 || t2.PSub.Float != 2.2 || t2.NPub != nil {
		t.Error(t2)
	}
}

func ExampleBindMapToStruct() {
	type testSubType struct {
		Float float32 `json:"float"`
	}
	type testType struct {
		Int  int
		Bool bool         `json:"bool"`
		Str  string       `json:"-"`
		Sub  testSubType  `json:"sub"`
		PSub *testSubType `json:"psub"`
	}

	ms := map[string]interface{}{
		"Int":  123,
		"bool": "on",
		"Str":  "abc",
		"str":  "xyz",
		"sub":  map[string]interface{}{"float": 1.2, "Float": 1.4},
		"psub": map[string]interface{}{"float": 2.2, "Float": 2.4},
	}

	v := testType{PSub: &testSubType{}}
	if err := BindMapToStruct(&v, ms); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(v.Int)
		fmt.Println(v.Bool)
		fmt.Println(v.Str)
		fmt.Println(v.Sub.Float)
		fmt.Println(v.PSub.Float)
	}

	// Output:
	// 123
	// true
	//
	// 1.2
	// 2.2
}
