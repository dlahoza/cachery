package redis

import (
	"errors"
	"testing"
	"time"

	"github.com/DLag/cachery"
	"github.com/stretchr/testify/assert"
)

var testErr = errors.New("test_error")

func TestRedisCache_Cache1SetAndGet(t *testing.T) {
	c1FetchCount := 0
	c1Fetcher := func(key interface{}) (interface{}, error) {
		c1FetchCount++
		switch cachery.Key(key) {
		case "a":
			return 1, nil
		case "b":
			return 2, nil
		}
		return 0, testErr
	}

	m := new(cachery.Manager)
	t.Run("Init", func(t *testing.T) {
		m.Add(New("CACHE1", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: new(cachery.GobSerializer),
		}))
		m.Add(New("CACHE2", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 3,
			Lifetime:   time.Second * 5,
			Serializer: new(cachery.GobSerializer),
		}))
	})

	a := assert.New(t)
	c1 := m.Get("CACHE1")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	c1.InvalidateAll()
	a.Nil(m.Get("NOCACHE"))

	t.Run("NoCache", func(t *testing.T) {
		var val int
		err := c1.Get("a", &val, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(1, c1FetchCount)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(time.Second)
		var val int
		err := c1.Get("a", &val, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(1, c1FetchCount)
	})
	t.Run("BackgroundFetch", func(t *testing.T) {
		time.Sleep(time.Second)
		var val int
		err := c1.Get("a", &val, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(2, c1FetchCount)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var val int
		err := c1.Get("a", &val, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(3, c1FetchCount)
	})
}

func TestRedisCache_Cache2SetAndGet(t *testing.T) {
	a := assert.New(t)
	type TestType struct {
		S string
	}
	c2FetchCount := 0
	c2Fetcher := func(key interface{}) (interface{}, error) {
		c2FetchCount++
		switch cachery.Key(key) {
		case "a":
			return TestType{"aa"}, nil
		case "b":
			return TestType{"bb"}, nil
		}
		return TestType{}, testErr
	}

	m := new(cachery.Manager)
	t.Run("Init", func(t *testing.T) {
		m.Add(New("CACHE1", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: new(cachery.GobSerializer),
		}))
		m.Add(New("CACHE2", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 3,
			Lifetime:   time.Second * 5,
			Serializer: new(cachery.GobSerializer),
		}))
	})

	c2 := m.Get("CACHE2")
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	c2.InvalidateAll()

	t.Run("NoCache", func(t *testing.T) {
		var val TestType
		err := c2.Get("a", &val, c2Fetcher)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(TestType{"aa"}, val)
		a.Equal(1, c2FetchCount)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(time.Second)
		var val TestType
		err := c2.Get("a", &val, c2Fetcher)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(TestType{"aa"}, val)
		a.Equal(1, c2FetchCount)
	})
	t.Run("BackgroundFetch", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var val TestType
		err := c2.Get("a", &val, c2Fetcher)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(TestType{"aa"}, val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(2, c2FetchCount)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(5 * time.Second)
		var val TestType
		err := c2.Get("a", &val, c2Fetcher)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(TestType{"aa"}, val)
		a.Equal(3, c2FetchCount)
	})
}

func TestRedisCache_Invalidate(t *testing.T) {
	a := assert.New(t)
	c1FetchCount := 0
	c2FetchCount := 0
	c1Fetcher := func(key interface{}) (interface{}, error) {
		c1FetchCount++
		switch cachery.Key(key) {
		case "a":
			return 1, nil
		case "b":
			return 2, nil
		}
		return 0, testErr
	}
	c2Fetcher := func(key interface{}) (interface{}, error) {
		c2FetchCount++
		switch cachery.Key(key) {
		case "a":
			return 11, nil
		case "b":
			return 22, nil
		}
		return 0, testErr
	}

	m := new(cachery.Manager)
	t.Run("Init", func(t *testing.T) {
		m.Add(New("CACHE1", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Serializer: new(cachery.GobSerializer),
			Tags:       []string{"tag12", "tag1"},
		}))
		m.Add(New("CACHE2", DefaultPool("127.0.0.1:6379", 3, 120), cachery.Config{
			Expire:     time.Second * 3,
			Lifetime:   time.Second * 5,
			Serializer: new(cachery.GobSerializer),
			Tags:       []string{"tag12", "tag2"},
		}))
	})

	c1 := m.Get("CACHE1")
	c2 := m.Get("CACHE2")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	c1.InvalidateAll()
	c2.InvalidateAll()

	c1.InvalidateAll()
	c2.InvalidateAll()

	t.Run("NoCache", func(t *testing.T) {
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(1, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2FetchCount)
	})
	t.Run("InvalidateCache1", func(t *testing.T) {
		c1.Invalidate("a")
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2FetchCount)
	})
	t.Run("InvalidateCache2", func(t *testing.T) {
		c2.Invalidate("a")
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2FetchCount)
	})
	t.Run("InvalidateTag1", func(t *testing.T) {
		c1.InvalidateTags("tag1")
		c2.InvalidateTags("tag1")
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(3, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2FetchCount)
	})
	t.Run("InvalidateTag12", func(t *testing.T) {
		c1.InvalidateTags("tag12")
		c2.InvalidateTags("tag12")
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(4, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2FetchCount)
	})
	t.Run("InvalidateTag1OnManager", func(t *testing.T) {
		m.InvalidateTags("tag1")
		var val1, val2 int
		err := c1.Get("a", &val1, c1Fetcher)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(5, c1FetchCount)

		err = c2.Get("a", &val2, c2Fetcher)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2FetchCount)
	})
}
