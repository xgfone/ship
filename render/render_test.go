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

package render

import (
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
)

func TestSimpleRenderer(t *testing.T) {
	r := SimpleRenderer("plain", "text/plain", func(w io.Writer, v interface{}) error {
		_, err := fmt.Fprintf(w, "%v", v)
		return err
	})

	rec := httptest.NewRecorder()
	r.Render(rec, "plain", 200, "body data")
	fmt.Println(rec.Body.String())

	// Output:
	// body data
}

func ExampleMuxRenderer() {
	type Data struct {
		Key1 int
		Key2 string
	}
	data := Data{Key1: 123, Key2: "abc"}

	mr := NewMuxRenderer()
	mr.Add("json", JSONRenderer())
	mr.Add("jsonpretty", JSONPrettyRenderer())
	mr.Add("xml", XMLRenderer())
	mr.Add("xmlpretty", XMLPrettyRenderer())

	rec := httptest.NewRecorder()
	mr.Render(rec, "json", 200, data)
	fmt.Print(rec.Body.String())

	rec = httptest.NewRecorder()
	mr.Render(rec, "jsonpretty", 200, data)
	fmt.Print(rec.Body.String())

	rec = httptest.NewRecorder()
	mr.Render(rec, "xml", 200, data)
	fmt.Println(rec.Body.String())

	rec = httptest.NewRecorder()
	mr.Render(rec, "xmlpretty", 200, data)
	fmt.Println(rec.Body.String())

	// Output:
	// {"Key1":123,"Key2":"abc"}
	// {
	//     "Key1": 123,
	//     "Key2": "abc"
	// }
	// <Data><Key1>123</Key1><Key2>abc</Key2></Data>
	// <Data>
	//     <Key1>123</Key1>
	//     <Key2>abc</Key2>
	// </Data>
}
