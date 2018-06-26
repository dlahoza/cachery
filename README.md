Cachery - caching framework
================================

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2FDLag%2Fcachery.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2FDLag%2Fcachery?ref=badge_shield)
[![Build Status](https://travis-ci.org/DLag/cachery.svg?branch=master)](https://travis-ci.org/DLag/cachery)
[![Coverage Status](https://coveralls.io/repos/github/DLag/cachery/badge.svg?branch=master)](https://coveralls.io/github/DLag/cachery?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/DLag/cachery)](https://goreportcard.com/report/github.com/DLag/cachery)
[![GoDoc](https://godoc.org/github.com/DLag/cachery?status.svg)](http://godoc.org/github.com/DLag/cachery)

Modular and feature-rich caching framework

Features include:

* Multiple cache pools
* Invalidation grouping with tags
* Cluster support
* Expire and stale behavior
* Expvar support
* Modularity:
  * serializers
  * cache logic modules
  * storage drivers
  * driver wrappers

## Quickstart
### Install/Update
 go get -u github.com/DLag/cachery
### Basic usage
```go
func fetcher(key interface{}) (interface{}, error) {
    res, err := db.GetDataByKey(key)
    return res, err
}

// Add cache to the cache manager
cachery.Add(cachery.NewDefault("some_cache", cachery.Config{
    // Time after cache become stale and should be updated from fetcher in background
    Expire: time.Second * 20,
    // Time after cache become outdated and should be updated immediately
    Lifetime: time.Second * 120,
    // Serializer is reusable.
    // There is JSON serializer as well, but it is slower and has some limitations like nanoseconds in time.Time.
    Serializer: &cachery.GobSerializer{},
    // Driver is reusable for different caches
    Driver:     inmemory.Default(),
    // Fetcher is function that fetch data from the underlying storage(e.g. database)
    // could be nil if you use fetcher parameter of Get function
    Fetcher:    fetcher,
    // Expvar will be used to populate cache statistics through expvar package
    // It could be nil if you don't need it
    Expvar: nil,
    // Tags allow you invalidate all caches in Manager which have specified tags
    // could be nil
    Tags: []string{"tag1", "tag2"},
}))

// Get cache from manager
c := cachery.Get("some_cache")
var val string

c.Get("some_key", &val, nil)
// Or override fetcher function from config
c.Get("some_key", &val, fetcher)

// Invalidate all cache
c.InvalidateAll()
// Invalidate single key
c.Invalidate("some_key")

//Invalidate all caches in manager
cachery.InvalidateAll()
//Invalidate caches in manager by tag
cachery.InvalidateTags("tag1")
```

## Examples
See examples to understand usage:
* [Simple](examples/simple)
* [Webserver](examples/webserver)

## Staying up to date

To update Cachery to the latest version, use `go get -u github.com/DLag/cachery`.

## Supported go versions

We support the three major Go versions, which are 1.8, 1.9 and 1.10 at the moment.

## Contributing

Please feel free to submit issues, fork the repository and send pull requests!

When submitting an issue, we ask that you please include a complete test function that demonstrates the issue.  Extra kudos for those using Cachery to write the test code that demonstrates it.