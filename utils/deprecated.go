// Copyright 2019 xgfone
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

package utils

import (
	"bytes"
	"sync"

	"github.com/xgfone/go-tools/function"
	"github.com/xgfone/go-tools/io2"
	"github.com/xgfone/go-tools/lifecycle"
	"github.com/xgfone/go-tools/types"
)

// Some function aliases.
//
// DEPRECATED!!! They won't be removed until the next major version.
var (
	// Converting functions
	ToBool    = types.ToBool
	ToInt64   = types.ToInt64
	ToUint64  = types.ToUint64
	ToFloat64 = types.ToFloat64
	ToString  = types.ToString

	// IO functions
	ReadN       = io2.ReadN
	ReadNWriter = io2.ReadNWriter

	// Setter functions
	SetValue        = function.SetValue
	SetStructValue  = function.SetStructValue
	BindMapToStruct = function.BindMapToStruct

	// OnExit functions
	CallOnExit = lifecycle.Stop
	OnExit     = func(f ...func()) { lifecycle.Register(f...) }
	Exit       = lifecycle.Exit
)

// SetValuer is the interface alias of function.SetValuer.
//
// DEPRECATED!!! It won't be removed until the next major version.
type SetValuer function.SetValuer

//////////////////////////////////////////////////////////////////////////////

// BufferPool is the bytes.Buffer wrapper of sync.Pool.
//
// DEPRECATED!!! It won't be removed until the next major version.
type BufferPool struct {
	pool *sync.Pool
	size int
}

func makeBuffer(size int) (b *bytes.Buffer) {
	b = bytes.NewBuffer(make([]byte, size))
	b.Reset()
	return
}

// NewBufferPool returns a new bytes.Buffer pool.
//
// bufSize is the initializing size of the buffer. If the size is equal to
// or less than 0, it will be ignored, and use the default size, 1024.
//
// DEPRECATED!!! It won't be removed until the next major version.
func NewBufferPool(bufSize ...int) BufferPool {
	size := 1024
	if len(bufSize) > 0 && bufSize[0] > 0 {
		size = bufSize[0]
	}

	return BufferPool{
		size: size,
		pool: &sync.Pool{New: func() interface{} { return makeBuffer(size) }},
	}
}

// Get returns a bytes.Buffer.
func (p BufferPool) Get() *bytes.Buffer {
	x := p.pool.Get()
	if x == nil {
		return makeBuffer(p.size)
	}
	return x.(*bytes.Buffer)
}

// Put places a bytes.Buffer to the pool.
func (p BufferPool) Put(b *bytes.Buffer) {
	if b != nil {
		b.Reset()
		p.pool.Put(b)
	}
}
