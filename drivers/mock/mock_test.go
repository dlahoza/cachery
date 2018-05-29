package mock

import (
	"testing"
	"time"
)

func TestDriver(t *testing.T) {
	m := new(Driver)
	m.On("Get", "a", "b").Return([]byte(nil), time.Second, nil)
	m.Get("a", "b")
	m.On("Set", "a", "b", []byte(nil), time.Second).Return(nil)
	m.Set("a", "b", []byte(nil), time.Second)
	m.On("Invalidate", "a", "b").Return(nil)
	m.Invalidate("a", "b")
	m.On("InvalidateAll", "a")
	m.InvalidateAll("a")
}
