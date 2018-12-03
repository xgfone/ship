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

package ship

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testBindOkay(t *testing.T, r io.Reader, ctype string) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := ctx.Bind(u)
	if err == nil {
		testEqual(t, 1, u.ID)
		testEqual(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func testBindError(t *testing.T, r io.Reader, ctype string, expectedInternal error) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := ctx.Bind(u)

	switch {
	case strings.HasPrefix(ctype, MIMEApplicationJSON),
		strings.HasPrefix(ctype, MIMEApplicationXML),
		strings.HasPrefix(ctype, MIMETextXML),
		strings.HasPrefix(ctype, MIMEApplicationForm),
		strings.HasPrefix(ctype, MIMEMultipartForm):
		if isType(NewHTTPError(200), err) {
			testEqual(t, http.StatusBadRequest, err.(HTTPError).Code())
			testIsType(t, expectedInternal, err.(HTTPError).InnerError())
		}
	default:
		if isType(NewHTTPError(200), err) {
			testEqual(t, ErrUnsupportedMediaType, err)
			testIsType(t, expectedInternal, err.(HTTPError).InnerError())
		}
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

func (t bindTestStruct) GetCantSet() string {
	return t.cantSet
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

func TestBindJSON(t *testing.T) {
	testBindOkay(t, strings.NewReader(userJSON), MIMEApplicationJSON)
	testBindError(t, strings.NewReader(invalidContent), MIMEApplicationJSON, &json.SyntaxError{})
	testBindError(t, strings.NewReader(userJSONInvalidType), MIMEApplicationJSON, &json.UnmarshalTypeError{})
}

func TestBindXML(t *testing.T) {
	testBindOkay(t, strings.NewReader(userXML), MIMEApplicationXML)
	testBindError(t, strings.NewReader(invalidContent), MIMEApplicationXML, errors.New(""))
	testBindError(t, strings.NewReader(userXMLConvertNumberError), MIMEApplicationXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), MIMEApplicationXML, &xml.SyntaxError{})
	testBindOkay(t, strings.NewReader(userXML), MIMETextXML)
	testBindError(t, strings.NewReader(invalidContent), MIMETextXML, errors.New(""))
	testBindError(t, strings.NewReader(userXMLConvertNumberError), MIMETextXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), MIMETextXML, &xml.SyntaxError{})
}

func TestBindForm(t *testing.T) {
	testBindOkay(t, strings.NewReader(userForm), MIMEApplicationForm)
	testBindError(t, nil, MIMEApplicationForm, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userForm))
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	err := ctx.Bind(&[]struct{ Field string }{})
	if err == nil {
		t.Fail()
	}
}

func TestBindQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	u := new(user)
	err := ctx.Bind(u)
	if err == nil {
		testEqual(t, 1, u.ID)
		testEqual(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindQueryParamsCaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ID=1&NAME=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	u := new(user)
	err := ctx.Bind(u)
	if err == nil {
		testEqual(t, 1, u.ID)
		testEqual(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindQueryParamsCaseSensitivePrioritized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&ID=2&NAME=Jon+Snow&name=Jon+Doe", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	u := new(user)
	err := ctx.Bind(u)
	if err == nil {
		testEqual(t, 1, u.ID)
		testEqual(t, "Jon Doe", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindUnmarshalBind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/?ts=2016-12-06T19:09:05Z&sa=one,two,three&ta=2016-12-06T19:09:05Z&ta=2016-12-06T19:09:05Z&ST=baz",
		nil)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	result := struct {
		T  Timestamp   `query:"ts"`
		TA []Timestamp `query:"ta"`
		SA StringArray `query:"sa"`
		ST Struct
	}{}
	err := ctx.Bind(&result)
	ts := Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC))

	if err == nil {
		testEqual(t, ts, result.T)
		testEqual(t, StringArray([]string{"one", "two", "three"}), result.SA)
		testEqual(t, []Timestamp{ts, ts}, result.TA)
		testEqual(t, Struct{"baz"}, result.ST)
	}
}

func TestBindUnmarshalBindPtr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ts=2016-12-06T19:09:05Z", nil)
	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	result := struct {
		Tptr *Timestamp `query:"ts"`
	}{}
	err := ctx.Bind(&result)
	if err == nil {
		testEqual(t, Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), *result.Tptr)
	} else {
		t.Fail()
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
	testBindError(t, strings.NewReader(invalidContent), MIMEApplicationJSON,
		&json.SyntaxError{})
}

func TestBindbindData(t *testing.T) {
	ts := new(bindTestStruct)
	b := new(defaultBinder)
	b.bindData(ts, values, "form")
	assertBindTestStruct(t, ts)
}

func TestBindUnmarshalTypeError(t *testing.T) {
	body := bytes.NewBufferString(`{ "id": "text" }`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := NewContext()
	ctx.SetRouter(NewRouter())
	ctx.SetReqResp(req, rec)
	u := new(user)

	err := ctx.Bind(u)
	he := NewHTTPError(http.StatusBadRequest,
		"Unmarshal type error: expected=int, got=string, field=id, offset=14")
	he = he.SetInnerError(err.(HTTPError))

	testEqual(t, he.Error(), err.Error())
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
		testEqual(t, 5, ts.I)
	} else {
		t.Fail()
	}
	if setIntField("", 0, val.FieldByName("I")) == nil {
		testEqual(t, 0, ts.I)
	} else {
		t.Fail()
	}

	// Uint
	if setUintField("10", 0, val.FieldByName("UI")) == nil {
		testEqual(t, uint(10), ts.UI)
	} else {
		t.Fail()
	}
	if setUintField("", 0, val.FieldByName("UI")) == nil {
		testEqual(t, uint(0), ts.UI)
	} else {
		t.Fail()
	}

	// Float
	if setFloatField("15.5", 0, val.FieldByName("F32")) == nil {
		testEqual(t, float32(15.5), ts.F32)
	} else {
		t.Fail()
	}
	if setFloatField("", 0, val.FieldByName("F32")) == nil {
		testEqual(t, float32(0.0), ts.F32)
	} else {
		t.Fail()
	}

	// Bool
	if setBoolField("true", val.FieldByName("B")) == nil {
		testEqual(t, true, ts.B)
	} else {
		t.Fail()
	}
	if setBoolField("", val.FieldByName("B")) == nil {
		testEqual(t, false, ts.B)
	} else {
		t.Fail()
	}

	ok, err := unmarshalFieldNonPtr("2016-12-06T19:09:05Z", val.FieldByName("T"))
	if err == nil {
		testEqual(t, ok, true)
		testEqual(t, Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), ts.T)
	}
}

func assertBindTestStruct(t *testing.T, ts *bindTestStruct) {
	testEqual(t, 0, ts.I)
	testEqual(t, int8(8), ts.I8)
	testEqual(t, int16(16), ts.I16)
	testEqual(t, int32(32), ts.I32)
	testEqual(t, int64(64), ts.I64)
	testEqual(t, uint(0), ts.UI)
	testEqual(t, uint8(8), ts.UI8)
	testEqual(t, uint16(16), ts.UI16)
	testEqual(t, uint32(32), ts.UI32)
	testEqual(t, uint64(64), ts.UI64)
	testEqual(t, true, ts.B)
	testEqual(t, float32(32.5), ts.F32)
	testEqual(t, float64(64.5), ts.F64)
	testEqual(t, "test", ts.S)
	testEqual(t, "", ts.GetCantSet())
}
