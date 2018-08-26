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

package nats

import (
	"time"

	"github.com/DLag/cachery"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
)

// Wrapper type satisfies cachery.Wrapper interface
type Wrapper struct {
	cachery.Driver
	nats    *nats.EncodedConn
	subject string
	id      string
}

type message struct {
	Sender    string
	Command   string
	CacheName string
	Key       string
}

// New creates an instance of Wrapper
func New(driver cachery.Driver, nc *nats.Conn, subject string) *Wrapper {
	wrapper := new(Wrapper)
	wrapper.Driver = driver
	u, _ := uuid.NewV4()
	wrapper.id = u.String()
	wrapper.nats, _ = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	wrapper.subject = subject
	_, err := wrapper.nats.Subscribe(wrapper.subject, wrapper.consumer)
	if err != nil {
		panic(err)
	}
	return wrapper
}

// Default creates an instance of Wrapper with driver
func Default(driver cachery.Driver, natsURL, subject string) *Wrapper {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return New(driver, conn, subject)
}

// Invalidate removes the key from the cache store
// it's atomic only for local data
func (c *Wrapper) Invalidate(cacheName string, key interface{}) error {
	k := cachery.Key(key)
	err := c.Driver.Invalidate(cacheName, k)
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
func (c *Wrapper) InvalidateAll(cacheName string) {
	c.Driver.InvalidateAll(cacheName)
	msg := message{
		Sender:    c.id,
		Command:   "InvalidateAll",
		CacheName: cacheName,
	}
	_ = c.send(msg)
}

// Set saves key to the cache store
func (c *Wrapper) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	return c.Driver.Set(cacheName, cachery.Key(key), val, ttl)
}

// Get loads key from the cache store if it is not outdated
func (c *Wrapper) Get(cacheName string, key interface{}) (val []byte, ttl time.Duration, err error) {
	return c.Driver.Get(cacheName, cachery.Key(key))
}

func (c *Wrapper) send(msg message) error {
	err := c.nats.Publish(c.subject, msg)
	if err != nil {
		return errors.Wrap(err, "NATS wrapper: cannot send message")
	}
	return nil
}

func (c *Wrapper) consumer(msg *message) {
	// Skip its own messages
	if c.id == msg.Sender {
		return
	}
	switch msg.Command {
	case "Invalidate":
		_ = c.Driver.Invalidate(msg.CacheName, msg.Key)
	case "InvalidateAll":
		c.Driver.InvalidateAll(msg.CacheName)
	}
}
