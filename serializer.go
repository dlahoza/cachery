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
