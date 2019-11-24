// The MIT License (MIT)
//
// Copyright (c) 2018 xgfone
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
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xgfone/ship/v2/herror"
)

var testMuxBinder = NewMuxBinder()

func init() {
	testMuxBinder.Add("application/json", JSONBinder())
	testMuxBinder.Add("application/xml", XMLBinder())
	testMuxBinder.Add("text/xml", XMLBinder())
	testMuxBinder.Add("multipart/form-data", FormBinder(1024))
	testMuxBinder.Add("application/x-www-form-urlencoded", FormBinder(1024))
}

//////////////////////////////////////////////////////////////////////////////

func objectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

func testBindOkay(t *testing.T, r io.Reader, ctype string) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	req.Header.Set("Content-Type", ctype)
	u := new(user)
	if err := testMuxBinder.Bind(req, u); err == nil {
		if u.ID != 1 {
			t.Errorf("ID: expect %d, got %d", 1, u.ID)
		} else if u.Name != "Jon Snow" {
			t.Errorf("Name: expect '%s', got '%s'", "Jon Snow", u.Name)
		}
	} else {
		t.Error(err)
	}
}

func testBindError(t *testing.T, r io.Reader, ctype string, expectedInternal error) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	req.Header.Set("Content-Type", ctype)
	err := testMuxBinder.Bind(req, new(user))
	if !objectsAreEqual(reflect.TypeOf(expectedInternal), reflect.TypeOf(err)) {
		t.Fail()
	}
}

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

type user struct {
	ID   int    `json:"id" xml:"id" form:"id" query:"id"`
	Name string `json:"name" xml:"name" form:"name" query:"name"`
}

const (
	userJSON                    = `{"id":1,"name":"Jon Snow"}`
	userXML                     = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm                    = `id=1&name=Jon Snow`
	invalidContent              = "invalid content"
	userJSONInvalidType         = `{"id":"1","name":"Jon Snow"}`
	userXMLConvertNumberError   = `<user><id>Number one</id><name>Jon Snow</name></user>`
	userXMLUnsupportedTypeError = `<user><>Number one</><name>Jon Snow</name></user>`
)

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

func TestBindJSON(t *testing.T) {
	testBindOkay(t, strings.NewReader(userJSON), "application/json")
	testBindError(t, strings.NewReader(invalidContent), "application/json",
		&json.SyntaxError{})
	testBindError(t, strings.NewReader(userJSONInvalidType),
		"application/json", &json.UnmarshalTypeError{})
}

func TestBindXML(t *testing.T) {
	testBindOkay(t, strings.NewReader(userXML), "application/xml")
	testBindError(t, strings.NewReader(invalidContent), "application/xml", herror.ErrMissingContentType)
	testBindError(t, strings.NewReader(userXMLConvertNumberError), "application/xml", &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), "application/xml", &xml.SyntaxError{})
	testBindOkay(t, strings.NewReader(userXML), "text/xml")
	testBindError(t, strings.NewReader(invalidContent), "text/xml", herror.ErrMissingContentType)
	testBindError(t, strings.NewReader(userXMLConvertNumberError), "text/xml", &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), "text/xml", &xml.SyntaxError{})
}

func TestBindForm(t *testing.T) {
	testBindOkay(t, strings.NewReader(userForm), "application/x-www-form-urlencoded")
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userForm))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := testMuxBinder.Bind(req, &[]struct{ Field string }{})
	if err == nil {
		t.Fail()
	}
}

func TestBindQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon+Snow", nil)
	u := new(user)
	if err := QueryBinder().Bind(req, u); err != nil {
		t.Error(err)
	} else if u.ID != 1 {
		t.Fail()
	} else if u.Name != "Jon Snow" {
		t.Errorf("Name: expect '%s', got '%s'", "Jon Snow", u.Name)
	}
}

func TestBindQueryParamsCaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ID=1&NAME=Jon+Snow", nil)
	u := new(user)
	if err := QueryBinder().Bind(req, u); err != nil {
		t.Error(err)
	} else if u.ID != 1 {
		t.Fail()
	} else if u.Name != "Jon Snow" {
		t.Errorf("Name: expect '%s', got '%s'", "Jon Snow", u.Name)
	}
}

func TestBindQueryParamsCaseSensitivePrioritized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&ID=2&NAME=Jon+Snow&name=Jon+Doe", nil)
	u := new(user)
	if err := QueryBinder().Bind(req, u); err != nil {
		t.Error(err)
	} else if u.ID != 1 {
		t.Fail()
	} else if u.Name != "Jon Doe" {
		t.Errorf("Name: expect '%s', got '%s'", "Jon Doe", u.Name)
	}
}

func TestBindUnmarshalBind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/?ts=2016-12-06T19:09:05Z&sa=one,two,three&ta=2016-12-06T19:09:05Z&ta=2016-12-06T19:09:05Z&ST=baz",
		nil)
	result := struct {
		T  Timestamp   `query:"ts"`
		TA []Timestamp `query:"ta"`
		SA StringArray `query:"sa"`
		ST Struct
	}{}

	ts := Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC))
	if err := QueryBinder().Bind(req, &result); err != nil {
		t.Error(err)
	} else if ts != result.T {
		t.Errorf("expect %v, got %v", ts, result.T)
	} else if len(result.SA) != 3 || len(result.TA) != 2 {
		t.Fail()
	} else if result.ST.Foo != "baz" {
		t.Errorf("expect '%v', got '%v'", result.ST.Foo, "baz")
	}
}

func TestBindUnmarshalBindPtr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ts=2016-12-06T19:09:05Z", nil)
	result := struct {
		Tptr *Timestamp `query:"ts"`
	}{}
	if err := QueryBinder().Bind(req, &result); err != nil {
		t.Error(err)
	} else if v := Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)); *result.Tptr != v {
		t.Errorf("expect '%v', got '%v'", v, *result.Tptr)
	}
}

func TestBindMultipartForm(t *testing.T) {
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	mw.WriteField("id", "1")
	mw.WriteField("name", "Jon Snow")
	mw.Close()

	testBindOkay(t, body, mw.FormDataContentType())
}

func TestBindUnsupportedMediaType(t *testing.T) {
	testBindError(t, strings.NewReader(invalidContent), "application/json",
		&json.SyntaxError{})
}

func TestBindUnmarshalTypeError(t *testing.T) {
	body := bytes.NewBufferString(`{ "id": "text" }`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")
	u := new(user)
	s := "json: cannot unmarshal string into Go struct field user.id of type int"
	if e := testMuxBinder.Bind(req, u).Error(); e != s {
		t.Errorf("expect '%s', got '%s'", s, e)
	}
}
