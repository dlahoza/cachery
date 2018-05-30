package inmemory_nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DLag/cachery"
	"github.com/DLag/cachery/drivers/inmemory"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
)

var DefaultTimeout = time.Minute

type Driver struct {
	inmemory *inmemory.Driver
	nats     *nats.Conn
	subject  string
}

type message struct {
	Command   string
	CacheName string
	Key       string
}

func New(gctimeout time.Duration, nats *nats.Conn, subject string) *Driver {
	driver := new(Driver)
	driver.inmemory = inmemory.New(gctimeout)
	driver.nats = nats
	driver.subject = subject
	_, err := driver.nats.Subscribe(driver.subject, driver.consumer)
	if err != nil {
		panic(err)
	}
	return driver
}

func Default(natsURL, subject string) *Driver {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return New(DefaultTimeout, conn, subject)
}

func (c *Driver) Invalidate(cacheName string, key interface{}) error {
	msg := message{
		Command:   "Invalidate",
		CacheName: cacheName,
		Key:       cachery.Key(key),
	}
	return c.send(msg)
}

func (c *Driver) InvalidateAll(cacheName string) {
	msg := message{
		Command:   "InvalidateAll",
		CacheName: cacheName,
	}
	_ = c.send(msg)
}

func (c *Driver) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	return c.inmemory.Set(cacheName, cachery.Key(key), val, ttl)
}

func (c *Driver) Get(cacheName string, key interface{}) (val []byte, ttl time.Duration, err error) {
	return c.inmemory.Get(cacheName, cachery.Key(key))
}

func (c *Driver) send(msg message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "In-memory NATS driver: cannot marshal message")
	}
	err = c.nats.Publish(c.subject, data)
	if err != nil {
		return errors.Wrap(err, "In-memory NATS driver: cannot send message")
	}
	return nil
}

func (c *Driver) consumer(m *nats.Msg) {
	fmt.Printf("Received a message: %s\n", string(m.Data))
	var msg message
	err := json.Unmarshal(m.Data, &msg)
	if err != nil {
		return
	}
	switch msg.Command {
	case "Invalidate":
		c.Invalidate(msg.CacheName, msg.Key)
	case "InvalidateAll":
		c.InvalidateAll(msg.CacheName)
	}
}
