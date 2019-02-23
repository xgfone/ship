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

package ship

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//////////////////////////////////////////////////////////////////////////////

func testBindOkay(t *testing.T, r io.Reader, ctype string) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := ctx.Bind(u)
	if err == nil {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func testBindError(t *testing.T, r io.Reader, ctype string, expectedInternal error) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := ctx.Bind(u)

	assert.IsType(t, expectedInternal, err)
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
	testBindOkay(t, strings.NewReader(userJSON), MIMEApplicationJSON)
	testBindError(t, strings.NewReader(invalidContent), MIMEApplicationJSON,
		&json.SyntaxError{})
	testBindError(t, strings.NewReader(userJSONInvalidType),
		MIMEApplicationJSON, &json.UnmarshalTypeError{})
}

func TestBindXML(t *testing.T) {
	testBindOkay(t, strings.NewReader(userXML), MIMEApplicationXML)
	testBindError(t, strings.NewReader(invalidContent), MIMEApplicationXML, ErrMissingContentType)
	testBindError(t, strings.NewReader(userXMLConvertNumberError), MIMEApplicationXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), MIMEApplicationXML, &xml.SyntaxError{})
	testBindOkay(t, strings.NewReader(userXML), MIMETextXML)
	testBindError(t, strings.NewReader(invalidContent), MIMETextXML, ErrMissingContentType)
	testBindError(t, strings.NewReader(userXMLConvertNumberError), MIMETextXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), MIMETextXML, &xml.SyntaxError{})
}

func TestBindForm(t *testing.T) {
	testBindOkay(t, strings.NewReader(userForm), MIMEApplicationForm)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userForm))
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	err := ctx.Bind(&[]struct{ Field string }{})
	if err == nil {
		t.Fail()
	}
}

func TestBindQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	u := new(user)
	err := ctx.BindQuery(u)
	if err == nil {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindQueryParamsCaseInsensitive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ID=1&NAME=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	u := new(user)
	err := ctx.BindQuery(u)
	if err == nil {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindQueryParamsCaseSensitivePrioritized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&ID=2&NAME=Jon+Snow&name=Jon+Doe", nil)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	u := new(user)
	err := ctx.BindQuery(u)
	if err == nil {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Doe", u.Name)
	} else {
		t.Fail()
	}
}

func TestBindUnmarshalBind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/?ts=2016-12-06T19:09:05Z&sa=one,two,three&ta=2016-12-06T19:09:05Z&ta=2016-12-06T19:09:05Z&ST=baz",
		nil)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	result := struct {
		T  Timestamp   `query:"ts"`
		TA []Timestamp `query:"ta"`
		SA StringArray `query:"sa"`
		ST Struct
	}{}
	err := ctx.Bind(&result)
	ts := Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC))

	if err == nil {
		assert.Equal(t, ts, result.T)
		assert.Equal(t, StringArray([]string{"one", "two", "three"}), result.SA)
		assert.Equal(t, []Timestamp{ts, ts}, result.TA)
		assert.Equal(t, Struct{"baz"}, result.ST)
	}
}

func TestBindUnmarshalBindPtr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?ts=2016-12-06T19:09:05Z", nil)
	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	result := struct {
		Tptr *Timestamp `query:"ts"`
	}{}
	err := ctx.BindQuery(&result)
	if err == nil {
		assert.Equal(t, Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), *result.Tptr)
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

func TestBindUnmarshalTypeError(t *testing.T) {
	body := bytes.NewBufferString(`{ "id": "text" }`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := New().NewContext(req, rec)
	u := new(user)

	err := ctx.Bind(u)
	assert.Equal(t, "json: cannot unmarshal string into Go struct field user.id of type int", err.Error())
}
