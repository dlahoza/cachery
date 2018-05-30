package cachery

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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

// JSONSerializer implements Serializer with JSON serialization
type JSONSerializer struct{}

// Serialize serializes object to []byte
func (JSONSerializer) Serialize(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// Deserialize deserializes []byte to object
func (JSONSerializer) Deserialize(src []byte, obj interface{}) error {
	return json.Unmarshal(src, obj)
}
