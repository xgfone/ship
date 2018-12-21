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

package render

import (
	"encoding/json"

	"github.com/xgfone/ship/core"
)

// JSON returns a JSON renderer.
//
// Example
//
//     renderer := JSON()
//     renderer.Render(ctx, "json", code, data)
//
// Notice: the renderer name must be "json".
func JSON(marshal ...Marshaler) core.Renderer {
	encode := json.Marshal
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("json", "application/json; charset=UTF-8", encode)
}

// JSONPretty returns a pretty JSON renderer.
//
// Example
//
//     renderer := JSONPretty("    ")
//     renderer.Render(ctx, "jsonpretty", code, data)
//
//     # Or appoint a specific Marshaler.
//     renderer := JSONPretty("", func(v interface{}) ([]byte, error) {
//         return json.MarshalIndent(v, "", "  ")
//     })
//     renderer.Render(ctx, "jsonpretty", code, data)
//
// Notice: the renderer name must be "jsonpretty".
func JSONPretty(indent string, marshal ...Marshaler) core.Renderer {
	encode := func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", indent) }
	if len(marshal) > 0 && marshal[0] != nil {
		encode = marshal[0]
	}

	return SimpleRenderer("jsonpretty", "application/json; charset=UTF-8", encode)
}
