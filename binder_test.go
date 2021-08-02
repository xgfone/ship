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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"testing"
)

type binderTestInfo struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
}

func TestMuxBinder(t *testing.T) {
	mb := NewMuxBinder()
	mb.Add(MIMEApplicationJSON, JSONBinder())
	mb.Add(MIMEApplicationXML, XMLBinder())
	mb.Add(MIMEApplicationForm, FormBinder(MaxMemoryLimit))

	data := map[string]string{
		"username": "xgfone",
		"password": "123456",
	}

	// Test JSON Binder
	jsonbuf := bytes.NewBuffer(nil)
	json.NewEncoder(jsonbuf).Encode(data)
	testBinder(t, mb, MIMEApplicationJSON, jsonbuf)

	// Test XML Binder
	xmlbuf := bytes.NewBuffer(nil)
	xml.NewEncoder(xmlbuf).Encode(binderTestInfo{
		Username: "xgfone",
		Password: "123456",
	})
	testBinder(t, mb, MIMEApplicationXML, xmlbuf)

	// Test Form Binder
	forms := make(url.Values, len(data))
	for k, v := range data {
		forms[k] = []string{v}
	}
	formbuf := bytes.NewBufferString(forms.Encode())
	testBinder(t, mb, MIMEApplicationForm, formbuf)
}

func testBinder(t *testing.T, mb *MuxBinder, ct string, body io.Reader) {
	req, _ := http.NewRequest("POST", "http://127.0.0.1", body)
	req.Header.Set(HeaderContentType, ct)

	expect := binderTestInfo{Username: "xgfone", Password: "123456"}
	var result binderTestInfo
	if err := mb.Bind(&result, req); err != nil {
		t.Error(err)
	} else if result != expect {
		t.Errorf("expect '%v', but got '%v'", expect, result)
	}
}
