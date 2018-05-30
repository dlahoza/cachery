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

package inmemory

import (
	"time"

	"sync"

	"github.com/pkg/errors"
)

// ErrNotFound item not found in the cache store
var ErrNotFound = errors.New("Item not found")

// DefaultTimeout default timeout for cache GC
var DefaultTimeout = time.Minute

// Driver type satisfies cachery.Driver interface
type Driver struct {
	storage     map[string]map[interface{}]*item
	storageLock sync.RWMutex
}

type item struct {
	value    []byte
	deadline time.Time
}

type path struct {
	cacheName string
	key       interface{}
}

// New creates an instance of Driver type
func New(gctimeout time.Duration) *Driver {
	driver := new(Driver)
	driver.storage = make(map[string]map[interface{}]*item)
	driver.gc(gctimeout)
	return driver
}

// Default creates an instance of Driver with default GC timeout
func Default() *Driver {
	return New(DefaultTimeout)
}

// Invalidate removes the key from the cache store
func (c *Driver) Invalidate(cacheName string, key interface{}) error {
	c.storageLock.Lock()
	if _, ok := c.storage[cacheName]; ok {
		delete(c.storage[cacheName], key)
	}
	c.storageLock.Unlock()
	return nil
}

// InvalidateAll removes all keys from the cache store
func (c *Driver) InvalidateAll(cacheName string) {
	c.storageLock.Lock()
	delete(c.storage, cacheName)
	c.storageLock.Unlock()
}

// Set saves key to the cache store
func (c *Driver) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	c.storageLock.Lock()
	if _, ok := c.storage[cacheName]; !ok {
		c.storage[cacheName] = make(map[interface{}]*item)
	}
	i := new(item)
	i.value = make([]byte, len(val))
	copy(i.value, val)
	i.deadline = time.Now().Add(ttl)
	c.storage[cacheName][key] = i
	c.storageLock.Unlock()
	return nil
}

// Get loads key from the cache store if it is not outdated
func (c *Driver) Get(cacheName string, key interface{}) (val []byte, ttl time.Duration, err error) {
	c.storageLock.RLock()
	if _, ok := c.storage[cacheName]; ok {
		if i, ok := c.storage[cacheName][key]; ok {
			if ttl = time.Until(i.deadline); ttl > 0 {
				val = make([]byte, len(i.value))
				copy(val, i.value)
				c.storageLock.RUnlock()
				return
			} else {
				c.storageLock.RUnlock()
				c.sweep([]path{{cacheName, key}})
				return nil, 0, ErrNotFound
			}
		}
	}
	c.storageLock.RUnlock()
	return nil, 0, ErrNotFound
}

func (c *Driver) gc(timeout time.Duration) {
	c.sweep(c.mark())
	time.AfterFunc(timeout, func() {
		c.gc(timeout)
	})
}

func (c *Driver) sweep(p []path) {
	c.storageLock.Lock()
	for pi := range p {
		if _, ok := c.storage[p[pi].cacheName]; ok {
			if i, ok := c.storage[p[pi].cacheName][p[pi].key]; ok {
				if ttl := time.Until(i.deadline); ttl <= 0 {
					delete(c.storage[p[pi].cacheName], p[pi].key)
				}
			}
		}
	}
	c.storageLock.Unlock()
}

func (c *Driver) mark() (marked []path) {
	c.storageLock.RLock()
	for k := range c.storage {
		for v := range c.storage[k] {
			i := c.storage[k][v]
			if time.Until(i.deadline) <= 0 {
				marked = append(marked, path{k, v})
			}
		}
	}
	c.storageLock.RUnlock()
	return
}
