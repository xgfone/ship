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

package session

import "sync"

// Session represents an interface about the session.
type Session interface {
	// If the session id does not exist, it should return (nil, nil).
	GetSession(id string) (value interface{}, err error)
	SetSession(id string, value interface{}) error
	DelSession(id string) error
}

// NewMemorySession return a Session implementation based on the memory.
func NewMemorySession() Session {
	return memorySession{store: new(sync.Map)}
}

type memorySession struct {
	store *sync.Map
}

func (m memorySession) GetSession(id string) (value interface{}, err error) {
	if value, ok := m.store.Load(id); ok {
		return value, nil
	}
	return
}

func (m memorySession) SetSession(id string, value interface{}) error {
	m.store.Store(id, value)
	return nil
}

func (m memorySession) DelSession(id string) error {
	m.store.Delete(id)
	return nil
}
