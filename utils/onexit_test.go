package utils

import (
	"bytes"
	"sync"
	"testing"
)

func TestCallOnExit(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	OnExit(func() { buf.WriteString("call1\n") }, func() { buf.WriteString("call2\n") })
	OnExit(CallOnExit)
	OnExit()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			CallOnExit()
			wg.Done()
		}()
	}
	wg.Wait()

	if buf.String() != "call1\ncall2\n" {
		t.Error(buf.String())
	}
}
