package cachery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct{}

func (test) Key() string {
	return "testKeyStruct"
}

type testStringer struct{}

func (testStringer) String() string {
	return "testStringer"
}

func TestKey(t *testing.T) {
	a := assert.New(t)
	a.Equal("testKeyStruct", Key(test{}))
	a.Equal("testStringer", Key(testStringer{}))
	a.Equal("test", Key("test"))
	a.Equal("123", Key(123))
}
