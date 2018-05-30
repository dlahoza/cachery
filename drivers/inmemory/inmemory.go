package inmemory

import (
	"time"

	"sync"

	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("Item not found")
var DefaultTimeout = time.Minute

type Driver struct {
	storage     map[string]map[interface{}]*item
	storageLock sync.RWMutex
	gctimer     *time.Timer
}

type item struct {
	value    []byte
	deadline time.Time
}

type path struct {
	cacheName string
	key       interface{}
}

func New(gctimeout time.Duration) *Driver {
	driver := new(Driver)
	driver.storage = make(map[string]map[interface{}]*item)
	driver.gc(gctimeout)
	return driver
}

func Default() *Driver {
	return New(DefaultTimeout)
}

func (c *Driver) Invalidate(cacheName string, key interface{}) error {
	c.storageLock.Lock()
	if _, ok := c.storage[cacheName]; ok {
		delete(c.storage[cacheName], key)
	}
	c.storageLock.Unlock()
	return nil
}

func (c *Driver) InvalidateAll(cacheName string) {
	c.storageLock.Lock()
	delete(c.storage, cacheName)
	c.storageLock.Unlock()
}

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
