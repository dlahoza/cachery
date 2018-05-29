package cachery

import (
	"bytes"
	"encoding/gob"
)

// Serializer describes serializer interface
type Serializer interface {
	// Serialize serializes object to []byte
	Serialize(obj interface{}) ([]byte, error)
	// Deserialize deserializes []byte to object
	Deserialize(src []byte, obj interface{}) error
}

// GobSerializer implements Serializer with Gob serialization
type GobSerializer struct{}

// Serialize serializes object to []byte
func (GobSerializer) Serialize(obj interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	w.Reset()
	e := gob.NewEncoder(w)
	err := e.Encode(obj)
	return w.Bytes(), err
}

// Deserialize deserializes []byte to object
func (GobSerializer) Deserialize(src []byte, obj interface{}) error {
	d := gob.NewDecoder(bytes.NewReader(src))
	return d.Decode(obj)
}
