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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var bufpool = sync.Pool{
	New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 1024)) },
}

func getBuffer() *bytes.Buffer    { return bufpool.Get().(*bytes.Buffer) }
func putBuffer(buf *bytes.Buffer) { buf.Reset(); bufpool.Put(buf) }

// OnceRunner is used to run the task only once, which is different from
// sync.Once, the second calling does not wait until the first calling finishes.
type OnceRunner struct {
	done uint32
	task func()
}

// NewOnceRunner returns a new OnceRunner.
func NewOnceRunner(task func()) *OnceRunner { return &OnceRunner{task: task} }

// Run runs the task.
func (r *OnceRunner) Run() {
	if atomic.CompareAndSwapUint32(&r.done, 0, 1) {
		r.task()
	}
}

// CopyNBuffer is the same as io.CopyN, but uses the given buf as the buffer.
//
// If buf is nil or empty, it will make a new one with 2048.
func CopyNBuffer(dst io.Writer, src io.Reader, n int64, buf []byte) (written int64, err error) {
	if len(buf) == 0 {
		buf = make([]byte, 1024)
	}

	// For like byte.Buffer, we maybe grow its capacity to avoid allocating
	// the memory more times.
	if b, ok := dst.(interface{ Grow(int) }); ok && n > 0 {
		if n < 32768 { // 32KB
			b.Grow(int(n))
		} else {
			b.Grow(32768)
		}
	}

	// (xgfone): Fix for compression, such as gzip or deflate.
	if n > 0 {
		src = io.LimitReader(src, n)
	}

	written, err = io.CopyBuffer(dst, src, buf)
	if written == n {
		return n, nil
	} else if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}

	return
}

// DisalbeRedirect is used to disalbe the default redirect behavior
// of http.Client, that's, http.Client won't handle the redirect response
// and just return it to the caller.
func DisalbeRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// InStrings reports whether s is in the string slice ss or not.
func InStrings(s string, ss []string) bool {
	for i, _len := 0, len(ss); i < _len; i++ {
		if s == ss[i] {
			return true
		}
	}
	return false
}

// SplitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func SplitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

var (
	errInvalidTagValue    = errors.New("invalid tag value")
	errNotPointerToStruct = errors.New("the argument must be a pointer to struct")
)

// SetStructFieldToDefault sets the default value of the fields of the struct v
// to the value of the tag "default" of the fields when the field value is ZERO.
//
// If v is not a struct, it does nothing; and not a pointer to struct, panic.
//
// For the type of the field, it only supports some base types as follow:
//   string
//   float32
//   float64
//   int
//   int8
//   int16
//   int32
//   int64
//   uint
//   uint8
//   uint16
//   uint32
//   uint64
//   struct
//   struct slice
//   interface{ SetDefault(_default interface{}) error }
//
// Notice: If the tag value starts with ".", it represents a field name and
// the default value of current field is set to the value of that field.
// But their types must be consistent, or panic.
func SetStructFieldToDefault(v interface{}) (err error) {
	vf := reflect.ValueOf(v)
	switch kind := vf.Kind(); kind {
	case reflect.Ptr:
		vf = vf.Elem()
		if vf.Kind() != reflect.Struct {
			return errNotPointerToStruct
		}
		err = setDefault(vf)
	case reflect.Struct:
		return errNotPointerToStruct
	}

	return
}

type setDefaulter interface {
	SetDefault(_default interface{}) error
}

func setDefault(vf reflect.Value) (err error) {
	vt := vf.Type()
	for i, _len := 0, vt.NumField(); i < _len; i++ {
		fieldv := vf.Field(i)
		if !fieldv.CanSet() {
			continue
		}

		tag := strings.TrimSpace(vt.Field(i).Tag.Get("default"))
		switch v := fieldv.Interface().(type) {
		case string:
			if v == "" && tag != "" {
				fieldv.SetString(tag)
			}
		case int:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int8:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int16:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int32:
			err = setFieldInt(vf, fieldv, int64(v), tag)
		case int64:
			err = setFieldInt(vf, fieldv, v, tag)
		case uint:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint8:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint16:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint32:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case uint64:
			err = setFieldUint(vf, fieldv, v, tag)
		case uintptr:
			err = setFieldUint(vf, fieldv, uint64(v), tag)
		case float32:
			err = setFieldFloat(vf, fieldv, float64(v), tag)
		case float64:
			err = setFieldFloat(vf, fieldv, v, tag)
		case setDefaulter:
			if tag != "" {
				err = v.SetDefault(tag)
			}

		default:
			switch fieldv.Kind() {
			case reflect.Struct:
				err = setDefault(fieldv)
			case reflect.Slice:
				for i, _len := 0, fieldv.Len(); i < _len; i++ {
					if _fieldv := fieldv.Index(i); _fieldv.Kind() == reflect.Struct {
						if err = setDefault(_fieldv); err != nil {
							return
						}
					}
				}
			}
		}

		if err != nil {
			return
		}
	}

	return
}

func setFieldInt(structv, fieldv reflect.Value, v int64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseInt(tag, 10, 64); err == nil {
			fieldv.SetInt(v)
		}
	}
	return
}

func setFieldUint(structv, fieldv reflect.Value, v uint64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseUint(tag, 10, 64); err == nil {
			fieldv.SetUint(v)
		}
	}
	return
}

func setFieldFloat(structv, fieldv reflect.Value, v float64, tag string) (err error) {
	if v == 0 && tag != "" {
		if tag[0] == '.' {
			if tag = tag[1:]; tag == "" {
				return errInvalidTagValue
			}
			fieldv.Set(structv.FieldByName(tag))
		} else if v, err = strconv.ParseFloat(tag, 64); err == nil {
			fieldv.SetFloat(v)
		}
	}
	return
}

// GetText is the same as GetJSON, but get the response body as the string,
// which has no request body but has the response body if successfully.
func GetText(url string) (body string, err error) {
	err = Request(context.Background(), http.MethodGet, url, nil, nil, &body)
	return
}

// GetJSON is the same as RequestJSON, but use the method GET instead,
// which has no request body but has the response body if successfully.
func GetJSON(url string, resp interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodGet, url, nil, nil, resp)
}

// PostJSON is the same as RequestJSON, but use the method POST instead,
// which has the request body and the response body if successfully.
func PostJSON(url string, req interface{}, resp interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodPost, url, nil, req, resp)
}

// PutJSON is the same as RequestJSON, but use the method PUT instead,
// which has the request body but no response body.
func PutJSON(url string, req interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodPut, url, nil, req, nil)
}

// PatchJSON is the same as RequestJSON, but use the method PATCH instead,
// which has the request body and the response body if successfully.
func PatchJSON(url string, req interface{}, resp interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodPatch, url, nil, req, resp)
}

// DeleteJSON is the same as RequestJSON, but use the method DELETE instead,
// which may has the request body and the response body.
func DeleteJSON(url string, req interface{}, resp interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodDelete, url, nil, req, resp)
}

// HeadJSON is the same as RequestJSON, but use the method HEADE instead,
// which has no request or response body.
func HeadJSON(url string) (respHeader http.Header, err error) {
	r := func(r *http.Response) error { respHeader = r.Header; return nil }
	err = RequestJSON(context.Background(), http.MethodHead, url, nil, nil, r)
	return
}

// OptionsJSON is the same as RequestJSON, but use the method OPTIONS instead,
// which has no request body but has the response body if successfully.
func OptionsJSON(url string, resp interface{}) (err error) {
	return RequestJSON(context.Background(), http.MethodOptions, url, nil, nil, resp)
}

// RequestJSON is the same as Request, but encodes the request
// and decodes response body as JSON.
func RequestJSON(ctx context.Context, method, url string, reqHeader http.Header,
	reqBody, respBody interface{}) error {
	if len(reqHeader) == 0 {
		reqHeader = http.Header{
			HeaderAccept:      MIMEApplicationJSONCharsetUTF8s,
			HeaderContentType: MIMEApplicationJSONCharsetUTF8s,
		}
	} else {
		reqHeader.Set(HeaderAccept, MIMEApplicationJSONCharsetUTF8)
		reqHeader.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	}

	switch data := reqBody.(type) {
	case nil:
	case []byte:
		reqBody = bytes.NewBuffer(data)
	case string:
		reqBody = bytes.NewBufferString(data)
	case io.Reader:
		reqBody = data
	default:
		buf := getBuffer()
		defer putBuffer(buf)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return NewHTTPClientError(method, url, 0, err)
		}
		reqBody = buf
	}

	if respBody != nil {
		oldRespBody := respBody
		respBody = func(r io.Reader) error {
			return json.NewDecoder(r).Decode(oldRespBody)
		}
	}

	return Request(ctx, method, url, reqHeader, reqBody, respBody)
}

// Request sends the http request and parses the response body into respBody.
//
// reqBody must be one of types:
//   - nil
//   - []byte
//   - string
//   - io.Reader
//   - func() (io.Reader, error)
//
// respBody must be one of types:
//   - nil: ignore the response body.
//   - *[]byte: read the response body and puts it into respBody as []byte.
//   - *string: read the response body and puts it into respBody as string.
//   - io.Writer: copy the response body into the given writer.
//   - xml.Unmarshaler: read and parse the response body as the XML.
//   - json.Unmarshaler: read and parse the response body as the JSON.
//   - func(io.Reader) error: call the function with the response body.
//   - func(*bytes.Buffer) error: read the response body into the buffer and call the function.
//   - func(*http.Response) error: call the function with the response.
//
// Notice: if the encoding of the response body is gzip, it will decode it firstly.
func Request(ctx context.Context, method, url string, reqHeader http.Header,
	reqBody, respBody interface{}) (err error) {
	var body io.Reader
	switch data := reqBody.(type) {
	case nil:
	case []byte:
		body = bytes.NewBuffer(data)
	case string:
		body = bytes.NewBufferString(data)
	case io.Reader:
		body = data
	case func() (io.Reader, error):
		if body, err = data(); err != nil {
			return NewHTTPClientError(method, url, 0, err)
		}
	default:
		panic(fmt.Errorf("unknown request body type '%T'", reqBody))
	}

	req, err := NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return NewHTTPClientError(method, url, 0, err)
	}

	if len(reqHeader) != 0 {
		req.Header = reqHeader
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NewHTTPClientError(method, url, 0, err)
	}
	defer resp.Body.Close()

	respbody := resp.Body
	if resp.Header.Get(HeaderContentEncoding) == "gzip" {
		var reader *gzip.Reader
		if reader, err = gzip.NewReader(resp.Body); err != nil {
			return NewHTTPClientError(method, url, resp.StatusCode, err)
		}
		respbody = reader
	}

	if resp.StatusCode >= 400 {
		data, _ := readAll(respbody)
		return NewHTTPClientError(method, url, resp.StatusCode, nil, data)
	}

	switch r := respBody.(type) {
	case nil:
	case *[]byte:
		buf := getBuffer()
		_, err = io.CopyBuffer(buf, respbody, make([]byte, 1024))
		*r = make([]byte, buf.Len())
		copy(*r, buf.Bytes())
		putBuffer(buf)
	case *string:
		*r, err = readAll(respbody)
	case xml.Unmarshaler:
		err = xml.NewDecoder(respbody).Decode(r)
	case json.Unmarshaler:
		err = json.NewDecoder(respbody).Decode(r)
	case io.Writer:
		_, err = io.CopyBuffer(r, respbody, make([]byte, 1024))
	case func(io.Reader) error:
		err = r(respbody)
	case func(*http.Response) error:
		resp.Body = respbody
		err = r(resp)
	case func(*bytes.Buffer) error:
		b := getBuffer()
		if _, err = io.CopyBuffer(b, respbody, make([]byte, 1024)); err == nil {
			err = r(b)
		}
		putBuffer(b)
	default:
		panic(fmt.Errorf("unknown response body type '%T'", respBody))
	}

	if err != nil {
		err = NewHTTPClientError(method, url, resp.StatusCode, err)
	}

	return err
}

func readAll(r io.Reader) (data string, err error) {
	buf := getBuffer()
	_, err = io.CopyBuffer(buf, r, make([]byte, 1024))
	data = buf.String()
	putBuffer(buf)
	return
}
