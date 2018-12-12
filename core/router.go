package core

// WildcardParam is used to represent the name of the wildcard parameter.
//
// Notice: for the universality, the router implementation shoud use
// the constant as the name of the wildcard parameter.
const WildcardParam = "*"

// Router stands for a router management.
type Router interface {
	// Generate a URL by the url name and parameters.
	URL(name string, params ...interface{}) string

	// Add a route with name, method , path and handler,
	// and return the number of the parameters if there are the parameters
	// in the route. Or return 0.
	//
	// If the router does not support the parameter, it should panic.
	//
	// Notice: for keeping consistent, the parameter should start with ":"
	// or "*". ":" stands for a single parameter, and "*" stands for
	// a wildcard parameter.
	Add(name string, method string, path string, handler Handler) (paramNum int)

	// Find a route handler by the method and path of the request.
	//
	// Return nil if the route does not exist.
	//
	// If the route has more than one parameter, the name and value
	// of the parameters should be stored `pnames` and `pvalues` respectively.
	Find(method string, path string, pnames []string, pvalues []string) (handler Handler)

	// Traverse each route.
	Each(func(name string, method string, path string))
}
