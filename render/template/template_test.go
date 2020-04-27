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

package template

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDirLoader(t *testing.T) {
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Error(err)
		return
	}

	loader := NewDirLoader(dir)
	files, err := loader.LoadAll()
	if err != nil {
		t.Error(err)
	} else {
		for _, file := range files {
			switch name := file.Name(); name {
			case "template.go":
			case "template_test.go":
			case tmplname:
			default:
				t.Error(name)
			}
		}
	}
}

var tmplname = "__html_template_test__.tmpl"
var htmlTmpl = `<!DOCTYPE html>
<html>
	<head>Test</head>
	<body>
		{{ . }}
	</body>
</html>
`
var htmpresp = `<!DOCTYPE html>
<html>
	<head>Test</head>
	<body>
		This is the content.
	</body>
</html>
`

func TestNewHTMLTemplateRender(t *testing.T) {
	err := ioutil.WriteFile(tmplname, []byte(htmlTmpl), 0600)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tmplname)

	dir, err := filepath.Abs(".")
	if err != nil {
		t.Error(err)
		return
	}

	loader := NewDirLoaderWithFilter(func(filename string) bool {
		return !strings.HasSuffix(filename, ".tmpl")
	}, dir)

	files, err := loader.LoadAll()
	if err != nil {
		t.Error(err)
		return
	}
	for _, file := range files {
		if file.Name() != tmplname {
			t.Error(file.Name())
		}
	}

	r := NewHTMLTemplateRender(loader)
	rec := httptest.NewRecorder()
	err = r.Render(rec, tmplname, 200, "This is the content.")
	if err != nil {
		t.Error(err)
	} else if body := rec.Body.String(); body != htmpresp {
		t.Error(body)
	}
}
