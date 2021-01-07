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

// Package general supplies a general Router implementation based on Radix Tree.
package general

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/xgfone/ship/v3/router"
)

// MethodHandler is used to manage the mapping from method to handler.
type MethodHandler interface {
	Reset()
	Methods() []string
	Handlers() map[string]interface{}
	AddHandler(method string, handler interface{})
	GetHandler(method string) (handler interface{})
	DelHandler(method string)
}

type methodHandler struct {
	any      interface{}
	handlers map[string]interface{}
}

func newMethodHandler() MethodHandler {
	return &methodHandler{handlers: make(map[string]interface{}, 8)}
}

func (mh *methodHandler) Reset() { *mh = methodHandler{} }

func (mh *methodHandler) Handlers() map[string]interface{} {
	handlers := make(map[string]interface{}, 12)
	for m, h := range mh.handlers {
		handlers[m] = h
	}
	if mh.any != nil {
		handlers["*"] = mh.any
	}
	return handlers
}

func (mh *methodHandler) Methods() []string {
	methods := make([]string, 0, len(mh.handlers)+1)
	for m := range mh.handlers {
		methods = append(methods, m)
	}
	if mh.any != nil {
		methods = append(methods, "*")
	}
	return methods
}

func (mh *methodHandler) DelHandler(m string) {
	if m == "" {
		mh.any = nil
	} else {
		delete(mh.handlers, m)
	}
}

func (mh *methodHandler) AddHandler(m string, h interface{}) {
	if m == "" {
		mh.any = h
	} else {
		mh.handlers[m] = h
	}
}

func (mh *methodHandler) GetHandler(method string) interface{} {
	if h, ok := mh.handlers[method]; ok {
		return h
	}
	return mh.any
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
	handlers MethodHandler
	parent   *node
}

func newNode(t kind, name, prefix, ppath string, parent *node, children []*node,
	mh MethodHandler, pnames []string) *node {
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
	n.handlers.Reset()
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

func (n *node) CheckMethodNotAllowed(r *Router, h interface{}) interface{} {
	if r.notAllowed != nil && len(n.handlers.Methods()) != 0 {
		return r.notAllowed
	}
	return h
}

/// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

var errInconsistentRouteName = fmt.Errorf("inconsistent route name")
var _ router.Router = &Router{}

// Router is the registry of all registered routes for request matching
// and URL path parameter parsing.
type Router struct {
	tree       *node
	bufpool    sync.Pool
	maxnum     int               // The maximum number of the parameter
	routes     map[string]string // Name -> Path
	notFound   interface{}       // The NotFound handler
	notAllowed interface{}       // The MethodNotAllowed handler
	mhFunc     func() MethodHandler
}

// NewRouter returns a new Router instance.
func NewRouter(notFoundHandler, methodNotAllowedHandler interface{},
	newMethodHandlerFunc ...func() MethodHandler) *Router {
	newMethodHandler := newMethodHandler
	if len(newMethodHandlerFunc) > 0 && newMethodHandlerFunc[0] != nil {
		newMethodHandler = newMethodHandlerFunc[0]
	}

	return &Router{
		mhFunc:     newMethodHandler,
		tree:       &node{handlers: newMethodHandler()},
		routes:     make(map[string]string, 32),
		notFound:   notFoundHandler,
		notAllowed: methodNotAllowedHandler,
		bufpool: sync.Pool{New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 64))
		}},
	}
}

// URL returns a url by the name and the params.
func (r *Router) URL(name string, params ...interface{}) (url string) {
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

// Routes returns the list of all the routes.
func (r *Router) Routes() []router.Route {
	routes := make([]router.Route, 0, 64)
	return r.getRoutes(r.tree, routes)
}

func (r *Router) getRoutes(n *node, routes []router.Route) []router.Route {
	if n.ppath != "" {
		for method, handler := range n.handlers.Handlers() {
			routes = append(routes, router.Route{
				Name:    n.name,
				Path:    n.ppath,
				Method:  method,
				Handler: handler,
			})
		}
	}

	for _, cn := range n.children {
		routes = r.getRoutes(cn, routes)
	}
	return routes
}

/// ----------------------------------------------------------------------- ///

// Add registers a new route for method and path with matching handler.
func (r *Router) Add(name, method, path string, h interface{}) (n int, err error) {
	if h == nil {
		return 0, fmt.Errorf("route handler must not be nil")
	}

	// Validate path
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
			cn.handlers = r.mhFunc()

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
				n = newNode(t, name, search[l:], ppath, cn, nil, r.mhFunc(), pnames)
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
			n := newNode(t, name, search, ppath, cn, nil, r.mhFunc(), pnames)
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

// Find lookups a handler registered for method and path,
// which also parses the path for the parameters.
func (r *Router) Find(method, path string, pnames, pvalues []string) (h interface{}, pn int) {
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
				return r.notFound, 0 // Not found
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

			return r.notFound, 0 // Not found
		}

		if hasp {
			pvalues[len(cn.pnames)-1] = search
		}
		break
	}

	if h = cn.handlers.GetHandler(method); h == nil { // NOTE: Slow zone...
		// Dig further for any, might have an empty value for *,
		// e.g. serving a directory. Issue #207.
		if n := cn.FindChildByKind(akind); n == nil {
			h = cn.CheckMethodNotAllowed(r, r.notFound)
		} else {
			if h = n.handlers.GetHandler(method); h == nil {
				h = n.CheckMethodNotAllowed(r, r.notFound)
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
func (r *Router) Del(name, method, path string) (err error) {
	if name != "" {
		path = r.routes[name]
	}
	if path != "" {
		err = r.delRoute(path, method)
	}
	return
}

func (r *Router) delRoute(path, method string) (err error) {
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
		if len(cn.handlers.Methods()) != 0 {
			return
		}
	}

	r.removeWholeNode(cn)
	return
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
