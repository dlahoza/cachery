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

cachery.Add(cachery.NewDefault(cacheName, cachery.Config{
    Serializer: &cachery.GobSerializer{},
    Driver:     inmemory.Default(),
    Fetcher:    fetcher,
    Expire:     time.Second,
    Lifetime:   2 * time.Second,
}))

c := cachery.Get("some_cache")
var val string

c.Get("some_key", &val, nil)
// Or override fetcher function from config
c.Get("some_key", &val, fetcher)
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