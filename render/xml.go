package render

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
