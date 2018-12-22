// Copyright 2018 xgfone <xgfone@126.com>
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

package core

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Context stands for a context interface.
type Context interface {
	// Report whether the response has been sent.
	IsResponse() bool

	// Find the registered router handler by the method and path of the request.
	//
	// Return nil if not found.
	FindHandler(method string, path string) Handler

	NotFoundHandler() Handler

	Request() *http.Request
	Response() http.ResponseWriter
	SetResponse(http.ResponseWriter)

	IsTLS() bool
	IsDebug() bool
	IsAjax() bool
	IsWebSocket() bool

	Header(name string) (value string)
	SetHeader(name string, value string)

	URLParams() map[string]string
	URLParamValues() []string
	URLParamByName(name string) (value string)

	Scheme() string
	RealIP() string
	ContentType() string

	QueryParam(name string) (value string)
	QueryParams() url.Values
	QueryRawString() string

	FormParams() (url.Values, error)
	FormValue(name string) (value string)
	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)

	Cookies() []*http.Cookie
	Cookie(name string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)

	// Get and Set are used to store the key-value information about the context.
	Store() map[string]interface{}
	Get(key string) (value interface{})
	Set(key string, value interface{})
	Del(key string)

	Logger() Logger
	URL(name string, params ...interface{}) string

	Bind(v interface{}) error
	BindQuery(v interface{}) error

	Render(name string, code int, data interface{}) error

	NoContent(code int) error
	Redirect(code int, toURL string) error
	String(code int, data string) error
	Blob(code int, contentType string, b []byte) error

	HTML(code int, html string) error
	HTMLBlob(code int, b []byte) error

	JSON(code int, i interface{}) error
	JSONBlob(code int, b []byte) error
	JSONPretty(code int, i interface{}, indent string) error

	JSONP(code int, callback string, i interface{}) error
	JSONPBlob(code int, callback string, b []byte) error

	XML(code int, i interface{}) error
	XMLBlob(code int, b []byte) error
	XMLPretty(code int, i interface{}, indent string) error

	File(file string) error
	Inline(file string, name string) error
	Attachment(file string, name string) error
	Stream(code int, contentType string, r io.Reader) error
}
