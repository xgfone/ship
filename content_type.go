// Copyright 2021 xgfone
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
	"net/http"
)

var (
	mimeApplicationJSONs                  = []string{MIMEApplicationJSON}
	mimeApplicationJSONCharsetUTF8s       = []string{MIMEApplicationJSONCharsetUTF8}
	mimeApplicationJavaScripts            = []string{MIMEApplicationJavaScript}
	mimeApplicationJavaScriptCharsetUTF8s = []string{MIMEApplicationJavaScriptCharsetUTF8}
	mimeApplicationXMLs                   = []string{MIMEApplicationXML}
	mimeApplicationXMLCharsetUTF8s        = []string{MIMEApplicationXMLCharsetUTF8}
	mimeTextXMLs                          = []string{MIMETextXML}
	mimeTextXMLCharsetUTF8s               = []string{MIMETextXMLCharsetUTF8}
	mimeApplicationForms                  = []string{MIMEApplicationForm}
	mimeApplicationProtobufs              = []string{MIMEApplicationProtobuf}
	mimeApplicationMsgpacks               = []string{MIMEApplicationMsgpack}
	mimeTextHTMLs                         = []string{MIMETextHTML}
	mimeTextHTMLCharsetUTF8s              = []string{MIMETextHTMLCharsetUTF8}
	mimeTextPlains                        = []string{MIMETextPlain}
	mimeTextPlainCharsetUTF8s             = []string{MIMETextPlainCharsetUTF8}
	mimeMultipartForms                    = []string{MIMEMultipartForm}
	mimeOctetStreams                      = []string{MIMEOctetStream}
)

var contenttypes = map[string][]string{}

// AddContentTypeMapping add a content type mapping to convert contentType
// to contentTypeSlice, which is used by SetContentType to set the header
// "Content-Type" to contentTypeSlice by contentType to avoid allocating
// the memory.
//
// If contentTypeSlice is empty, it is []string{contentType} by default.
func AddContentTypeMapping(contentType string, contentTypeSlice []string) {
	if contentType == "" {
		panic(fmt.Errorf("the Content-Type is empty"))
	}

	if len(contentTypeSlice) == 0 {
		contentTypeSlice = []string{contentType}
	}

	contenttypes[contentType] = contentTypeSlice
}

// SetContentType sets the header "Content-Type" to ct.
func SetContentType(header http.Header, ct string) {
	var cts []string
	switch ct {
	case "":
		return
	case MIMEApplicationJSON:
		cts = mimeApplicationJSONs
	case MIMEApplicationJSONCharsetUTF8:
		cts = mimeApplicationJSONCharsetUTF8s
	case MIMEApplicationJavaScript:
		cts = mimeApplicationJavaScripts
	case MIMEApplicationJavaScriptCharsetUTF8:
		cts = mimeApplicationJavaScriptCharsetUTF8s
	case MIMEApplicationXML:
		cts = mimeApplicationXMLs
	case MIMEApplicationXMLCharsetUTF8:
		cts = mimeApplicationXMLCharsetUTF8s
	case MIMETextXML:
		cts = mimeTextXMLs
	case MIMETextXMLCharsetUTF8:
		cts = mimeTextXMLCharsetUTF8s
	case MIMEApplicationForm:
		cts = mimeApplicationForms
	case MIMEApplicationProtobuf:
		cts = mimeApplicationProtobufs
	case MIMEApplicationMsgpack:
		cts = mimeApplicationMsgpacks
	case MIMETextHTML:
		cts = mimeTextHTMLs
	case MIMETextHTMLCharsetUTF8:
		cts = mimeTextHTMLCharsetUTF8s
	case MIMETextPlain:
		cts = mimeTextPlains
	case MIMETextPlainCharsetUTF8:
		cts = mimeTextPlainCharsetUTF8s
	case MIMEMultipartForm:
		cts = mimeMultipartForms
	case MIMEOctetStream:
		cts = mimeOctetStreams
	default:
		if ss := contenttypes[ct]; ss != nil {
			cts = ss
		} else {
			header.Set(HeaderContentType, ct)
			return
		}
	}

	header[HeaderContentType] = cts
}
