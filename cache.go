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
	// ExpVar expvar.Map to save cache stats
	ExpVar *expvar.Map
}
