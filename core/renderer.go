package core

// Renderer is the interface to render the response.
type Renderer interface {
	Render(ctx Context, name string, code int, data interface{}) error
}
