package cachery

import (
	"expvar"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

// DefaultCache default implementation of caching logic
type DefaultCache struct {
	name       string
	config     Config
	driver     Driver
	updating   int32
	fetchLock  sync.RWMutex
	expvar     *expvar.Map
	serializer Serializer
}

func NewDefault(name string, config Config, driver Driver, expvar *expvar.Map) *DefaultCache {
	cache := new(DefaultCache)
	cache.name = name
	cache.driver = driver
	cache.expvar = expvar
	cache.config = config
	cache.serializer = config.Serializer
	return cache
}

func (c *DefaultCache) Name() string {
	return c.name
}

func (c *DefaultCache) Get(key interface{}, obj interface{}, fetcher Fetcher) error {
	attempts := 0
	for {
		// Trying to get item from Redis server
		attempts++
		val, ttl, err := c.driver.Get(c.name, key)
		c.expvarAdd("gets", 1)
		if err == nil {
			// Item isn't expired
			err = c.serializer.Deserialize(val, obj)
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

func (c *DefaultCache) Invalidate(key interface{}) error {
	c.expvarAdd("invalidate_key", 1)
	return c.driver.Invalidate(c.name, key)
}

func (c *DefaultCache) InvalidateTags(tags ...string) {
	c.expvarAdd("invalidate_tags", 1)
	for _, t := range tags {
		for _, ct := range c.config.Tags {
			if ct == t {
				c.driver.InvalidateAll(c.name)
				return
			}
		}
	}
}

func (c *DefaultCache) InvalidateAll() {
	c.expvarAdd("invalidate_all", 1)
	c.driver.InvalidateAll(c.name)
}

func (c *DefaultCache) expvarAdd(key string, delta int64) {
	if c.expvar != nil {
		c.expvar.Add(key, delta)
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
		val, err := c.serializer.Serialize(obj)
		if err != nil {
			c.expvarAdd("fetch_serialize_errors", 1)
			return
		}
		err = c.driver.Set(c.name, key, val, c.config.Lifetime)
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
