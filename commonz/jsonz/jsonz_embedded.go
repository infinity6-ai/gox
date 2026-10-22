package jsonz

import (
	"encoding/json"
	"fmt"
)

// Embedded wraps a value of type T, marshaling it as an escaped JSON string
// and unmarshaling it from an escaped JSON string.
type Embedded[T any] struct {
	Ref T
}

func (p Embedded[T]) Jsonify() string {
	return MustFormat(p).String()
}

// MarshalJSON implements json.Marshaler. It marshals the inner value as JSON,
// and then serializes that result as an escaped JSON string.
func (p Embedded[T]) MarshalJSON() ([]byte, error) {
	r, err := json.Marshal(p.Ref)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Embedded[%T]: %w", p.Ref, err)
	}
	r, err = json.Marshal(string(r))
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Embedded[%T] as string: %w", p.Ref, err)
	}
	return r, nil
}

// UnmarshalJSON implements json.Unmarshaler. It unmarshals the JSON string
// and then parses that string into the inner value.
func (p *Embedded[T]) UnmarshalJSON(data []byte) error {
	if p == nil {
		return fmt.Errorf("cannot unmarshal into nil *Embedded")
	}
	if string(data) == "null" {
		var zero T
		p.Ref = zero
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal Embedded[%T] string: %w", p.Ref, err)
	}
	if len(s) == 0 {
		var zero T
		p.Ref = zero
		return nil
	}
	if err := json.Unmarshal([]byte(s), &p.Ref); err != nil {
		return fmt.Errorf("cannot unmarshal inner Embedded[%T]: %w", p.Ref, err)
	}
	return nil
}
