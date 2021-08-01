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
	"sync/atomic"
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

// OnceRunner is used to run the task only once, which is different from
// sync.Once, the second calling does not wait until the first calling finishes.
type OnceRunner struct {
	done uint32
	task func()
}

// NewOnceRunner returns a new OnceRunner.
func NewOnceRunner(task func()) *OnceRunner { return &OnceRunner{task: task} }

// Run runs the task.
func (r *OnceRunner) Run() {
	if atomic.CompareAndSwapUint32(&r.done, 0, 1) {
		r.task()
	}
}

// Runner is a HTTP Server runner.
type Runner struct {
	Name      string
	Logger    Logger
	Server    *http.Server
	Signals   []os.Signal
	ConnState func(net.Conn, http.ConnState)

	done   chan struct{}
	shut   *OnceRunner
	stop   *OnceRunner
	stopfs []*OnceRunner
}

// StartServer is convenient function to new a runner to start the http server.
func StartServer(addr string, handler http.Handler) {
	NewRunner(handler).Start(addr)
}

// StartServerTLS is the same as StartServer, and tries to start the http server
// with the cert and key file. If certFile or keyFile is empty, however, it is
// equal to StartServer.
func StartServerTLS(addr string, handler http.Handler, certFile, keyFile string) {
	NewRunner(handler).Start(addr, certFile, keyFile)
}

// NewRunner returns a new Runner.
//
// If the handler is a ship, it will set the name and the logger of runner
// to those of ship.
func NewRunner(handler http.Handler) *Runner {
	var name string
	var logger Logger
	if s, ok := handler.(*Ship); ok {
		name = s.Name
		logger = s.Logger
	}

	r := &Runner{
		Name:    name,
		Logger:  logger,
		Server:  &http.Server{Handler: handler},
		Signals: DefaultSignals,
		done:    make(chan struct{}),
	}

	r.shut = NewOnceRunner(r.runShutdown)
	r.stop = NewOnceRunner(r.runStopfs)
	return r
}

// Link registers the shutdown function between itself and other.
func (r *Runner) Link(other *Runner) {
	other.RegisterOnShutdown(r.Stop)
	r.RegisterOnShutdown(other.Stop)
}

// RegisterOnShutdown registers some shutdown functions to run
// when the http server is shut down.
func (r *Runner) RegisterOnShutdown(functions ...func()) {
	for _, f := range functions {
		r.stopfs = append(r.stopfs, NewOnceRunner(f))
	}
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

// Start starts a HTTP server with addr until it is closed.
//
// If tlsFiles is not nil, it must be certFile and keyFile. For example,
//    runner := NewRunner()
//    runner.Start(":80", certFile, keyFile)
func (r *Runner) Start(addr string, tlsFiles ...string) {
	var cert, key string
	if len(tlsFiles) == 2 && tlsFiles[0] != "" && tlsFiles[1] != "" {
		cert = tlsFiles[0]
		key = tlsFiles[1]
	}

	if addr != "" {
		r.Server.Addr = addr
	}

	r.startServer(cert, key)
}

func (r *Runner) startServer(certFile, keyFile string) {
	if r.Server.Addr == "" {
		panic("Runner: Server.Addr is empty")
	} else if r.Server.Handler == nil {
		panic("Runner: Server.Handler is nil")
	}

	name, server := r.Name, r.Server
	if name == "" {
		r.infof("The HTTP Server is running on %s", server.Addr)
	} else {
		r.infof("The HTTP Server [%s] is running on %s", name, server.Addr)
	}

	go r.handleSignals(r.done)

	var isTLS bool
	if certFile != "" && keyFile != "" {
		isTLS = true
	} else if server.TLSConfig != nil &&
		(len(server.TLSConfig.Certificates) > 0 ||
			server.TLSConfig.GetCertificate != nil) {
		isTLS = true
	}

	var err error
	if isTLS {
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = server.ListenAndServe()
	}

	r.Stop()
	<-r.done

	if err == nil || err == http.ErrServerClosed {
		if name == "" {
			r.infof("The HTTP Server listening on %s is shutdown", server.Addr)
		} else {
			r.infof("The HTTP Server [%s] listening on %s is shutdown",
				name, server.Addr)
		}
	} else {
		if name == "" {
			r.errorf("The HTTP Server listening on %s is shutdown: %s",
				server.Addr, err)
		} else {
			r.errorf("The HTTP Server [%s] listening on %s is shutdown: %s",
				name, server.Addr, err)
		}
	}
}

func (r *Runner) infof(format string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Infof(format, args...)
	}
}

func (r *Runner) errorf(format string, args ...interface{}) {
	if r.Logger != nil {
		r.Logger.Errorf(format, args...)
	}
}

func (r *Runner) handleSignals(exit <-chan struct{}) {
	if len(r.Signals) > 0 {
		ss := make(chan os.Signal, 1)
		signal.Notify(ss, r.Signals...)

		select {
		case <-exit:
			return
		case <-ss:
			r.Stop()
		}
	}
}
