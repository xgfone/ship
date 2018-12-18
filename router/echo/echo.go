// The MIT License (MIT)
//
// Copyright (c) 2018 xgfone <xgfone@126.com>
// Copyright (c) 2017 LabStack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package echo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xgfone/ship/core"
)

// PROPFIND stands for a PROPFIND HTTP method.
var PROPFIND = "PROPFIND"

type (
	// Route contains a handler and information for matching against requests.
	route struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Name   string `json:"name"`
	}
	// router is the registry of all registered routes for an `Echo` instance
	// for request matching and URL path parameter parsing.
	router struct {
		tree       *node
		routes     map[string]*route
		allroutes  []*route
		options    core.Handler
		notAllowed core.Handler
	}
	node struct {
		router        *router
		kind          kind
		label         byte
		prefix        string
		parent        *node
		children      children
		ppath         string
		pnames        []string
		methodHandler *methodHandler
	}
	kind          uint8
	children      []*node
	methodHandler struct {
		connect  core.Handler
		delete   core.Handler
		get      core.Handler
		head     core.Handler
		options  core.Handler
		patch    core.Handler
		post     core.Handler
		put      core.Handler
		trace    core.Handler
		propfind core.Handler
	}
)

const (
	skind kind = iota
	pkind
	akind
)

var methods = [...]string{
	http.MethodConnect,
	http.MethodOptions,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodTrace,
	http.MethodHead,
	http.MethodPost,
	http.MethodGet,
	http.MethodPut,
	PROPFIND,
}

// NewRouter returns a new Router instance.
func NewRouter(methodNotAllowedHandler, optionsHandler core.Handler) core.Router {
	r := &router{
		tree:       &node{methodHandler: new(methodHandler)},
		routes:     map[string]*route{},
		allroutes:  []*route{},
		options:    optionsHandler,
		notAllowed: methodNotAllowedHandler,
	}
	r.tree.router = r
	return r
}

// Each implements github.com/xgfone/ship:Router#Each.
func (r *router) Each(f func(name, method, path string)) {
	for _, route := range r.allroutes {
		f(route.Name, route.Method, route.Path)
	}
}

// URL implements github.com/xgfone/ship:Router#URL.
func (r *router) URL(name string, params ...interface{}) string {
	route := r.routes[name]
	if route == nil {
		return ""
	}

	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for i, l := 0, len(route.Path); i < l; i++ {
		if route.Path[i] == ':' && n < ln {
			for ; i < l && route.Path[i] != '/'; i++ {
			}
			uri.WriteString(fmt.Sprintf("%v", params[n]))
			n++
		}
		if i < l {
			uri.WriteByte(route.Path[i])
		}
	}

	return uri.String()
}

// Add implements github.com/xgfone/ship:Router#Add, which will register
// a new route for method and path with matching handler.
func (r *router) Add(name, path string, method string, handler core.Handler) int {
	for _, _r := range r.allroutes {
		if _r.Method == method && _r.Path == path {
			panic(fmt.Errorf("the route('%s', '%s') has been registered", method, path))
		}
	}

	_route := &route{Name: name, Method: method, Path: path}
	if len(name) > 0 {
		if _r, ok := r.routes[name]; ok && _r.Path != path {
			panic(fmt.Errorf("the url name '%s' has been registered for the path '%s'",
				name, _r.Path))
		}
		r.routes[name] = _route
	}
	r.allroutes = append(r.allroutes, _route)

	return r.add(path, method, handler)
}

func (r *router) add(path string, method string, h core.Handler) int {
	// Validate path
	if path == "" {
		panic(errors.New("echo: path cannot be empty"))
	}
	if path[0] != '/' {
		path = "/" + path
	}

	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			if path[i-1] != '/' {
				panic(errors.New("':' msut be following with '/'"))
			}

			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			if i == j {
				panic(errors.New("':' is not followed by any argument"))
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames)
				return len(pnames)
			}
			r.insert(method, path[:i], nil, pkind, "", nil)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)

			name := strings.TrimRight(path[i+1:], "/ ")
			if name == "" {
				name = "*"
			}
			pnames = append(pnames, name)

			r.insert(method, path[:i+1], h, akind, ppath, pnames)
			return len(pnames)
		}
	}

	r.insert(method, path, h, skind, ppath, pnames)
	return len(pnames)
}

func (r *router) insert(method, path string, h core.Handler, t kind,
	ppath string, pnames []string) {

	cn := r.tree // Current node as root
	if cn == nil {
		panic(errors.New("echo: invalid method"))
	}
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if h != nil {
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			}
		} else if l < pl {
			// Split node
			n := newNode(r, cn.kind, cn.prefix[l:], cn, cn.children,
				cn.methodHandler, cn.ppath, cn.pnames)

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodHandler = new(methodHandler)
			cn.ppath = ""
			cn.pnames = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addHandler(method, h)
				cn.ppath = ppath
				cn.pnames = pnames
			} else {
				// Create child node
				n = newNode(r, t, search[l:], cn, nil, new(methodHandler), ppath, pnames)
				n.addHandler(method, h)
				cn.addChild(n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(r, t, search, cn, nil, new(methodHandler), ppath, pnames)
			n.addHandler(method, h)
			cn.addChild(n)
		} else {
			// Node already exists
			if h != nil {
				cn.addHandler(method, h)
				cn.ppath = ppath
				if len(cn.pnames) == 0 { // Issue #729
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

func newNode(r *router, t kind, pre string, p *node, c children,
	mh *methodHandler, ppath string, pnames []string) *node {
	return &node{
		router:        r,
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		ppath:         ppath,
		pnames:        pnames,
		methodHandler: mh,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) addHandler(method string, h core.Handler) {
	switch method {
	case http.MethodConnect:
		n.methodHandler.connect = h
	case http.MethodDelete:
		n.methodHandler.delete = h
	case http.MethodGet:
		n.methodHandler.get = h
	case http.MethodHead:
		n.methodHandler.head = h
	case http.MethodOptions:
		n.methodHandler.options = h
	case http.MethodPatch:
		n.methodHandler.patch = h
	case http.MethodPost:
		n.methodHandler.post = h
	case http.MethodPut:
		n.methodHandler.put = h
	case http.MethodTrace:
		n.methodHandler.trace = h
	case "PROPFIND":
		n.methodHandler.propfind = h
	default:
		panic(errors.New("not support the method +'" + method + "'"))
	}
}

func (n *node) findHandler(method string) core.Handler {
	switch method {
	case http.MethodConnect:
		return n.methodHandler.connect
	case http.MethodDelete:
		return n.methodHandler.delete
	case http.MethodGet:
		return n.methodHandler.get
	case http.MethodHead:
		return n.methodHandler.head
	case http.MethodOptions:
		return n.methodHandler.options
	case http.MethodPatch:
		return n.methodHandler.patch
	case http.MethodPost:
		return n.methodHandler.post
	case http.MethodPut:
		return n.methodHandler.put
	case http.MethodTrace:
		return n.methodHandler.trace
	case "PROPFIND":
		return n.methodHandler.propfind
	default:
		return nil
	}
}

func (n *node) checkMethodNotAllowed(method string) core.Handler {
	if n.router.notAllowed == nil || method == http.MethodConnect {
		return nil
	}

	ms := make([]string, 0, len(methods))
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			ms = append(ms, m)
		}
	}

	if len(ms) == 0 {
		return nil
	}

	return func(ctx core.Context) error {
		ctx.Response().Header().Set("Allow", strings.Join(ms, ", "))
		return n.router.notAllowed(ctx)
	}
}

func (n *node) checkOptions(method string) core.Handler {
	if n.router.options == nil || method != http.MethodOptions {
		return nil
	}

	ms := make([]string, 0, len(methods))
	h := n.methodHandler
	if h.connect != nil {
		ms = append(ms, http.MethodConnect)
	}
	if h.delete != nil {
		ms = append(ms, http.MethodDelete)
	}
	if h.get != nil {
		ms = append(ms, http.MethodGet)
	}
	if h.head != nil {
		ms = append(ms, http.MethodHead)
	}
	if h.patch != nil {
		ms = append(ms, http.MethodPatch)
	}
	if h.post != nil {
		ms = append(ms, http.MethodPost)
	}
	if h.put != nil {
		ms = append(ms, http.MethodPut)
	}
	if h.trace != nil {
		ms = append(ms, http.MethodTrace)
	}
	if h.propfind != nil {
		ms = append(ms, PROPFIND)
	}

	if len(ms) == 0 {
		return nil
	}

	return func(ctx core.Context) error {
		ctx.Response().Header().Set("Allow", strings.Join(ms, ", "))
		return n.router.options(ctx)
	}
}

// Find implements github.com/xgfone/ship:Router#Find.
func (r *router) Find(method, path string, pnames, pvalues []string) (handler core.Handler) {
	cn := r.tree // Current node as root

	var (
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
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
			// Not found
			return
		}

		if search == "" {
			break
		}

		// Static node
		if child = cn.findChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = pkind
				nn = cn
				ns = search
			}
			cn = child
			continue
		}

		// Param node
	Param:
		if child = cn.findChildByKind(pkind); child != nil {
			// Issue #378
			if len(pvalues) == n {
				continue
			}

			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' { // Issue #623
				nk = akind
				nn = cn
				ns = search
			}

			cn = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Any node
	Any:
		if cn = cn.findChildByKind(akind); cn == nil {
			if nn != nil {
				cn = nn
				nn = cn.parent // Next (Issue #954)
				search = ns
				if nk == pkind {
					goto Param
				} else if nk == akind {
					goto Any
				}
			}
			// Not found
			return
		}
		pvalues[len(cn.pnames)-1] = search
		break
	}

	handler = cn.findHandler(method)
	copy(pnames, cn.pnames)

	// NOTE: Slow zone...
	if handler == nil {
		if handler = cn.checkOptions(method); handler != nil {
			return
		}
		_cn := cn

		// Dig further for any, might have an empty value for *, e.g.
		// serving a directory. Issue #207.
		if cn = cn.findChildByKind(akind); cn == nil {
			handler = _cn.checkMethodNotAllowed(method)
			return
		}
		if handler = cn.findHandler(method); handler == nil {
			handler = cn.checkMethodNotAllowed(method)
		}
		copy(pnames, cn.pnames)
		pvalues[len(cn.pnames)-1] = ""
	}

	return
}

//////////////////////////////////////////////////////////////////////////////

var kindtypes = map[kind]string{skind: "static", pkind: "param", akind: "any"}

// PrintRouterTree prints the tree structure of the router.
func PrintRouterTree(w io.Writer, r core.Router) {
	if _r, ok := r.(*router); ok {
		_r.tree.printTree(w, "", true)
		return
	}
	panic(errors.New("the router is not a ECHO implementation"))
}

func (n *node) printTree(w io.Writer, pfx string, tail bool) {
	p := prefix(tail, pfx, "└── ", "├── ")
	w.Write([]byte(fmt.Sprintf("%s%s, %p: type=%s, lable=%c, path=%s, parent=%p, pnames=%v\n",
		p, n.prefix, n, kindtypes[n.kind], n.label, n.ppath, n.parent, n.pnames)))

	children := n.children
	l := len(children)
	p = prefix(tail, pfx, "    ", "│   ")
	for i := 0; i < l-1; i++ {
		children[i].printTree(w, p, false)
	}
	if l > 0 {
		children[l-1].printTree(w, p, true)
	}
}

func prefix(tail bool, p, on, off string) string {
	if tail {
		return fmt.Sprintf("%s%s", p, on)
	}
	return fmt.Sprintf("%s%s", p, off)
}
