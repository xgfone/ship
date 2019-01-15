package utils

import (
	"os"
	"sync/atomic"
)

var (
	exited int32
	funcs  = make([]func(), 0)
)

// OnExit registers some exit function.
func OnExit(f ...func()) {
	funcs = append(funcs, f...)
}

// CallOnExit calls the exit functions.
//
// This function can be called many times.
func CallOnExit() {
	if atomic.CompareAndSwapInt32(&exited, 0, 1) {
		for _, f := range funcs {
			f()
		}
	}
}

// Exit exits the process with the code, but calling the exit functions
// before exiting.
func Exit(code int) {
	CallOnExit()
	os.Exit(code)
}
