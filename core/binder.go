package core

// Binder is the interface to bind the value to v from ctx.
type Binder interface {
	Bind(ctx Context, v interface{}) error
}
