package redis

import (
	"expvar"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DLag/cachery"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

type RedisCache struct {
	name       string
	config     cachery.Config
	client     *redis.Pool
	updating   int32
	fetchLock  sync.RWMutex
	expvar     *expvar.Map
	serializer cachery.Serializer
}

func New(name string, redis *redis.Pool, config cachery.Config) *RedisCache {
	cache := new(RedisCache)
	cache.name = name
	cache.client = redis
	cache.expvar = config.ExpVar
	cache.config = config
	cache.serializer = config.Serializer
	return cache
}

func DefaultPool(host string, maxIdle int, idleTimeout time.Duration) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", host) },
	}
}

func (c *RedisCache) Name() string {
	return c.name
}

func (c *RedisCache) Get(key interface{}, obj interface{}, fetcher cachery.Fetcher) error {
	attempts := 0
	for {
		// Trying to get item from Redis server
		attempts++
		val, ttl, err := c.get(cachery.Key(key))
		if err == nil {
			// Item isn't expired
			err = c.serializer.Deserialize(val, obj)
			// If object is expired but still alive use stale value but start background update
			if (int(c.config.Lifetime.Seconds()) - int(c.config.Expire.Seconds())) > ttl {
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
			return errors.New("Cannot get data from cache on second attempt.")
		}
	}
}

func (c *RedisCache) Invalidate(key interface{}) error {
	return c.del(cachery.Key(key))
}

func (c *RedisCache) InvalidateTags(tags ...string) {
	for _, t := range tags {
		for _, ct := range c.config.Tags {
			if ct == t {
				_ = c.delSet()
				return
			}
		}
	}
}

func (c *RedisCache) InvalidateAll() {
	_ = c.delSet()
}

func (c *RedisCache) set(key string, val []byte) (err error) {
	client := c.client.Get()
	defer func() {
		e := client.Close()
		if err == nil {
			err = e
		}
	}()
	if err = client.Send("SADD", c.name, c.name+":"+key); err != nil {
		return
	}
	if err = client.Send("SET", c.name+":"+key, val); err != nil {
		return
	}
	if err = client.Send("EXPIRE", c.name+":"+key, c.config.Lifetime.Seconds()); err != nil {
		return
	}
	if err = client.Flush(); err != nil {
		return
	}
	c.expvarAdd("sets", 1)
	return
}

func (c *RedisCache) get(key string) (val []byte, ttl int, err error) {
	client := c.client.Get()
	defer func() {
		e := client.Close()
		if err == nil {
			err = e
		}
	}()
	val, err = redis.Bytes(client.Do("GET", c.name+":"+key))
	if err != nil {
		return
	}
	ttl, err = redis.Int(client.Do("TTL", c.name+":"+key))
	if err != nil {
		return
	}
	c.expvarAdd("gets", 1)
	return val, ttl, nil
}

func (c *RedisCache) delSet() (err error) {
	client := c.client.Get()
	defer func() {
		e := client.Close()
		if err == nil {
			err = e
		}
	}()
	members, err := redis.Strings(client.Do("SMEMBERS", c.name))
	if err != nil {
		return err
	}
	for _, m := range members {
		_ = client.Send("SREM", c.name, m)
		_ = client.Send("DEL", m)
	}
	err = client.Flush()
	return
}

func (c *RedisCache) del(key string) (err error) {
	client := c.client.Get()
	defer func() {
		e := client.Close()
		if err == nil {
			err = e
		}
	}()
	_ = client.Send("SREM", c.name, c.name+":"+key)
	_ = client.Send("DEL", c.name+":"+key)
	err = client.Flush()
	c.expvarAdd("deletes", 1)
	return
}

func (c *RedisCache) expvarAdd(key string, delta int64) {
	if c.expvar != nil {
		c.expvar.Add(key, delta)
	}
}

func (c *RedisCache) fetch(key interface{}, fetcher cachery.Fetcher) {
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
		err = c.set(cachery.Key(key), val)
		if err != nil {
			c.expvarAdd("fetch_write_to_cache_errors", 1)
			return
		}
		c.expvarAdd("fetches", 1)
	} else {
		// Waiting for another fetch
		c.fetchLock.RLock()
		c.fetchLock.RUnlock()
	}
}
