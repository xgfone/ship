package render

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

import (
	"encoding/xml"

	"github.com/xgfone/ship/core"
)

// XML returns a XML renderer.
//
// Example
//
//     renderer := XML()
//     renderer.Render(ctx, "xml", code, data)
//
// Notice: the default marshaler won't add the XML header.
// and the renderer name must be "xml".
func XML(marshal ...Marshaler) core.Renderer {
	encode := xml.Marshal
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("xml", "application/xml; charset=UTF-8", encode)
}

// XMLPretty returns a pretty XML renderer.
//
// Example
//
//     renderer := XMLPretty("    ")
//     renderer.Render(ctx, "xmlpretty", code, data)
//
//     # Or appoint a specific Marshaler.
//     renderer := XMLPretty("", func(v interface{}) ([]byte, error) {
//         return xml.MarshalIndent(v, "", "  ")
//     })
//     renderer.Render(ctx, "xmlpretty", code, data)
//
// Notice: the default marshaler won't add the XML header,
// and the renderer name must be "xmlpretty".
func XMLPretty(indent string, marshal ...Marshaler) core.Renderer {
	encode := func(v interface{}) ([]byte, error) { return xml.MarshalIndent(v, "", indent) }
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("xmlpretty", "application/xml; charset=UTF-8", encode)
}
