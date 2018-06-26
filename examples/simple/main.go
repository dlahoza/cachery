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
	"errors"
	"log"
	"time"

	"github.com/DLag/cachery"
	"github.com/DLag/cachery/drivers/inmemory"
)

const (
	cacheName = "some_cache"
	keyName   = "some_key"
	data      = "some_data"
)

var errVar = errors.New("some error")

func fetcher(key interface{}) (interface{}, error) {
	log.Print("==> Database hit")
	if key == keyName {
		time.Sleep(time.Millisecond * 500)
		log.Print("<== Database response")
		return data, nil
	}
	return nil, errVar
}

func init() {
	cachery.Add(cachery.NewDefault(cacheName, cachery.Config{
		Serializer: &cachery.GobSerializer{},
		Driver:     inmemory.Default(),
		Expire:     time.Second,
		Lifetime:   2 * time.Second,
	}))
}

func main() {
	c := cachery.Get(cacheName)
	var val string
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Print("First Get is slow")
	c.Get(keyName, &val, fetcher)
	log.Printf("%q=%q", keyName, val)

	log.Print("Second is from cache")
	c.Get(keyName, &val, fetcher)
	log.Printf("%q=%q", keyName, val)

	log.Printf("If we wait for Expire timeout it gets from stale cache with background DB fetch")
	time.Sleep(time.Second + time.Millisecond*50)
	c.Get(keyName, &val, fetcher)
	log.Printf("%q=%q", keyName, val)
	time.Sleep(time.Millisecond * 500)

	log.Print("If we wait for Lifetime timeout it gets from DB")
	time.Sleep(2*time.Second + time.Millisecond*50)
	c.Get(keyName, &val, fetcher)
	log.Printf("%q=%q", keyName, val)
}
