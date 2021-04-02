// Copyright 2020 xgfone
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

import "sync"

var _ RWLocker = &sync.RWMutex{}

// RWLocker represents an object that can be locked and unlocked
// with read and write.
type RWLocker interface {
	RLock()
	RUnlock()
	sync.Locker
}

// NewNoopRWLocker returns a No-Op RLocker, which does nothing.
func NewNoopRWLocker() RWLocker { return noopRWLocker{} }

type noopRWLocker struct{}

func (l noopRWLocker) Lock()    {}
func (l noopRWLocker) Unlock()  {}
func (l noopRWLocker) RLock()   {}
func (l noopRWLocker) RUnlock() {}
