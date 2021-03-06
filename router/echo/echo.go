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

// Package echo supplies a customized Router implementation by referring to
// github.com/labstack/echo.
package echo

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/xgfone/ship/v4/router"
)

// PROPFIND Method can be used on collection and property resources.
const PROPFIND = "PROPFIND"

// REPORT Method can be used to get information about a resource, see rfc 3253
const REPORT = "REPORT"

// removeTrailingSlash is used to replace the stdlib fucntion strings.TrimRight
// to improve the performance when finding the handler by the route path.
func removeTrailingSlash(path string) string {
	_len := len(path) - 1

	var i int
	for i = _len; i >= 0; i-- {
		if path[i] != '/' {
			break
		}
	}

	if i == _len {
		return path
	} else if i < 0 {
		return ""
	}
	return path[:i+1]
}

type methodHandler struct {
	get      interface{}
	put      interface{}
	post     interface{}
	head     interface{}
	patch    interface{}
	trace    interface{}
	delete   interface{}
	options  interface{}
	connect  interface{}
	propfind interface{}
	report   interface{}
}

func newMethodHandler() *methodHandler { return &methodHandler{} }

func (mh *methodHandler) Routes() []router.Route {
	routes := make([]router.Route, 0, 11)
	if mh.get != nil {
		routes = append(routes, router.Route{Handler: mh.get, Method: http.MethodGet})
	}
	if mh.put != nil {
		routes = append(routes, router.Route{Handler: mh.put, Method: http.MethodPut})
	}
	if mh.post != nil {
		routes = append(routes, router.Route{Handler: mh.post, Method: http.MethodPost})
	}
	if mh.head != nil {
		routes = append(routes, router.Route{Handler: mh.head, Method: http.MethodHead})
	}
	if mh.patch != nil {
		routes = append(routes, router.Route{Handler: mh.patch, Method: http.MethodPatch})
	}
	if mh.trace != nil {
		routes = append(routes, router.Route{Handler: mh.trace, Method: http.MethodTrace})
	}
	if mh.delete != nil {
		routes = append(routes, router.Route{Handler: mh.delete, Method: http.MethodDelete})
	}
	if mh.options != nil {
		routes = append(routes, router.Route{Handler: mh.options, Method: http.MethodOptions})
	}
	if mh.connect != nil {
		routes = append(routes, router.Route{Handler: mh.connect, Method: http.MethodConnect})
	}
	if mh.propfind != nil {
		routes = append(routes, router.Route{Handler: mh.propfind, Method: PROPFIND})
	}
	if mh.report != nil {
		routes = append(routes, router.Route{Handler: mh.report, Method: REPORT})
	}
	return routes
}

func (mh *methodHandler) Methods() []string {
	routes := mh.Routes()
	ms := make([]string, len(routes))
	for i, r := range routes {
		ms[i] = r.Method
	}
	return ms
}

func (mh *methodHandler) DelHandler(method string) { mh.AddHandler(method, nil) }
func (mh *methodHandler) AddHandler(method string, handler interface{}) {
	switch method {
	case "": // For Any Method
		*mh = methodHandler{
			get:      handler,
			put:      handler,
			post:     handler,
			head:     handler,
			patch:    handler,
			trace:    handler,
			delete:   handler,
			options:  handler,
			connect:  handler,
			propfind: handler,
			report:   handler,
		}
	case http.MethodGet:
		mh.get = handler
	case http.MethodPut:
		mh.put = handler
	case http.MethodPost:
		mh.post = handler
	case http.MethodHead:
		mh.head = handler
	case http.MethodPatch:
		mh.patch = handler
	case http.MethodDelete:
		mh.delete = handler
	case http.MethodOptions:
		mh.options = handler
	case http.MethodConnect:
		mh.connect = handler
	case http.MethodTrace:
		mh.trace = handler
	case PROPFIND:
		mh.propfind = handler
	case REPORT:
		mh.report = handler
	}
}

func (mh *methodHandler) FindHandler(method string) interface{} {
	switch method {
	case http.MethodGet:
		return mh.get
	case http.MethodPut:
		return mh.put
	case http.MethodPost:
		return mh.post
	case http.MethodHead:
		return mh.head
	case http.MethodPatch:
		return mh.patch
	case http.MethodDelete:
		return mh.delete
	case http.MethodOptions:
		return mh.options
	case http.MethodConnect:
		return mh.connect
	case http.MethodTrace:
		return mh.trace
	case PROPFIND:
		return mh.propfind
	case REPORT:
		return mh.report
	default:
		return nil
	}
}

func (mh *methodHandler) HasHandler() bool {
	if mh.get != nil {
		return true
	} else if mh.put != nil {
		return true
	} else if mh.post != nil {
		return true
	} else if mh.head != nil {
		return true
	} else if mh.patch != nil {
		return true
	} else if mh.delete != nil {
		return true
	} else if mh.options != nil {
		return true
	} else if mh.connect != nil {
		return true
	} else if mh.trace != nil {
		return true
	} else if mh.propfind != nil {
		return true
	} else if mh.report != nil {
		return true
	}
	return false
}

/// *********************************************************************** ///

const (
	skind kind = iota
	pkind
	akind
)

type kind uint8

func (k kind) String() string {
	switch k {
	case skind:
		return "static"
	case pkind:
		return "param"
	case akind:
		return "any"
	default:
		panic("unknown node kind")
	}
}

/// *********************************************************************** ///

type node struct {
	name     string   // The name of the node
	ppath    string   // The Pristine full path
	kind     kind     // The kind of the current node
	label    byte     // The first byte of the prefix
	prefix   string   // The same prefix of the paths of all the children
	pnames   []string // The parameters in the registered path
	children []*node
	handlers *methodHandler
	parent   *node
}

func newNode(t kind, name, prefix, ppath string, parent *node, children []*node,
	mh *methodHandler, pnames []string) *node {
	n := &node{
		kind:     t,
		name:     name,
		label:    prefix[0],
		prefix:   prefix,
		parent:   parent,
		children: children,
		ppath:    ppath,
		pnames:   pnames,
		handlers: mh,
	}

	// Fix the parent node of the children nodes.
	for _, c := range n.children {
		c.parent = n
	}
	return n
}

func (n *node) Reset() {
	n.name = ""
	n.ppath = ""
	n.pnames = nil
	*n.handlers = methodHandler{}
}

func (n *node) AddChild(c *node) { n.children = append(n.children, c) }
func (n *node) DelChild(c *node) {
	for i, cn := range n.children {
		if cn == c {
			copy(n.children[i:], n.children[i+1:])
			n.children = n.children[:len(n.children)-1]
			break
		}
	}
}

func (n *node) FindChildByLabel(label byte) *node {
	for _, c := range n.children {
		if c.label == label {
			return c
		}
	}
	return nil
}

func (n *node) FindChild(label byte, t kind) *node {
	for _, c := range n.children {
		if c.label == label && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) FindChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) CheckMethodNotAllowed(r *Router) interface{} {
	if r.conf.MethodNotAllowedHandler == nil || !n.handlers.HasHandler() {
		return r.conf.NotFoundHandler
	} else if f, ok := r.conf.MethodNotAllowedHandler.(func([]string) interface{}); ok {
		return f(n.handlers.Methods())
	}
	return r.conf.MethodNotAllowedHandler
}

/// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

var errInconsistentRouteName = fmt.Errorf("inconsistent route name")
var _ router.Router = &Router{}

// Config is used to configure the router.
type Config struct {
	// NotFoundHandler is returned when not finding the route handler.
	//
	// Default: nil
	NotFoundHandler interface{}

	// MethodNotAllowedHandler is returned when not finding the route handler
	// with the given method and path, which will be called with the allowed
	// methods if it is the function func(allowedMethods []string) interface{}.
	//
	// Default: nil
	MethodNotAllowedHandler interface{}

	// If true, the trailing slash will be removed before adding, deleting
	// and finding the route.
	//
	// Default: false.
	RemoveTrailingSlash bool
}

// Router is the registry of all registered routes to match the request
// with the url method and path, which supports one or more path parameters.
//
// For the single parameter, it starts with ":" followed by the parameter name,
// such as "/v1/:first/path/:second/to/:third".
//
// For the wildcard parameter, its starts with "*" followed by the optional
// parameter name. If no the parameter name, it is "*" by default. such as
// "/v1/path/to/*" or "/v1/path/to/*wildcard".
//
// Moreover, the single and wildcard parameters may used in combination.
// But the wildcard parameter must be the last.
//
// Supported methods:
//   - GET
//   - PUT
//   - HEAD
//   - POST
//   - PATCH
//   - TRACE
//   - DELETE
//   - CONNECT
//   - OPTIONS
//   - PROPFIND
//   - REPORT
//
type Router struct {
	conf    Config
	tree    *node
	bufpool sync.Pool
	maxnum  int               // The maximum number of the parameter
	routes  map[string]string // Name -> Path
}

// NewRouter returns a new Router instance with the config.
//
// If c is nil, use the default configuration.
func NewRouter(c *Config) *Router {
	var conf Config
	if c != nil {
		conf = *c
	}

	return &Router{
		conf:   conf,
		tree:   &node{handlers: new(methodHandler)},
		routes: make(map[string]string, 32),
		bufpool: sync.Pool{New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 64))
		}},
	}
}

// Path returns a url path by the path name and the parameters.
func (r *Router) Path(name string, params ...interface{}) (url string) {
	path := r.routes[name]
	if path == "" {
		return ""
	}

	n := 0
	ln := len(params)
	buf := r.bufpool.Get().(*bytes.Buffer)
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' && n < ln {
			for ; i < l && path[i] != '/'; i++ {
			}

			switch v := params[n].(type) {
			case error:
				buf.WriteString(v.Error())
			case fmt.Stringer:
				buf.WriteString(v.String())
			case io.WriterTo:
				v.WriteTo(buf)
			case float32:
				buf.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
			case float64:
				buf.WriteString(strconv.FormatFloat(v, 'f', -1, 32))
			default:
				fmt.Fprintf(buf, "%v", v)
			}
			n++
		}
		if i < l {
			buf.WriteByte(path[i])
		}
	}

	url = buf.String()
	buf.Reset()
	r.bufpool.Put(buf)
	return
}

// Routes returns the list of the routes, which are filtered by the filter
// if it returns true.
func (r *Router) Routes(filter func(name, path, method string) bool) []router.Route {
	routes := make([]router.Route, 0, 32)
	return r.getRoutes(r.tree, routes, filter)
}

func (r *Router) getRoutes(n *node, routes []router.Route,
	filter func(string, string, string) bool) []router.Route {
	if n.ppath != "" {
		for _, route := range n.handlers.Routes() {
			if filter == nil || filter(n.name, n.ppath, route.Method) {
				routes = append(routes, router.Route{
					Name:    n.name,
					Path:    n.ppath,
					Method:  route.Method,
					Handler: route.Handler,
				})
			}
		}
	}

	for _, cn := range n.children {
		routes = r.getRoutes(cn, routes, filter)
	}
	return routes
}

/// ----------------------------------------------------------------------- ///

// Add registers a new route for method and path with matching handler.
//
// If method is empty, it'll override the handlers of all the supported methods
// with h.
func (r *Router) Add(name, path, method string, h interface{}) (n int, err error) {
	if h == nil {
		return 0, fmt.Errorf("route handler must not be nil")
	}

	// Validate path
	if r.conf.RemoveTrailingSlash {
		path = strings.TrimRight(path, "/")
	}
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}

	var addRoute bool
	if name != "" {
		if orig, ok := r.routes[name]; ok {
			if orig != path {
				return 0, errInconsistentRouteName
			}
		} else {
			addRoute = true
		}

		for n, p := range r.routes {
			if p == path {
				if n != name {
					return 0, errInconsistentRouteName
				}
				break
			}
		}
	}

	notend := true
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert("", method, "", path[:i], skind, nil, nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(name, method, ppath, path[:i], pkind, h, pnames)
				notend = false
				break
			} else {
				r.insert("", method, "", path[:i], pkind, nil, nil)
			}
		} else if path[i] == '*' {
			r.insert("", method, "", path[:i], skind, nil, nil)
			name := strings.TrimRight(path[i+1:], "/ ")
			if name == "" {
				name = "*"
			}
			pnames = append(pnames, name)
			r.insert(name, method, ppath, path[:i+1], akind, h, pnames)
			notend = false
			break
		}
	}

	if notend {
		r.insert(name, method, ppath, path, skind, h, pnames)
	}

	if addRoute {
		r.routes[name] = ppath
	}
	return r.maxnum, nil
}

func (r *Router) insert(name, method, ppath, prefix string, t kind, h interface{}, pnames []string) {
	// Adjust max param
	l := len(pnames)
	if r.maxnum < l {
		r.maxnum = l
	}

	cn := r.tree
	search := prefix

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP: Longest Common Prefix
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 { // No Common Prefix, only for the first route.

			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.name = name
				cn.ppath = ppath
				cn.pnames = pnames
				cn.handlers.AddHandler(method, h)
			}

		} else if l < pl { // The inserted path is the full LCP of the current node.

			// Split node
			n := newNode(cn.kind, cn.name, cn.prefix[l:], cn.ppath, cn, cn.children,
				cn.handlers, cn.pnames)

			// Reset parent node
			cn.name = ""
			cn.kind = skind
			cn.ppath = ""
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.pnames = nil
			cn.children = []*node{n}
			cn.handlers = newMethodHandler()

			if l == sl {
				// At parent node, that's, the inserted path is the new parent node.
				cn.kind = t
				cn.name = name
				cn.ppath = ppath
				cn.pnames = pnames
				cn.handlers.AddHandler(method, h)
			} else {
				// Create child node, that's, the path of the new parent node
				// is the full LCP of the inserted path.
				n = newNode(t, name, search[l:], ppath, cn, nil, newMethodHandler(), pnames)
				n.handlers.AddHandler(method, h)
				cn.AddChild(n)
			}

		} else if l < sl { // The path of the current node is the full LCP of the inserted path.

			search = search[l:]
			c := cn.FindChildByLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}

			// Create child node
			n := newNode(t, name, search, ppath, cn, nil, newMethodHandler(), pnames)
			n.handlers.AddHandler(method, h)
			cn.AddChild(n)

		} else {

			// Node already exists, that's, the insert path is the current node.
			// We override it with the new inserted route, but the parameters.
			if h != nil {
				cn.ppath = ppath
				cn.handlers.AddHandler(method, h)
				if len(cn.pnames) == 0 { // Issue #729
					cn.pnames = pnames
				}
				if cn.name == "" {
					cn.name = name
				}
			}

		}

		return
	}
}

// Match lookups a handler registered for method and path,
// which also parses the path for the parameters.
func (r *Router) Match(path, method string, pnames, pvalues []string) (h interface{}, pn int) {
	if r.conf.RemoveTrailingSlash {
		// path = strings.TrimRight(path, "/")
		path = removeTrailingSlash(path)
	}
	if path == "" {
		path = "/"
	}

	var (
		hasp   = len(pnames) > 0 && len(pvalues) > 0
		cn     = r.tree
		search = path
		child  *node  // Child node
		n      int    // Param counter
		nk     kind   // Next kind
		nn     *node  // Next node
		ns     string // Next search
	)

	// Search order static > param > any
	for {
		if search == "" {
			break
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl := len(search); sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			if nn == nil { // Issue #1348
				return r.conf.NotFoundHandler, 0 // Not found
			}

			cn = nn
			search = ns
			switch nk {
			case pkind:
				goto Param
			case akind:
				goto Any
			}
		}

		if search == "" {
			break
		}

		// Search Static Node
		if child = cn.FindChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nn = cn // For backtrack
				nk = pkind
				ns = search
			}
			cn = child
			continue
		}

		// Search Param Node
	Param:
		if child = cn.FindChildByKind(pkind); child != nil {
			if hasp && len(pvalues) == n { // Issue #378
				continue
			}

			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = akind
				nn = cn
				ns = search
			}
			cn = child

			var i int
			for l := len(search); i < l && search[i] != '/'; i++ {
			}
			if hasp {
				pvalues[n] = search[:i]
			}
			n++

			search = search[i:]
			continue
		}

		// Search Any Node
	Any:
		if cn = cn.FindChildByKind(akind); cn == nil {
			if nn != nil {
				cn = nn
				nn = cn.parent // Next (Issue #954)
				if nn != nil {
					nk = nn.kind
				}

				search = ns
				switch nk {
				case pkind:
					goto Param
				case akind:
					goto Any
				}
			}

			return r.conf.NotFoundHandler, 0 // Not found
		}

		if hasp {
			pvalues[len(cn.pnames)-1] = search
		}
		break
	}

	if h = cn.handlers.FindHandler(method); h == nil { // NOTE: Slow zone...
		// Dig further for any, might have an empty value for *,
		// e.g. serving a directory. Issue #207.
		if n := cn.FindChildByKind(akind); n == nil {
			h = cn.CheckMethodNotAllowed(r)
		} else {
			if h = n.handlers.FindHandler(method); h == nil {
				h = n.CheckMethodNotAllowed(r)
			}

			if pn = len(n.pnames); pn > 0 && hasp {
				copy(pnames, n.pnames)
				pvalues[pn-1] = ""
			}
		}
	} else if pn = len(cn.pnames); pn > 0 && hasp {
		copy(pnames, cn.pnames)
	}

	return
}

/// ----------------------------------------------------------------------- ///

// Del deletes the given route.
func (r *Router) Del(path, method string) (err error) {
	if path != "" {
		err = r.delRoute(path, method)
	}
	return
}

func (r *Router) delRoute(path, method string) (err error) {
	if r.conf.RemoveTrailingSlash {
		// path = strings.TrimRight(path, "/")
		path = removeTrailingSlash(path)
	}
	if path == "" {
		path = "/"
	}

	var (
		cn     = r.tree
		search = path
		child  *node  // Child node
		nk     kind   // Next kind
		nn     *node  // Next node
		ns     string // Next search
	)

	// Search order static > param > any
	for {
		if search == "" {
			break
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl := len(search); sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			if nn == nil { // Issue #1348
				return // Not found
			}

			cn = nn
			search = ns
			switch nk {
			case pkind:
				goto Param
			case akind:
				goto Any
			}
		}

		if search == "" {
			break
		}

		// Search Static Node
		if child = cn.FindChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nn = cn // For backtrack
				nk = pkind
				ns = search
			}
			cn = child
			continue
		}

		// Search Param Node
	Param:
		if child = cn.FindChildByKind(pkind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nn = cn // For backtrack
				nk = akind
				ns = search
			}
			cn = child

			var i int
			for _len := len(search); i < _len && search[i] != '/'; i++ {
			}

			search = search[i:]
			continue
		}

		// Search Any Node
	Any:
		if cn = cn.FindChildByKind(akind); cn == nil {
			if nn != nil {
				cn = nn
				nn = cn.parent // Next (Issue #954)
				if nn != nil {
					nk = nn.kind
				}

				search = ns
				switch nk {
				case pkind:
					goto Param
				case akind:
					goto Any
				}
			}

			return // Not found
		}

		break
	}

	// Delete the found node.
	r.removeNode(cn, method)
	return
}

func (r *Router) removeNode(cn *node, method string) {
	if cn == nil {
		return
	}

	if method != "" {
		cn.handlers.DelHandler(method)
		if cn.handlers.HasHandler() {
			return
		}
	}

	r.removeWholeNode(cn)
}

func (r *Router) replaceParentWithChild(cn *node) {
	child := cn.children[0]
	if cn.parent == nil {
		// The removed node is root and use the child to replace it.
		prefix := cn.prefix
		r.tree = child
		r.tree.parent = nil
		r.tree.prefix = prefix + r.tree.prefix
		r.tree.label = r.tree.prefix[0]
		r.removeRouteNameByPath(cn.ppath)
	} else {
		parent := cn.parent
		child.parent = parent
		child.prefix = cn.prefix + child.prefix
		child.label = child.prefix[0]
		parent.DelChild(cn)
		parent.AddChild(child)
		r.removeRouteNameByPath(cn.ppath)
	}
}

func (r *Router) removeRouteNameByPath(path string) {
	if path != "" {
		for n, p := range r.routes {
			if p == path {
				delete(r.routes, n)
				break
			}
		}
	}
}

func (r *Router) removeWholeNode(cn *node) {
	if cn == nil {
		return
	}

	switch cn {
	case nil:
		return
	case r.tree:
		r.removeRouteNameByPath(cn.ppath)
		if len(cn.children) == 0 {
			cn.prefix = ""
			cn.label = 0
		}

		cn.Reset()
		return
	}

	switch len(cn.children) {
	case 0: // Leaf node, and remove it from the parent node.
		parent := cn.parent
		parent.DelChild(cn)
		r.removeRouteNameByPath(cn.ppath)

		if _len := len(parent.children); _len == 0 {
			// Remove the useless intermediate node recursively.
			if parent.ppath == "" {
				r.removeWholeNode(parent)
			}
		} else if _len == 1 {
			// The parent node is useless intermediate node,
			// and only contains one the leaf node, so merge them.
			if parent.kind == skind && parent.ppath == "" {
				r.replaceParentWithChild(parent)
			}
		}
	case 1: // Not leaf node, but only contain the one child node.
		// The child node is not the static node.
		// So clean instead of removing it.
		switch cn.kind {
		case skind:
			// Static node, remove the current node and use the child node
			// instead of it.
			r.replaceParentWithChild(cn)
		default:
			r.removeRouteNameByPath(cn.ppath)
			// Param or Any node, only clean it.
			cn.Reset()
		}
	default:
		r.removeRouteNameByPath(cn.ppath)

		// The current node contains more than one child node,
		// so we only clean it, not remove it.
		cn.Reset()
	}
}

/// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// PrintTree prints the tree structure of the router.
func (r *Router) PrintTree(w io.Writer) {
	if r.tree.prefix != "" {
		r.tree.printTree(w, "", true, true)
	}
}

func (n *node) printTree(w io.Writer, pfx string, first, tail bool) {
	prefix := getPrefix(first, tail, pfx, "└── ", "├── ")
	w.Write([]byte(fmt.Sprintf("%s%s @%p, parent=%p, type=%s, name=%s, path=%s, pnames=%v, methods=%v\n",
		prefix, n.prefix, n, n.parent, n.kind, n.name, n.ppath, n.pnames, n.handlers.Methods())))

	_len := len(n.children)
	prefix = getPrefix(first, tail, pfx, "    ", "│   ")
	for i := 0; i < _len-1; i++ {
		n.children[i].printTree(w, prefix, false, false)
	}
	if _len > 0 {
		n.children[_len-1].printTree(w, prefix, false, true)
	}
}

func getPrefix(first, tail bool, p, on, off string) string {
	if tail {
		if first {
			return ""
		}
		return fmt.Sprintf("%s%s", p, on)
	}
	return fmt.Sprintf("%s%s", p, off)
}
