package utils

import (
	"bytes"
	"sync"
	"testing"
)

func TestCallOnExit(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	OnExit(func() { buf.WriteString("call") })

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			CallOnExit()
			wg.Done()
		}()
	}
	wg.Wait()

	if buf.String() != "call" {
		t.Error(buf.String())
	}
}
