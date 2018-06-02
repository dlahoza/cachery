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
	"sync/atomic"

	"github.com/pkg/errors"
)

// DefaultCache default implementation of caching logic
type DefaultCache struct {
	name      string
	config    Config
	updating  int32
	fetchLock sync.RWMutex
}

// NewDefault creates an instance of DefaultCache
func NewDefault(name string, config Config) *DefaultCache {
	cache := new(DefaultCache)
	cache.name = name
	cache.config = config
	return cache
}

// Name returns name of the cache
func (c *DefaultCache) Name() string {
	return c.name
}

// Get loads data to dst from cache or from fetcher function
func (c *DefaultCache) Get(key interface{}, obj interface{}, fetcher Fetcher) error {
	attempts := 0
	for {
		// Trying to get item from Redis server
		attempts++
		val, ttl, err := c.config.Driver.Get(c.name, key)
		c.expvarAdd("gets", 1)
		if err == nil {
			// Item isn't expired
			err = c.config.Serializer.Deserialize(val, obj)
			// If object is expired but still alive use stale value but start background update
			if (c.config.Lifetime - c.config.Expire) > ttl {
				c.expvarAdd("stale", 1)
				go c.fetch(key, fetcher)
			}
			c.expvarAdd("hits", 1)
			return err
		}
		switch attempts {
		case 1:
			c.fetch(key, fetcher)
		case 2:
			c.expvarAdd("get_after_fetch_errors", 1)
			return errors.Wrap(err, "Cannot get data from cache on second attempt.")
		}
	}
}

// Invalidate specific key
func (c *DefaultCache) Invalidate(key interface{}) error {
	c.expvarAdd("invalidate_key", 1)
	return c.config.Driver.Invalidate(c.name, key)
}

// InvalidateTags invalidates cache if finds necessary tags
func (c *DefaultCache) InvalidateTags(tags ...string) {
	c.expvarAdd("invalidate_tags", 1)
	for _, t := range tags {
		for _, ct := range c.config.Tags {
			if ct == t {
				c.config.Driver.InvalidateAll(c.name)
				return
			}
		}
	}
}

// InvalidateAll invalidates all data from this cache
func (c *DefaultCache) InvalidateAll() {
	c.expvarAdd("invalidate_all", 1)
	c.config.Driver.InvalidateAll(c.name)
}

func (c *DefaultCache) expvarAdd(key string, delta int64) {
	if c.config.Expvar != nil {
		c.config.Expvar.Add(key, delta)
	}
}

func (c *DefaultCache) fetch(key interface{}, fetcher Fetcher) {
	// If it is not updating now
	if atomic.CompareAndSwapInt32(&c.updating, 0, 1) {
		defer atomic.CompareAndSwapInt32(&c.updating, 1, 0)
		c.fetchLock.Lock()
		defer c.fetchLock.Unlock()
		// Getting from fetcher
		obj, err := fetcher(key)
		if err != nil {
			c.expvarAdd("fetch_get_errors", 1)
			return
		}
		// Writing to Redis
		val, err := c.config.Serializer.Serialize(obj)
		if err != nil {
			c.expvarAdd("fetch_serialize_errors", 1)
			return
		}
		err = c.config.Driver.Set(c.name, key, val, c.config.Lifetime)
		c.expvarAdd("sets", 1)
		if err != nil {
			c.expvarAdd("fetch_write_to_cache_errors", 1)
			return
		}
		c.expvarAdd("fetches", 1)
	} else {
		// Waiting for another fetch
		c.fetchLock.RLock()
		c.expvarAdd("fetch_waits", 1)
		c.fetchLock.RUnlock()
	}
}
