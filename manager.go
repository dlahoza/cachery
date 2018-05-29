package cachery

import (
	"sync"
)

var (
	caches Manager
)

// Manager consolidates caches and allows manipulations on them
type Manager struct {
	caches map[string]Cache
	sync.Mutex
}

// Add cache to Manager
func (m *Manager) Add(cache Cache) {
	m.Lock()
	if m.caches == nil {
		m.caches = make(map[string]Cache)
	}
	m.caches[cache.Name()] = cache
	m.Unlock()
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

// Add cache to the internal cache Manager
func Add(cache Cache) {
	caches.Add(cache)
}

// Get cache from the internal Manager by its name or returns nil if could not find it
func Get(name string) Cache {
	return caches.Get(name)
}

// InvalidateTags invalidates caches of internal Manager which have specific tags
func InvalidateTags(tags ...string) {
	caches.InvalidateTags(tags...)
}

// InvalidateAll invalidates all caches of internal Manager
func InvalidateAll() {
	caches.InvalidateAll()
}
