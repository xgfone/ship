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

	err    error
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
// If the handler has implemented the interface { GetName() string },
// it will set the name to handler.GetName().
//
// If the handler has implemented the interface { GetLogger() Logger },
// it will set the logger to handler.GetLogger().
func NewRunner(handler http.Handler) *Runner {
	var name string
	if h, ok := handler.(interface{ GetName() string }); ok {
		name = h.GetName()
	}

	var logger Logger
	if h, ok := handler.(interface{ GetLogger() Logger }); ok {
		logger = h.GetLogger()
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

// SetName sets the name to name and returns itself.
func (r *Runner) SetName(name string) *Runner {
	r.Name = name
	return r
}

// SetLogger sets the logger to logger and returns itself.
func (r *Runner) SetLogger(logger Logger) *Runner {
	r.Logger = logger
	return r
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
	r.logShutdown()
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

	if r.Name == "" {
		r.infof("The HTTP Server is running on %s", r.Server.Addr)
	} else {
		r.infof("The HTTP Server [%s] is running on %s", r.Name, r.Server.Addr)
	}

	go r.handleSignals(r.done)

	var isTLS bool
	if certFile != "" && keyFile != "" {
		isTLS = true
	} else if r.Server.TLSConfig != nil &&
		(len(r.Server.TLSConfig.Certificates) > 0 ||
			r.Server.TLSConfig.GetCertificate != nil) {
		isTLS = true
	}

	if isTLS {
		r.err = r.Server.ListenAndServeTLS(certFile, keyFile)
	} else {
		r.err = r.Server.ListenAndServe()
	}

	r.Stop()
	<-r.done
}

func (r *Runner) logShutdown() {
	if r.err == nil || r.err == http.ErrServerClosed {
		if r.Name == "" {
			r.infof("The HTTP Server listening on %s is shutdown",
				r.Server.Addr)
		} else {
			r.infof("The HTTP Server [%s] listening on %s is shutdown",
				r.Name, r.Server.Addr)
		}
	} else {
		if r.Name == "" {
			r.errorf("The HTTP Server listening on %s is shutdown: %s",
				r.Server.Addr, r.err)
		} else {
			r.errorf("The HTTP Server [%s] listening on %s is shutdown: %s",
				r.Name, r.Server.Addr, r.err)
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
