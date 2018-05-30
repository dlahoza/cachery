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
	"errors"
	"testing"
	"time"

	"github.com/DLag/cachery"
	"github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("TEST ERROR")

type cacheFetcher struct {
	values map[interface{}]interface{}
	calls  int
}

func (f *cacheFetcher) fetch(key interface{}) (interface{}, error) {
	f.calls++
	if val, ok := f.values[key]; ok {
		return val, nil
	}
	return nil, errTest
}

func TestDriver_Cache1SetAndGet(t *testing.T) {
	c1Fetcher := cacheFetcher{
		values: map[interface{}]interface{}{
			"a": 1,
			"b": 2,
		},
	}

	s := new(cachery.GobSerializer)
	m := new(cachery.Manager)
	d := Default(nats.DefaultURL, "cachery-test")
	t.Run("Init", func(t *testing.T) {
		m.Add(cachery.NewDefault("CACHE1", cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: s,
		},
			d,
			nil,
		),
			cachery.NewDefault("CACHE2", cachery.Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Serializer: s,
			},
				d,
				nil),
		)
	})

	a := assert.New(t)
	c1 := m.Get("CACHE1")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	c1.InvalidateAll()
	time.Sleep(time.Millisecond * 100)
	a.Nil(m.Get("NOCACHE"))

	key := "a"
	t.Run("NoKey", func(t *testing.T) {
		var val int
		wrongKey := "wrong"
		err := c1.Get(wrongKey, &val, c1Fetcher.fetch)
		a.Error(err)
		a.IsType(int(0), val)
		a.Equal(0, val)
		a.Equal(1, c1Fetcher.calls)
	})
	t.Run("NoCache", func(t *testing.T) {
		var val int

		err := c1.Get(key, &val, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(2, c1Fetcher.calls)
	})
	t.Run("Cached", func(t *testing.T) {
		time.Sleep(time.Millisecond * 500)
		var val int

		err := c1.Get(key, &val, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(2, c1Fetcher.calls)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(time.Second)
		var val int

		err := c1.Get("a", &val, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(3, c1Fetcher.calls)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var val int

		err := c1.Get("a", &val, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(4, c1Fetcher.calls)
	})
}

func TestDriver_Cache2SetAndGet(t *testing.T) {
	a := assert.New(t)
	type TestType struct {
		S string
	}

	c2Fetcher := cacheFetcher{
		values: map[interface{}]interface{}{
			"a": TestType{"aa"},
			"b": TestType{"bb"},
		},
	}

	s := new(cachery.GobSerializer)
	m := new(cachery.Manager)
	d := Default(nats.DefaultURL, "cachery-test")
	key := "a"

	t.Run("Init", func(t *testing.T) {
		m.Add(cachery.NewDefault("CACHE1", cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: s,
		},
			d,
			nil,
		),
			cachery.NewDefault("CACHE2", cachery.Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Serializer: s,
			},
				d,
				nil),
		)
	})

	c2 := m.Get("CACHE2")
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	c2.InvalidateAll()
	time.Sleep(time.Millisecond * 100)

	t.Run("NoCache", func(t *testing.T) {
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.values[key], val)
		a.Equal(1, c2Fetcher.calls)
	})
	t.Run("Cached", func(t *testing.T) {
		time.Sleep(time.Millisecond * 500)
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.values[key], val)
		a.Equal(1, c2Fetcher.calls)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.values[key], val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(2, c2Fetcher.calls)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(5 * time.Second)
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.values[key], val)
		a.Equal(3, c2Fetcher.calls)
	})
}

func TestDriver_Invalidate(t *testing.T) {
	a := assert.New(t)
	c1Fetcher := cacheFetcher{
		values: map[interface{}]interface{}{
			"a": 1,
			"b": 2,
		},
	}
	c2Fetcher := cacheFetcher{
		values: map[interface{}]interface{}{
			"a": 11,
			"b": 22,
		},
	}

	s := new(cachery.GobSerializer)
	m := new(cachery.Manager)
	d1 := Default(nats.DefaultURL, "cachery-test")
	d2 := Default(nats.DefaultURL, "cachery-test")
	key := "a"

	t.Run("Init", func(t *testing.T) {
		m.Add(cachery.NewDefault("CACHE1", cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: s,
			Tags:       []string{"tag12", "tag1"},
		},
			d1,
			nil,
		),
			cachery.NewDefault("CACHE2", cachery.Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Serializer: s,
				Tags:       []string{"tag12", "tag2"},
			},
				d2,
				nil),
		)
	})

	c1 := m.Get("CACHE1")
	c2 := m.Get("CACHE2")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	c1.InvalidateAll()
	c2.InvalidateAll()
	time.Sleep(time.Millisecond * 100)

	t.Run("NoCache", func(t *testing.T) {
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(1, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2Fetcher.calls)
	})
	t.Run("InvalidateCache1", func(t *testing.T) {
		c1.Invalidate(key)
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2Fetcher.calls)
	})
	t.Run("InvalidateCache2", func(t *testing.T) {
		c2.Invalidate("a")
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2Fetcher.calls)
	})
	t.Run("InvalidateTag1", func(t *testing.T) {
		c1.InvalidateTags("tag1")
		c2.InvalidateTags("tag1")
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(3, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2Fetcher.calls)
	})
	t.Run("InvalidateTag12", func(t *testing.T) {
		c1.InvalidateTags("tag12")
		c2.InvalidateTags("tag12")
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(4, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2Fetcher.calls)
	})

	t.Run("InvalidateTag1OnManager", func(t *testing.T) {
		m.InvalidateTags("tag1")
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(5, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2Fetcher.calls)
	})

	t.Run("InvalidateAllOnManager", func(t *testing.T) {
		m.InvalidateAll()
		time.Sleep(time.Millisecond * 100)
		var val1, val2 int

		err := c1.Get("a", &val1, c1Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(6, c1Fetcher.calls)

		err = c2.Get("a", &val2, c2Fetcher.fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(4, c2Fetcher.calls)
	})
}
