// The MIT License (MIT)
//
// Copyright (c) 2019 xgfone
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

// Package echo supplies a Router implementation based on github.com/labstack/echo.
package echo

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// PROPFIND Method can be used on collection and property resources.
const PROPFIND = "PROPFIND"

// REPORT Method can be used to get information about a resource, see rfc 3253
const REPORT = "REPORT"

var methods = [...]string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	PROPFIND,
	http.MethodPut,
	http.MethodTrace,
	REPORT,
}

var bufPool = sync.Pool{New: func() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 64))
}}

type (
	// Router is the registry of all registered routes for request matching
	// and URL path parameter parsing.
	Router struct {
		tree   *node
		pnum   int
		routes map[string]string

		methodNotAllowed interface{}
	}
	node struct {
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
		connect  interface{}
		delete   interface{}
		get      interface{}
		head     interface{}
		options  interface{}
		patch    interface{}
		post     interface{}
		propfind interface{}
		put      interface{}
		trace    interface{}
		report   interface{}
	}
)

const (
	skind kind = iota
	pkind
	akind
)

// NewRouter returns a new Router instance.
func NewRouter(methodNotAllowedHandler interface{}) *Router {
	return &Router{
		tree:   &node{methodHandler: new(methodHandler)},
		routes: make(map[string]string, 32),

		methodNotAllowed: methodNotAllowedHandler,
	}
}

// URL returns a url by the name and the params.
func (r *Router) URL(name string, params ...interface{}) string {
	path := r.routes[name]
	if path == "" {
		return ""
	}

	buf := bufPool.Get().(*bytes.Buffer)
	ln := len(params)
	n := 0
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

	uri := buf.String()
	buf.Reset()
	bufPool.Put(buf)
	return uri
}

// Add registers a new route for method and path with matching handler.
func (r *Router) Add(name, method, path string, h interface{}) (paramNum int) {
	// Validate path
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], h, pkind, ppath, pnames)
			} else {
				r.insert(method, path[:i], nil, pkind, "", nil)
			}
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			name := strings.TrimRight(path[i+1:], "/ ")
			if name == "" {
				name = "*"
			}
			pnames = append(pnames, name)
			r.insert(method, path[:i+1], h, akind, ppath, pnames)
		}
	}

	r.insert(method, path, h, skind, ppath, pnames)
	if name != "" {
		r.routes[name] = ppath
	}
	return r.pnum
}

func (r *Router) insert(method, path string, h interface{}, t kind,
	ppath string, pnames []string) {
	// Adjust max param
	l := len(pnames)
	if r.pnum < l {
		r.pnum = l
	}

	cn := r.tree // Current node as root
	if cn == nil {
		panic("echo: invalid method")
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
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children,
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
				n = newNode(t, search[l:], cn, nil, new(methodHandler), ppath, pnames)
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
			n := newNode(t, search, cn, nil, new(methodHandler), ppath, pnames)
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

func newNode(t kind, pre string, p *node, c children, mh *methodHandler,
	ppath string, pnames []string) *node {
	return &node{
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

func (n *node) addHandler(method string, h interface{}) {
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
	case PROPFIND:
		n.methodHandler.propfind = h
	case http.MethodPut:
		n.methodHandler.put = h
	case http.MethodTrace:
		n.methodHandler.trace = h
	case REPORT:
		n.methodHandler.report = h
	}
}

func (n *node) findHandler(method string) interface{} {
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
	case PROPFIND:
		return n.methodHandler.propfind
	case http.MethodPut:
		return n.methodHandler.put
	case http.MethodTrace:
		return n.methodHandler.trace
	case REPORT:
		return n.methodHandler.report
	default:
		return nil
	}
}

func (n *node) checkMethodNotAllowed(r *Router, h interface{}) interface{} {
	if r.methodNotAllowed != nil {
		for _, m := range methods {
			if n.findHandler(m) != nil {
				return r.methodNotAllowed
			}
		}
	}
	return h
}

// Find lookup a handler registered for method and path. It also parses URL
// for path parameters and load them into context.
func (r *Router) Find(method, path string, pnames, pvalues []string,
	defaultHandler interface{}) (handler interface{}) {
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
			if nn == nil { // Issue #1348
				return defaultHandler // Not found
			}
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
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
				if nn != nil {
					nk = nn.kind
				}
				search = ns
				if nk == pkind {
					goto Param
				} else if nk == akind {
					goto Any
				}
			}
			return defaultHandler // Not found
		}
		pvalues[len(cn.pnames)-1] = search
		break
	}

	if handler = cn.findHandler(method); handler == nil { // NOTE: Slow zone...
		handler = cn.checkMethodNotAllowed(r, defaultHandler)

		// Dig further for any, might have an empty value for *, e.g.
		// serving a directory. Issue #207.
		if cn = cn.findChildByKind(akind); cn == nil {
			return
		} else if handler = cn.findHandler(method); handler == nil {
			handler = cn.checkMethodNotAllowed(r, defaultHandler)
		}

		if len(cn.pnames) > 0 {
			copy(pnames, cn.pnames)
			pvalues[len(cn.pnames)-1] = ""
		}
	} else if len(cn.pnames) > 0 {
		copy(pnames, cn.pnames)
	}

	return
}

//////////////////////////////////////////////////////////////////////////////

var kindtypes = map[kind]string{skind: "static", pkind: "param", akind: "any"}

// PrintRouterTree prints the tree structure of the router.
func (r *Router) PrintRouterTree(w io.Writer) {
	r.tree.printTree(w, "", true)
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
