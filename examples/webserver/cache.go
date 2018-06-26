/*
 * Copyright (c) 2018 Dmytro Lahoza <dmitry@lagoza.name>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package main

import (
	"expvar"
	"time"

	"github.com/DLag/cachery"
	"github.com/DLag/cachery/drivers/inmemory"
	"github.com/DLag/cachery/drivers/redis"
	"github.com/DLag/cachery/wrappers/nats"
)

var expvarMap = expvar.NewMap("cachery")

func initRedisCache() {
	driver := redis.New(redis.DefaultPool("127.0.0.1:6379", 3, time.Second*120))
	cachery.Add(generateCacheInstance("redis", driver)...)
}

func initInMemCache() {
	driver := inmemory.Default()
	cachery.Add(generateCacheInstance("inmemory", driver)...)
}

func initInMemNATSCache() {
	driver := nats.Default(inmemory.Default(), "nats://localhost:4222", "cachery")
	cachery.Add(generateCacheInstance("nats", driver)...)
}

func generateCacheInstance(name string, driver cachery.Driver) []cachery.Cache {
	serializer := new(cachery.GobSerializer)
	return []cachery.Cache{
		// Returns DefaultCache caching logic module which
		cachery.NewDefault(
			// Name of the cache
			name+"_SHORT_CACHE",
			cachery.Config{
				// Time after cache become stale and should be updated from fetcher in background
				Expire: time.Second * 20,
				// Time after cache become outdated and should be updated immediately
				Lifetime: time.Second * 30,
				// Serializer is reusable.
				// There is JSON serializer as well, but it is slower and has some limitations like nanoseconds in time.Time.
				Serializer: serializer,
				// Driver is reusable for different caches
				Driver: driver,
				// Fetcher is function that fetch data from the underlying storage(e.g. database)
				// could be nil if you use fetcher parameter of Get function
				Fetcher: fetcherOrders,
				// Expvar will be used to populate cache statistics through expvar package
				// It could be nil if you don't need it
				Expvar: expvarMap,
				// Tags allow you invalidate all caches in Manager which have specified tags
				Tags: []string{"orders", "goods"},
			}),
		cachery.NewDefault(
			name+"_LONG_CACHE",
			cachery.Config{
				Expire:     time.Minute * 10,
				Lifetime:   time.Minute * 60,
				Serializer: serializer,
				Driver:     driver,
				Fetcher:    fetcherGoods,
				Expvar:     expvarMap,
				Tags:       []string{"goods"},
			}),
	}
}
