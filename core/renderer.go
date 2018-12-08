package core

import "io"

// Renderer is the interface to render the response.
type Renderer interface {
	Render(ctx Context, w io.Writer, name string, code int, data interface{}) error
}
