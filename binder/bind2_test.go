// The MIT License (MIT)
//
// Copyright (c) 2018 xgfone <xgfone@126.com>
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
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	bindTestStruct struct {
		I           int
		PtrI        *int
		I8          int8
		PtrI8       *int8
		I16         int16
		PtrI16      *int16
		I32         int32
		PtrI32      *int32
		I64         int64
		PtrI64      *int64
		UI          uint
		PtrUI       *uint
		UI8         uint8
		PtrUI8      *uint8
		UI16        uint16
		PtrUI16     *uint16
		UI32        uint32
		PtrUI32     *uint32
		UI64        uint64
		PtrUI64     *uint64
		B           bool
		PtrB        *bool
		F32         float32
		PtrF32      *float32
		F64         float64
		PtrF64      *float64
		S           string
		PtrS        *string
		cantSet     string
		DoesntExist string
		T           Timestamp
		Tptr        *Timestamp
		SA          StringArray
	}
	Timestamp   time.Time
	TA          []Timestamp
	StringArray []string
	Struct      struct {
		Foo string
	}
)

func (t bindTestStruct) GetCantSet() string {
	return t.cantSet
}

func (t *Timestamp) UnmarshalBind(src string) error {
	ts, err := time.Parse(time.RFC3339, src)
	*t = Timestamp(ts)
	return err
}

func (a *StringArray) UnmarshalBind(src string) error {
	*a = StringArray(strings.Split(src, ","))
	return nil
}

func (s *Struct) UnmarshalBind(src string) error {
	*s = Struct{
		Foo: src,
	}
	return nil
}

var values = map[string][]string{
	"I":       {"0"},
	"PtrI":    {"0"},
	"I8":      {"8"},
	"PtrI8":   {"8"},
	"I16":     {"16"},
	"PtrI16":  {"16"},
	"I32":     {"32"},
	"PtrI32":  {"32"},
	"I64":     {"64"},
	"PtrI64":  {"64"},
	"UI":      {"0"},
	"PtrUI":   {"0"},
	"UI8":     {"8"},
	"PtrUI8":  {"8"},
	"UI16":    {"16"},
	"PtrUI16": {"16"},
	"UI32":    {"32"},
	"PtrUI32": {"32"},
	"UI64":    {"64"},
	"PtrUI64": {"64"},
	"B":       {"true"},
	"PtrB":    {"true"},
	"F32":     {"32.5"},
	"PtrF32":  {"32.5"},
	"F64":     {"64.5"},
	"PtrF64":  {"64.5"},
	"S":       {"test"},
	"PtrS":    {"test"},
	"cantSet": {"test"},
	"T":       {"2016-12-06T19:09:05+01:00"},
	"Tptr":    {"2016-12-06T19:09:05+01:00"},
	"ST":      {"bar"},
}

func assertBindTestStruct(t *testing.T, ts *bindTestStruct) {
	assert.Equal(t, 0, ts.I)
	assert.Equal(t, int8(8), ts.I8)
	assert.Equal(t, int16(16), ts.I16)
	assert.Equal(t, int32(32), ts.I32)
	assert.Equal(t, int64(64), ts.I64)
	assert.Equal(t, uint(0), ts.UI)
	assert.Equal(t, uint8(8), ts.UI8)
	assert.Equal(t, uint16(16), ts.UI16)
	assert.Equal(t, uint32(32), ts.UI32)
	assert.Equal(t, uint64(64), ts.UI64)
	assert.Equal(t, true, ts.B)
	assert.Equal(t, float32(32.5), ts.F32)
	assert.Equal(t, float64(64.5), ts.F64)
	assert.Equal(t, "test", ts.S)
	assert.Equal(t, "", ts.GetCantSet())
}

func TestBindbindData(t *testing.T) {
	ts := new(bindTestStruct)
	b := new(defaultBinder)
	b.bindData(ts, values, "form")
	assertBindTestStruct(t, ts)
}

func TestBindSetWithProperType(t *testing.T) {
	ts := new(bindTestStruct)
	typ := reflect.TypeOf(ts).Elem()
	val := reflect.ValueOf(ts).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		if len(values[typeField.Name]) == 0 {
			continue
		}
		val := values[typeField.Name][0]
		err := setWithProperType(typeField.Type.Kind(), val, structField)
		if err != nil {
			t.Fail()
		}
	}
	assertBindTestStruct(t, ts)

	type foo struct {
		Bar bytes.Buffer
	}
	v := &foo{}
	typ = reflect.TypeOf(v).Elem()
	val = reflect.ValueOf(v).Elem()
	if setWithProperType(typ.Field(0).Type.Kind(), "5", val.Field(0)) == nil {
		t.Fail()
	}
}

func TestBindSetFields(t *testing.T) {
	ts := new(bindTestStruct)
	val := reflect.ValueOf(ts).Elem()
	// Int
	if setIntField("5", 0, val.FieldByName("I")) == nil {
		assert.Equal(t, 5, ts.I)
	} else {
		t.Fail()
	}
	if setIntField("", 0, val.FieldByName("I")) == nil {
		assert.Equal(t, 0, ts.I)
	} else {
		t.Fail()
	}

	// Uint
	if setUintField("10", 0, val.FieldByName("UI")) == nil {
		assert.Equal(t, uint(10), ts.UI)
	} else {
		t.Fail()
	}
	if setUintField("", 0, val.FieldByName("UI")) == nil {
		assert.Equal(t, uint(0), ts.UI)
	} else {
		t.Fail()
	}

	// Float
	if setFloatField("15.5", 0, val.FieldByName("F32")) == nil {
		assert.Equal(t, float32(15.5), ts.F32)
	} else {
		t.Fail()
	}
	if setFloatField("", 0, val.FieldByName("F32")) == nil {
		assert.Equal(t, float32(0.0), ts.F32)
	} else {
		t.Fail()
	}

	// Bool
	if setBoolField("true", val.FieldByName("B")) == nil {
		assert.Equal(t, true, ts.B)
	} else {
		t.Fail()
	}
	if setBoolField("", val.FieldByName("B")) == nil {
		assert.Equal(t, false, ts.B)
	} else {
		t.Fail()
	}

	ok, err := unmarshalFieldNonPtr("2016-12-06T19:09:05Z", val.FieldByName("T"))
	if err == nil {
		assert.Equal(t, ok, true)
		assert.Equal(t, Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), ts.T)
	}
}
