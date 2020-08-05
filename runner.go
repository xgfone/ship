// Copyright 2018 xgfone
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
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// DefaultSignals is a set of default signals.
var DefaultSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGINT,
}

// Runner is a HTTP Server runner.
type Runner struct {
	Name      string
	Logger    Logger
	Server    *http.Server
	Handler   http.Handler
	Signals   []os.Signal
	ConnState func(net.Conn, http.ConnState)

	done   chan struct{}
	shut   *OnceRunner
	stop   *OnceRunner
	stopfs []*OnceRunner
}

// NewRunner returns a new Runner.
func NewRunner(name string, handler http.Handler) *Runner {
	r := &Runner{
		Name:    name,
		Server:  &http.Server{Handler: handler},
		Signals: DefaultSignals,
		Handler: handler, done: make(chan struct{}),
	}

	r.shut = NewOnceRunner(r.runShutdown)
	r.stop = NewOnceRunner(r.runStopfs)
	return r
}

// Link registers the shutdown function between itself and other,
// then returns itself.
func (r *Runner) Link(other *Runner) *Runner {
	other.RegisterOnShutdown(r.Stop)
	return r.RegisterOnShutdown(other.Stop)
}

// RegisterOnShutdown registers some functions to run when the http server is
// shut down.
func (r *Runner) RegisterOnShutdown(functions ...func()) *Runner {
	for _, f := range functions {
		r.stopfs = append(r.stopfs, NewOnceRunner(f))
	}
	return r
}

// Shutdown stops the HTTP server.
func (r *Runner) Shutdown(ctx context.Context) (err error) {
	err = r.Server.Shutdown(ctx)
	r.stop.Run()
	return
}

// Stop is the same as r.Shutdown(context.Background()).
func (r *Runner) Stop()        { r.shut.Run() }
func (r *Runner) runShutdown() { r.Shutdown(context.Background()) }
func (r *Runner) runStopfs() {
	defer close(r.done)
	for i := len(r.stopfs) - 1; i >= 0; i-- {
		r.stopfs[i].Run()
	}
}

// Wait waits until all the registered shutdown functions have finished.
func (r *Runner) Wait() { <-r.done }

// Start starts a HTTP server with addr and ends when the server is closed.
//
// If tlsFiles is not nil, it must be certFile and keyFile. For example,
//    runner := NewRunner()
//    runner.Start(":80", certFile, keyFile)
func (r *Runner) Start(addr string, tlsFiles ...string) *Runner {
	var cert, key string
	if len(tlsFiles) == 2 && tlsFiles[0] != "" && tlsFiles[1] != "" {
		cert = tlsFiles[0]
		key = tlsFiles[1]
	}

	if r.Server == nil {
		r.Server = &http.Server{Addr: addr, Handler: r.Handler}
	}

	if r.Server.Handler == nil {
		r.Server.Handler = r.Handler
	}

	if addr != "" {
		r.Server.Addr = addr
	}

	r.startServer(cert, key)
	return r
}

func (r *Runner) handleSignals() {
	if len(r.Signals) > 0 {
		ss := make(chan os.Signal, 1)
		signal.Notify(ss, r.Signals...)
		for {
			<-ss
			r.Stop()
			return
		}
	}
}

func (r *Runner) startServer(certFile, keyFile string) {
	if r.Server.Addr == "" {
		panic("Runner: Server.Addr is empty")
	} else if r.Server.Handler == nil {
		panic("Runner: Server.Handler is nil")
	}

	defer r.Stop()
	name := r.Name
	server := r.Server
	logger := r.Logger

	if logger != nil {
		if name == "" {
			logger.Infof("The HTTP Server is running on %s", server.Addr)
		} else {
			logger.Infof("The HTTP Server [%s] is running on %s", name, server.Addr)
		}
	}

	var err error
	// server.RegisterOnShutdown(r.Stop)
	r.RegisterOnShutdown(func() {
		if logger == nil {
			return
		}

		if err == nil || err == http.ErrServerClosed {
			if name == "" {
				logger.Infof("The HTTP Server is shutdown")
			} else {
				logger.Infof("The HTTP Server [%s] is shutdown", name)
			}
		} else {
			if name == "" {
				logger.Errorf("The HTTP Server is shutdown: %s", err)
			} else {
				logger.Errorf("The HTTP Server [%s] is shutdown: %s", name, err)
			}
		}
	})

	go r.handleSignals()

	var isTLS bool
	if certFile != "" && keyFile != "" {
		isTLS = true
	} else if server.TLSConfig != nil &&
		(len(server.TLSConfig.Certificates) > 0 ||
			server.TLSConfig.GetCertificate != nil) {
		isTLS = true
	}

	if isTLS {
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = server.ListenAndServe()
	}
}
