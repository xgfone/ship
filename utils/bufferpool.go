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

package utils

import (
	"bytes"
	"sync"
)

// BufferPool is the bytes.Buffer wrapper of sync.Pool
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
