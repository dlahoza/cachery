package mock

import (
	"time"

	"github.com/stretchr/testify/mock"
)

//
// This is a mock of the driver. It could be used to test github.com/DLag/cachery behaviour.
//

type Driver struct {
	mock.Mock
}

func (m *Driver) Get(cacheName string, key interface{}) ([]byte, time.Duration, error) {
	args := m.Called(cacheName, key)
	return args.Get(0).([]byte), args.Get(1).(time.Duration), args.Error(2)
}

func (m *Driver) Set(cacheName string, key interface{}, val []byte, ttl time.Duration) (err error) {
	args := m.Called(cacheName, key, val, ttl)
	return args.Error(0)
}

func (m *Driver) Invalidate(cacheName string, key interface{}) error {
	return m.Called(cacheName, key).Error(0)
}

func (m *Driver) InvalidateAll(cacheName string) {
	m.Called(cacheName)
}
