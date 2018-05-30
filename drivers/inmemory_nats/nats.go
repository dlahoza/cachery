package inmemory_nats

import (
	"encoding/json"
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
	nats     *nats.Conn
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
func New(gctimeout time.Duration, nats *nats.Conn, subject string) *Driver {
	driver := new(Driver)
	driver.inmemory = inmemory.New(gctimeout)
	uuid, _ := uuid.NewV4()
	driver.id = uuid.String()
	driver.nats = nats
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
	//fmt.Printf("Received a message: %s\n", string(m.Data))
	var msg message
	err := json.Unmarshal(m.Data, &msg)
	if err != nil {
		return
	}
	// Skip its own messages
	if c.id == msg.Sender {
		return
	}
	switch msg.Command {
	case "Invalidate":
		c.inmemory.Invalidate(msg.CacheName, msg.Key)
	case "InvalidateAll":
		c.inmemory.InvalidateAll(msg.CacheName)
	}
}
