// Copyright (c) 2018 Dmytro Lahoza <dmitry@lagoza.name>
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cachery

import (
	"sync"
)

var (
	caches Manager
	// Add cache to the internal cache Manager
	Add = caches.Add
	// Get cache from the internal Manager by its name or returns nil if could not find it
	Get = caches.Get
	// InvalidateTags invalidates caches of internal Manager which have specific tags
	InvalidateTags = caches.InvalidateTags
	// InvalidateAll invalidates all caches of internal Manager
	InvalidateAll = caches.InvalidateAll
)

// Manager consolidates caches and allows manipulations on them
type Manager struct {
	caches map[string]Cache
	sync.Mutex
}

// Add cache to Manager
func (m *Manager) Add(cache ...Cache) *Manager {
	m.Lock()
	if m.caches == nil {
		m.caches = make(map[string]Cache)
	}
	for i := range cache {
		m.caches[cache[i].Name()] = cache[i]
	}
	m.Unlock()
	return m
}

// Get cache from Manager by its name or returns nil if could not find it
func (m *Manager) Get(name string) Cache {
	m.Lock()
	defer m.Unlock()
	if c, ok := m.caches[name]; ok {
		return c
	}
	return nil
}

// InvalidateTags invalidates caches which have specific tags
func (m *Manager) InvalidateTags(tags ...string) {
	m.Lock()
	defer m.Unlock()
	for i := range m.caches {
		m.caches[i].InvalidateTags(tags...)
	}
}

// InvalidateAll invalidates all caches in Manager
func (m *Manager) InvalidateAll() {
	m.Lock()
	defer m.Unlock()
	for i := range m.caches {
		m.caches[i].InvalidateAll()
	}
}
