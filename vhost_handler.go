// Copyright 2021 xgfone
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

package ship

import (
	"net/http"
	"sync/atomic"
)

// HostHandler is a http handler to dispatch the request to the host handler
// by the request host, which may be used to implement the virtual hosts.
type HostHandler interface {
	http.Handler
	HostManager
}

type defaultHostHandler struct {
	Handler http.Handler
	Host    string
}

// HostManagerHandler is an implementation of HostHandler.
type HostManagerHandler struct {
	HostManager

	// HandleHTTP is used to handle the matched host and handler.
	//
	// If not found the matched host and handler, matchedHost and matchedHandler
	// are ZERO, that's, "" and nil.
	//
	// Default: w.WriteHeader(404)
	HandleHTTP func(w http.ResponseWriter, r *http.Request,
		matchedHost string, matchedHandler http.Handler)

	defaultHost atomic.Value
}

// NewHostManagerHandler returns a new HostManagerHandler.
//
// If hostManager is nil, it is NewHostManager(nil) by default.
func NewHostManagerHandler(hostManager HostManager) *HostManagerHandler {
	if hostManager == nil {
		hostManager = NewHostManager(nil)
	}
	return &HostManagerHandler{HostManager: hostManager}
}

// GetDefaultHost returns the default host and handler.
//
// Return ("", nil) if the default host is not set.
func (h *HostManagerHandler) GetDefaultHost() (host string, handler http.Handler) {
	if v := h.defaultHost.Load(); v != nil {
		d := v.(defaultHostHandler)
		host, handler = d.Host, d.Handler
	}
	return
}

// SetDefaultHost sets the default host and handler.
func (h *HostManagerHandler) SetDefaultHost(host string, handler http.Handler) {
	h.defaultHost.Store(defaultHostHandler{Host: host, Handler: handler})
}

// ServeHTTP implements the interface http.Handler.
func (h *HostManagerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var matchedHost string
	var matchedHandler http.Handler
	if r.Host != "" && h.Len() > 0 {
		matchedHost, matchedHandler = h.MatchHost(r.Host)
	}

	if matchedHandler == nil {
		matchedHost, matchedHandler = h.GetDefaultHost()
	}

	if h.HandleHTTP == nil {
		h.handleHTTP(w, r, matchedHost, matchedHandler)
	} else {
		h.HandleHTTP(w, r, matchedHost, matchedHandler)
	}
}

func (h *HostManagerHandler) handleHTTP(w http.ResponseWriter, r *http.Request,
	matchedHost string, matchedHandler http.Handler) {
	if matchedHandler == nil {
		w.WriteHeader(404)
	} else {
		matchedHandler.ServeHTTP(w, r)
	}
}
