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
	"expvar"
	"time"
)

// Fetcher is a function which returns data from origin data store
type Fetcher func(key interface{}) (interface{}, error)

// Cache describes cache object
type Cache interface {
	// Get loads data to dst from cache or from fetcher function
	Get(key interface{}, dst interface{}, fetcher Fetcher) error
	// Name returns name of the cache
	Name() string
	// Invalidate specific key
	Invalidate(key interface{}) error
	// InvalidateTags invalidates cache if finds necessary tags
	InvalidateTags(tags ...string)
	// InvalidateAll invalidates all data from this cache
	InvalidateAll()
}

// Driver describes storage driver interface
type Driver interface {
	// Get loads key from the cache store if it is not outdated
	Get(cacheName string, key interface{}) (val []byte, ttl time.Duration, err error)
	// Set saves key to the cache store
	Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error)
	// Invalidate removes the key from the cache store
	Invalidate(cacheName string, key interface{}) error
	// InvalidateAll removes all keys from the cache store
	InvalidateAll(cacheName string)
}

// Config describes configuration of cache
type Config struct {
	// Expire when data in cache becomes stale but still usable and needs to be updated from fetcher
	Expire time.Duration
	// Lifetime when data in cache becomes outdated and needs to be updated from fetcher before use
	Lifetime time.Duration
	// Tags of the cache
	Tags []string
	// Serializer for objects
	Serializer Serializer
	// Driver cache storage driver (e.g. Redis, Memcached, Memory)
	Driver Driver
	// Fetcher optional instance of Fetcher function, could be nil if fetcher parameter of Get function is used
	Fetcher Fetcher
	// Driver cache storage driver (e.g. Redis, Memcached, Memory)
	Expvar *expvar.Map
}
