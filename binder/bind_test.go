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

package binder_test

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

	"github.com/xgfone/ship"
)

func testMatchPanic(t *testing.T, f func(), err error) {
	defer func() {
		perr := recover()

		if perr != nil && err != nil {
			if perr.(error).Error() != err.Error() {
				t.Logf("%s <--> %s", perr, err)
				t.Fail()
			}
		} else if perr != nil || err != nil {
			t.Logf("%s <--> %s", perr, err)
			t.Fail()
		}
	}()

	f()
}

func isEqual(v1, v2 interface{}) bool {
	if v1 == nil || v2 == nil {
		return v1 == v2
	}

	return reflect.DeepEqual(v1, v2)
}

func isType(v1, v2 interface{}) bool {
	return isEqual(reflect.TypeOf(v1), reflect.TypeOf(v2))
}

func testEqual(t *testing.T, v1, v2 interface{}) {
	if !isEqual(v1, v2) {
		t.Logf("%+v != %+v", v1, v2)
		t.Fail()
	}
}

func testIsType(t *testing.T, v1, v2 interface{}) {
	if !isType(v1, v2) {
		t.Fail()
	}
}

var defaultHandler = func(ctx ship.Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.Request().Method)); err != nil {
		code := http.StatusInternalServerError
		err = ship.NewHTTPError(code).SetInnerError(err)
	}
	return
}

var idHandler = func(ctx ship.Context) (err error) {
	resp := ctx.Response()
	if _, err = resp.Write([]byte(ctx.URLParamByName("id"))); err != nil {
		code := http.StatusInternalServerError
		err = ship.NewHTTPError(code).SetInnerError(err)
	}
	return
}

var params2Handler = func(ctx ship.Context) (err error) {
	resp := ctx.Response()
	get := ctx.URLParamByName
	if _, err = resp.Write([]byte(get("p1") + "|" + get("p2"))); err != nil {
		code := http.StatusInternalServerError
		err = ship.NewHTTPError(code).SetInnerError(err)
	}
	return
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func sendTestRequest(method, path string, s *ship.Ship) (int, string) {
	r, _ := http.NewRequest(method, path, nil)
	w := &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}

	s.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}

//////////////////////////////////////////////////////////////////////////////

func testBindOkay(t *testing.T, r io.Reader, ctype string) {
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	ctx := ship.New().NewContext(req, rec)
	req.Header.Set(ship.HeaderContentType, ctype)
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
	ctx := ship.New().NewContext(req, rec)
	req.Header.Set(ship.HeaderContentType, ctype)
	u := new(user)
	err := ctx.Bind(u)

	switch {
	case strings.HasPrefix(ctype, ship.MIMEApplicationJSON),
		strings.HasPrefix(ctype, ship.MIMEApplicationXML),
		strings.HasPrefix(ctype, ship.MIMETextXML),
		strings.HasPrefix(ctype, ship.MIMEApplicationForm),
		strings.HasPrefix(ctype, ship.MIMEMultipartForm):
		if isType(ship.NewHTTPError(200), err) {
			testEqual(t, http.StatusBadRequest, err.(ship.HTTPError).Code())
			testIsType(t, expectedInternal, err.(ship.HTTPError).InnerError())
		}
	default:
		if isType(ship.NewHTTPError(200), err) {
			testEqual(t, ship.ErrUnsupportedMediaType, err)
			testIsType(t, expectedInternal, err.(ship.HTTPError).InnerError())
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

func TestBindJSON(t *testing.T) {
	testBindOkay(t, strings.NewReader(userJSON), ship.MIMEApplicationJSON)
	testBindError(t, strings.NewReader(invalidContent), ship.MIMEApplicationJSON,
		&json.SyntaxError{})
	testBindError(t, strings.NewReader(userJSONInvalidType),
		ship.MIMEApplicationJSON, &json.UnmarshalTypeError{})
}

func TestBindXML(t *testing.T) {
	testBindOkay(t, strings.NewReader(userXML), ship.MIMEApplicationXML)
	testBindError(t, strings.NewReader(invalidContent), ship.MIMEApplicationXML, errors.New(""))
	testBindError(t, strings.NewReader(userXMLConvertNumberError), ship.MIMEApplicationXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), ship.MIMEApplicationXML, &xml.SyntaxError{})
	testBindOkay(t, strings.NewReader(userXML), ship.MIMETextXML)
	testBindError(t, strings.NewReader(invalidContent), ship.MIMETextXML, errors.New(""))
	testBindError(t, strings.NewReader(userXMLConvertNumberError), ship.MIMETextXML, &strconv.NumError{})
	testBindError(t, strings.NewReader(userXMLUnsupportedTypeError), ship.MIMETextXML, &xml.SyntaxError{})
}

func TestBindForm(t *testing.T) {
	testBindOkay(t, strings.NewReader(userForm), ship.MIMEApplicationForm)
	testBindError(t, nil, ship.MIMEApplicationForm, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userForm))
	rec := httptest.NewRecorder()
	ctx := ship.New().NewContext(req, rec)
	req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationForm)
	err := ctx.Bind(&[]struct{ Field string }{})
	if err == nil {
		t.Fail()
	}
}

func TestBindQueryParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	ctx := ship.New().NewContext(req, rec)
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
	ctx := ship.New().NewContext(req, rec)
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
	ctx := ship.New().NewContext(req, rec)
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
	ctx := ship.New().NewContext(req, rec)
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
	ctx := ship.New().NewContext(req, rec)
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
	testBindError(t, strings.NewReader(invalidContent), ship.MIMEApplicationJSON,
		&json.SyntaxError{})
}

func TestBindUnmarshalTypeError(t *testing.T) {
	body := bytes.NewBufferString(`{ "id": "text" }`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(ship.HeaderContentType, ship.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	ctx := ship.New().NewContext(req, rec)
	u := new(user)

	err := ctx.Bind(u)
	he := ship.NewHTTPError(http.StatusBadRequest,
		"Unmarshal type error: expected=int, got=string, offset=14")
	he = he.SetInnerError(err.(ship.HTTPError))

	testEqual(t, he.Error(), err.Error())
}
