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

package cachery

import (
	"testing"
	"time"

	"errors"
	"sync"

	"github.com/DLag/cachery/drivers/mock"
	"github.com/stretchr/testify/assert"
)

var ErrTest = errors.New("TEST ERROR")

type CacheFetcher struct {
	Values map[interface{}]interface{}
	calls  int
	sync.Mutex
}

func (f *CacheFetcher) Fetch(key interface{}) (interface{}, error) {
	f.Lock()
	defer f.Unlock()
	f.calls++
	if val, ok := f.Values[key]; ok {
		return val, nil
	}
	return nil, ErrTest
}

func (f *CacheFetcher) Calls() int {
	f.Lock()
	defer f.Unlock()
	return f.calls
}

func TestDefaultCache_Cache1SetAndGet(t *testing.T) {
	c1Fetcher := CacheFetcher{
		Values: map[interface{}]interface{}{
			"a": 1,
			"b": 2,
		},
	}

	s := new(GobSerializer)
	m := new(Manager)
	d1 := new(mock.Driver)
	d2 := new(mock.Driver)
	t.Run("Init", func(t *testing.T) {
		m.Add(NewDefault("CACHE1", Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Driver:     d1,
			Serializer: s,
		}),
			NewDefault("CACHE2", Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Driver:     d2,
				Serializer: s,
			}),
		)
	})

	a := assert.New(t)
	c1 := m.Get("CACHE1")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	d1.On("InvalidateAll", c1.Name())
	c1.InvalidateAll()
	a.Nil(m.Get("NOCACHE"))

	key := "a"
	valSerialized, _ := s.Serialize(c1Fetcher.Values[key])
	t.Run("NoKey", func(t *testing.T) {
		var val int
		wrongKey := "wrong"
		d1.On("Get", c1.Name(), wrongKey).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Get", c1.Name(), wrongKey).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		err := c1.Get(wrongKey, &val, c1Fetcher.Fetch)
		d1.AssertExpectations(t)
		a.Error(err)
		a.IsType(int(0), val)
		a.Equal(0, val)
		a.Equal(1, c1Fetcher.Calls())
		d1.AssertExpectations(t)
	})
	t.Run("NoCache", func(t *testing.T) {
		var val int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, valSerialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(valSerialized, time.Second*3, nil).Once()
		err := c1.Get(key, &val, c1Fetcher.Fetch)
		d1.AssertExpectations(t)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(2, c1Fetcher.Calls())
		d1.AssertExpectations(t)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(time.Second)
		var val int

		d1.On("Get", c1.Name(), key).
			Return(valSerialized, time.Second*2, nil).Once()

		err := c1.Get(key, &val, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(2, c1Fetcher.Calls())
		d1.AssertExpectations(t)
	})
	t.Run("BackgroundFetch", func(t *testing.T) {
		time.Sleep(time.Second)
		var val int

		d1.On("Get", c1.Name(), key).
			Return(valSerialized, time.Second*1, nil).Once()
		d1.On("Set", c1.Name(), key, valSerialized, time.Second*3).
			Return(nil).Once()

		err := c1.Get("a", &val, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(3, c1Fetcher.Calls())
		d1.AssertExpectations(t)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		var val int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, valSerialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(valSerialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val)
		a.Equal(1, val)
		a.Equal(4, c1Fetcher.Calls())
		d1.AssertExpectations(t)
	})
}

func TestDefaultCache_Cache2SetAndGet(t *testing.T) {
	a := assert.New(t)
	type TestType struct {
		S string
	}

	c2Fetcher := CacheFetcher{
		Values: map[interface{}]interface{}{
			"a": TestType{"aa"},
			"b": TestType{"bb"},
		},
	}

	s := new(GobSerializer)
	m := new(Manager)
	d1 := new(mock.Driver)
	d2 := new(mock.Driver)
	key := "a"
	valSerialized, _ := s.Serialize(c2Fetcher.Values[key])

	t.Run("Init", func(t *testing.T) {
		m.Add(NewDefault("CACHE1",
			Config{
				Expire:     time.Second * 1,
				Lifetime:   time.Second * 3,
				Driver:     d1,
				Serializer: s,
			}),
			NewDefault("CACHE2", Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Driver:     d2,
				Serializer: s,
			}),
		)
	})

	c2 := m.Get("CACHE2")
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	d2.On("InvalidateAll", c2.Name())
	c2.InvalidateAll()

	t.Run("NoCache", func(t *testing.T) {
		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, valSerialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(valSerialized, time.Second*5, nil).Once()

		var val TestType
		err := c2.Get(key, &val, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.Values[key], val)
		a.Equal(1, c2Fetcher.Calls())
		d2.AssertExpectations(t)
	})
	t.Run("StaleCache", func(t *testing.T) {
		time.Sleep(time.Second)
		d2.On("Get", c2.Name(), key).
			Return(valSerialized, time.Second*4, nil).Once()
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.Values[key], val)
		a.Equal(1, c2Fetcher.Calls())
		d2.AssertExpectations(t)
	})
	t.Run("BackgroundFetch", func(t *testing.T) {
		time.Sleep(3 * time.Second)
		d2.On("Get", c2.Name(), key).
			Return(valSerialized, time.Second*1, nil).Once()
		d2.On("Set", c2.Name(), key, valSerialized, time.Second*5).
			Return(nil).Once()
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.Values[key], val)
		time.Sleep(100 * time.Millisecond)
		a.Equal(2, c2Fetcher.Calls())
		d2.AssertExpectations(t)
	})
	t.Run("Expired", func(t *testing.T) {
		time.Sleep(5 * time.Second)
		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, valSerialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(valSerialized, time.Second*5, nil).Once()
		var val TestType
		err := c2.Get(key, &val, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(TestType{}, val)
		a.Equal(c2Fetcher.Values[key], val)
		a.Equal(3, c2Fetcher.Calls())
		d2.AssertExpectations(t)
	})
}

func TestDefaultCache_Invalidate(t *testing.T) {
	a := assert.New(t)
	c1Fetcher := CacheFetcher{
		Values: map[interface{}]interface{}{
			"a": 1,
			"b": 2,
		},
	}
	c2Fetcher := CacheFetcher{
		Values: map[interface{}]interface{}{
			"a": 11,
			"b": 22,
		},
	}

	s := new(GobSerializer)
	m := new(Manager)
	d1 := new(mock.Driver)
	d2 := new(mock.Driver)
	key := "a"
	val1Serialized, _ := s.Serialize(c1Fetcher.Values[key])
	val2Serialized, _ := s.Serialize(c2Fetcher.Values[key])

	t.Run("Init", func(t *testing.T) {
		m.Add(NewDefault("CACHE1", Config{
			Expire:     time.Second * 1,
			Lifetime:   time.Second * 3,
			Driver:     d1,
			Serializer: s,
			Tags:       []string{"tag12", "tag1"},
		}),
			NewDefault("CACHE2", Config{
				Expire:     time.Second * 3,
				Lifetime:   time.Second * 5,
				Driver:     d2,
				Serializer: s,
				Tags:       []string{"tag12", "tag2"},
			}),
		)
	})

	c1 := m.Get("CACHE1")
	c2 := m.Get("CACHE2")
	a.NotNil(c1)
	a.Equal("CACHE1", c1.Name())
	a.NotNil(c2)
	a.Equal("CACHE2", c2.Name())
	d1.On("InvalidateAll", c1.Name()).Once()
	c1.InvalidateAll()
	d2.On("InvalidateAll", c2.Name()).Once()
	c2.InvalidateAll()

	t.Run("NoCache", func(t *testing.T) {
		var val1, val2 int
		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()
		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(1, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, val2Serialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()

		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})
	t.Run("InvalidateCache1", func(t *testing.T) {
		d1.On("Invalidate", c1.Name(), key).Return(nil).Once()
		c1.Invalidate(key)
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()
		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(1, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})
	t.Run("InvalidateCache2", func(t *testing.T) {
		d2.On("Invalidate", c2.Name(), key).Return(nil).Once()
		c2.Invalidate("a")
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(2, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, val2Serialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()

		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})
	t.Run("InvalidateTag1", func(t *testing.T) {
		d1.On("InvalidateAll", c1.Name()).Once()
		c1.InvalidateTags("tag1")
		c2.InvalidateTags("tag1")
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(3, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()
		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(2, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})
	t.Run("InvalidateTag12", func(t *testing.T) {
		d1.On("InvalidateAll", c1.Name()).Once()
		c1.InvalidateTags("tag12")
		d2.On("InvalidateAll", c2.Name()).Once()
		c2.InvalidateTags("tag12")
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(4, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, val2Serialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()

		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})

	t.Run("InvalidateTag1OnManager", func(t *testing.T) {
		d1.On("InvalidateAll", c1.Name()).Once()
		m.InvalidateTags("tag1")
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(5, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()
		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(3, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})

	t.Run("InvalidateAllOnManager", func(t *testing.T) {
		d1.On("InvalidateAll", c1.Name()).Once()
		d2.On("InvalidateAll", c2.Name()).Once()
		m.InvalidateAll()
		var val1, val2 int

		d1.On("Get", c1.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d1.On("Set", c1.Name(), key, val1Serialized, time.Second*3).
			Return(nil).Once()
		d1.On("Get", c1.Name(), key).
			Return(val1Serialized, time.Second*3, nil).Once()

		err := c1.Get("a", &val1, c1Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val1)
		a.Equal(1, val1)
		a.Equal(6, c1Fetcher.Calls())

		d2.On("Get", c2.Name(), key).
			Return([]byte(nil), time.Duration(0), ErrTest).Once()
		d2.On("Set", c2.Name(), key, val2Serialized, time.Second*5).
			Return(nil).Once()
		d2.On("Get", c2.Name(), key).
			Return(val2Serialized, time.Second*5, nil).Once()
		err = c2.Get("a", &val2, c2Fetcher.Fetch)
		a.NoError(err)
		a.IsType(int(0), val2)
		a.Equal(11, val2)
		a.Equal(4, c2Fetcher.Calls())
		d1.AssertExpectations(t)
		d2.AssertExpectations(t)
	})
}
