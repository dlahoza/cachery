package cachery

import "fmt"

// Keyer provides idempotent key
type Keyer interface {
	Key() string
}

// Key returns representation of value which satisfies consistent key requirements
func Key(key interface{}) string {
	switch key.(type) {
	case Keyer:
		return key.(Keyer).Key()
	case string:
		return key.(string)
	case fmt.Stringer:
		return key.(fmt.Stringer).String()
	default:
		return fmt.Sprint(key)
	}
}
