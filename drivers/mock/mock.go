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

package mock

import (
	"time"

	"github.com/stretchr/testify/mock"
)

//
// This is a mock of the driver. It could be used to test github.com/DLag/cachery behaviour.
//

// Driver type satisfies cachery.Driver interface
type Driver struct {
	mock.Mock
}

// Get loads key from the cache store if it is not outdated
func (m *Driver) Get(cacheName string, key interface{}) ([]byte, time.Duration, error) {
	args := m.Called(cacheName, key)
	return args.Get(0).([]byte), args.Get(1).(time.Duration), args.Error(2)
}

// Set saves key to the cache store
func (m *Driver) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	args := m.Called(cacheName, key, val, ttl)
	return args.Error(0)
}

// Invalidate removes the key from the cache store
func (m *Driver) Invalidate(cacheName string, key interface{}) error {
	return m.Called(cacheName, key).Error(0)
}

// InvalidateAll removes all keys from the cache store
func (m *Driver) InvalidateAll(cacheName string) {
	m.Called(cacheName)
}
