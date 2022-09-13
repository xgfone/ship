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
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func BenchmarkBindURLValues(b *testing.B) {
	type T struct {
		Bool    bool    `query:"bool"`
		Int     int     `query:"int"`
		Uint    uint    `query:"uint"`
		String  string  `query:"string"`
		Float64 float64 `query:"float64"`
	}

	var v T
	data := url.Values{
		"bool":    []string{"true"},
		"int":     []string{"123"},
		"uint":    []string{"456"},
		"string":  []string{"abc"},
		"float64": []string{"789"},
	}

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			BindURLValues(&v, data, "query")
		}
	})
}

type paramBinder struct {
	Value int `query:"Value"`
}

func (b *paramBinder) UnmarshalBind(param string) error {
	v, err := strconv.ParseInt(param, 10, 64)
	if err == nil {
		b.Value = int(v)
	}
	return err
}

type AnonymousString string
type AnonymousStruct struct {
	Embed string `query:"Embed"`
}

func TestBindURLValues(t *testing.T) {
	type T struct {
		Bool       bool            `query:"bool"`
		Int        int             `query:"int"`
		Int8       int8            `query:"int8"`
		Int16      int16           `query:"int16"`
		Int32      int32           `query:"int32"`
		Int64      int64           `query:"int64"`
		Uint       uint            `query:"uint"`
		Uint8      uint8           `query:"uint8"`
		Uint16     uint16          `query:"uint16"`
		Uint32     uint32          `query:"uint32"`
		Uint64     uint64          `query:"uint64"`
		String     string          `query:"string"`
		Float32    float32         `query:"float32"`
		Float64    float64         `query:"float64"`
		Duration   time.Duration   `query:"duration"`
		Time       time.Time       `query:"time"`
		Interface1 paramBinder     `query:"interface1"`
		Interface2 BindUnmarshaler `query:"interface2"`

		BindUnmarshaler `query:"anonymous1"`
		paramBinder     `query:"anonymous2"`
		AnonymousStruct `query:"anonymous3"`
		AnonymousString `query:"anonymous4"`

		Ingore int `query:"-"`
		Ptr    *int
		Slice1 []*int
		Slice2 []int
		// Slice3 []BindUnmarshaler // Not Support
	}

	data := url.Values{
		"bool":       []string{"true"},
		"int":        []string{"11"},
		"int8":       []string{"12"},
		"int16":      []string{"13"},
		"int32":      []string{"14"},
		"int64":      []string{"15"},
		"uint":       []string{"21"},
		"uint8":      []string{"22"},
		"uint16":     []string{"23"},
		"uint32":     []string{"24"},
		"uint64":     []string{"25"},
		"string":     []string{"abc"},
		"float32":    []string{"31"},
		"float64":    []string{"32"},
		"duration":   []string{"1s"},
		"time":       []string{"2022-02-10T14:12:02Z"},
		"interface1": []string{"41"},
		"interface2": []string{"42"},
		"anonymous1": []string{"43"},
		"anonymous4": []string{"44"},
		"Embed":      []string{"45"},

		"Ptr":    []string{"51"},
		"Value":  []string{"51"},
		"Ingore": []string{"51"},
		"Slice1": []string{"51", "52"},
		"Slice2": []string{"55", "56"},
	}

	int1, int2 := 51, 52
	result := T{
		Bool:     true,
		Int:      11,
		Int8:     12,
		Int16:    13,
		Int32:    14,
		Int64:    15,
		Uint:     21,
		Uint8:    22,
		Uint16:   23,
		Uint32:   24,
		Uint64:   25,
		String:   "abc",
		Float32:  31,
		Float64:  32,
		Duration: time.Second,
		Time:     time.Date(2022, time.February, 10, 14, 12, 02, 0, time.UTC),

		Interface1:      paramBinder{41},
		Interface2:      &paramBinder{42},
		BindUnmarshaler: &paramBinder{43},
		AnonymousString: "44",
		AnonymousStruct: AnonymousStruct{Embed: "45"},

		paramBinder: paramBinder{int1},
		Ptr:         &int1,
		Slice1:      []*int{&int1, &int2},
		Slice2:      []int{55, 56},
	}

	v := T{Interface2: &paramBinder{}, BindUnmarshaler: &paramBinder{}}
	if err := BindURLValues(&v, data, "query"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v, result) {
		t.Errorf("expect '%+v', but got '%+v'", result, v)
		t.Error(*v.Slice1[0], *v.Slice1[1])
	}
}
