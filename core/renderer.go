package core

import "fmt"

// Renderer is the interface to render the response.
type Renderer interface {
	Render(ctx Context, name string, code int, data interface{}) error
}

type funcRenderer func(Context, string, int, interface{}) error

func (f funcRenderer) Render(ctx Context, name string, code int, data interface{}) error {
	return f(ctx, name, code, data)
}

// RendererFunc converts a function to Renderer.
func RendererFunc(f func(ctx Context, name string, code int, v interface{}) error) Renderer {
	return funcRenderer(f)
}

// SimpleRenderer returns a simple renderer, which is the same as follow:
//
//     b, err := encode(data)
//     if err != nil {
//         return err
//     }
//     return ctx.Blob(code, contentType, b)
//
func SimpleRenderer(name string, contentType string, encode func(interface{}) ([]byte, error)) Renderer {
	return RendererFunc(func(ctx Context, _name string, code int, v interface{}) error {
		if name != _name {
			return fmt.Errorf("not support the renderer named '%s'", _name)
		}
		b, err := encode(v)
		if err != nil {
			return err
		}
		return ctx.Blob(code, contentType, b)
	})
}
