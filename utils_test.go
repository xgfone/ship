// Copyright 2020 xgfone
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

package ship

import (
	"fmt"
	"time"
)

func ExampleSetStructFieldToDefault() {
	type Struct struct {
		InnerInt int `default:"123"`
		_        int `default:"-"`
	}

	type S struct {
		Ignore  bool    `default:"true"`
		Int     int     `default:"123"`
		Int8    int8    `default:"123"`
		Int16   int16   `default:"123"`
		Int32   int32   `default:"123"`
		Int64   int64   `default:"123"`
		Uint    uint    `default:"123"`
		Uint8   uint8   `default:"123"`
		Uint16  uint16  `default:"123"`
		Uint32  uint32  `default:"123"`
		Uint64  uint64  `default:"123"`
		Uintptr uintptr `default:"123"`
		Float32 float32 `default:"1.2"`
		Float64 float64 `default:"1.2"`
		FloatN  float64 `default:".Float64"` // Set the default value to other field
		String  string  `default:"abc"`
		Struct  Struct
		Structs []Struct
		_       int `default:"-"`

		DurationInt time.Duration `default:"1000"`
		DurationStr time.Duration `default:"2s"`
		TimeInt     time.Time     `default:"1618059388"`
		TimeStr     time.Time     `default:"2021-04-10T12:56:28Z"`

		NoneP       *int
		IntPtr      *int           `default:"456"`
		TimePtr     *time.Time     `default:"2021-04-10T12:56:28Z"`
		DurationPtr *time.Duration `default:"3s"`
	}

	s := S{Structs: make([]Struct, 2)}
	err := SetStructFieldToDefault(&s)
	fmt.Println(err)

	fmt.Println(s.Ignore)
	fmt.Println(s.Int)
	fmt.Println(s.Int8)
	fmt.Println(s.Int16)
	fmt.Println(s.Int32)
	fmt.Println(s.Int64)
	fmt.Println(s.Uint)
	fmt.Println(s.Uint8)
	fmt.Println(s.Uint16)
	fmt.Println(s.Uint32)
	fmt.Println(s.Uint64)
	fmt.Println(s.Uintptr)
	fmt.Println(s.Float32)
	fmt.Println(s.Float64)
	fmt.Println(s.FloatN)
	fmt.Println(s.String)
	fmt.Println(s.Struct.InnerInt)
	fmt.Println(s.Structs[0].InnerInt)
	fmt.Println(s.Structs[1].InnerInt)
	fmt.Println(s.DurationInt)
	fmt.Println(s.DurationStr)
	fmt.Println(s.TimeInt.UTC().Format(time.RFC3339))
	fmt.Println(s.TimeStr.UTC().Format(time.RFC3339))
	fmt.Println(s.NoneP == nil)
	fmt.Println(*s.IntPtr)
	fmt.Println(s.TimePtr.UTC().Format(time.RFC3339))
	fmt.Println(*s.DurationPtr)

	// Output:
	// <nil>
	// false
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 123
	// 1.2
	// 1.2
	// 1.2
	// abc
	// 123
	// 123
	// 123
	// 1s
	// 2s
	// 2021-04-10T12:56:28Z
	// 2021-04-10T12:56:28Z
	// true
	// 456
	// 2021-04-10T12:56:28Z
	// 3s
}

func ExampleSplitHostPort() {
	var host, port string

	host, port = SplitHostPort("www.example.com")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("www.example.com:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort(":80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("1.2.3.4:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("[fe80::1122:3344:5566:7788]")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	host, port = SplitHostPort("[fe80::1122:3344:5566:7788]:80")
	fmt.Printf("Host: %s, Port: %s#\n", host, port)

	// Output:
	// Host: www.example.com, Port: #
	// Host: www.example.com, Port: 80#
	// Host: , Port: 80#
	// Host: 1.2.3.4, Port: 80#
	// Host: fe80::1122:3344:5566:7788, Port: #
	// Host: fe80::1122:3344:5566:7788, Port: 80#
}
