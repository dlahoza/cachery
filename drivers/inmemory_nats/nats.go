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

package inmemory_nats

import (
	"time"

	"github.com/DLag/cachery"
	"github.com/DLag/cachery/drivers/inmemory"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Driver type satisfies cachery.Driver interface
type Driver struct {
	inmemory *inmemory.Driver
	nats     *nats.EncodedConn
	subject  string
	id       string
}

type message struct {
	Sender    string
	Command   string
	CacheName string
	Key       string
}

// New creates an instance of Driver
func New(gctimeout time.Duration, nc *nats.Conn, subject string) *Driver {
	driver := new(Driver)
	driver.inmemory = inmemory.New(gctimeout)
	u, _ := uuid.NewV4()
	driver.id = u.String()
	driver.nats, _ = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	driver.subject = subject
	_, err := driver.nats.Subscribe(driver.subject, driver.consumer)
	if err != nil {
		panic(err)
	}
	return driver
}

// Default creates an instance of Driver with default GC timeout
func Default(natsURL, subject string) *Driver {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return New(inmemory.DefaultTimeout, conn, subject)
}

// Invalidate removes the key from the cache store
// it's atomic only for local data
func (c *Driver) Invalidate(cacheName string, key interface{}) error {
	k := cachery.Key(key)
	err := c.inmemory.Invalidate(cacheName, k)
	if err != nil {
		return err
	}
	msg := message{
		Sender:    c.id,
		Command:   "Invalidate",
		CacheName: cacheName,
		Key:       cachery.Key(k),
	}
	return c.send(msg)
}

// InvalidateAll removes all keys from the cache store
// it's atomic only for local data
func (c *Driver) InvalidateAll(cacheName string) {
	c.inmemory.InvalidateAll(cacheName)
	msg := message{
		Sender:    c.id,
		Command:   "InvalidateAll",
		CacheName: cacheName,
	}
	_ = c.send(msg)
}

// Set saves key to the cache store
func (c *Driver) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	return c.inmemory.Set(cacheName, cachery.Key(key), val, ttl)
}

// Get loads key from the cache store if it is not outdated
func (c *Driver) Get(cacheName string, key interface{}) (val []byte, ttl time.Duration, err error) {
	return c.inmemory.Get(cacheName, cachery.Key(key))
}

func (c *Driver) send(msg message) error {
	err := c.nats.Publish(c.subject, msg)
	if err != nil {
		return errors.Wrap(err, "In-memory NATS driver: cannot send message")
	}
	return nil
}

func (c *Driver) consumer(msg *message) {
	// Skip its own messages
	if c.id == msg.Sender {
		return
	}
	switch msg.Command {
	case "Invalidate":
		_ = c.inmemory.Invalidate(msg.CacheName, msg.Key)
	case "InvalidateAll":
		c.inmemory.InvalidateAll(msg.CacheName)
	}
}
