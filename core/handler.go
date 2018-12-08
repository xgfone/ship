package core

// Handler is a handler of the HTTP request.
type Handler func(Context) error

// Middleware stands for a middleware.
type Middleware func(Handler) Handler
