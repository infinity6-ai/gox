package jsonz

import (
	"encoding/json"
	"fmt"
)

// EmbeddedJson wraps a value of type T, marshaling it as an escaped JSON string
// and unmarshaling it from an escaped JSON string.
type EmbeddedJson[T any] struct {
	ref T
}

// Json is an alias for EmbeddedJson.
type Json[T any] = EmbeddedJson[T]

// NewEmbedded creates a new EmbeddedJson[T] wrapper around value.
func NewEmbedded[T any](value T) EmbeddedJson[T] {
	return EmbeddedJson[T]{ref: value}
}

// NewEmbeddedJson creates a new EmbeddedJson[T] wrapper around value.
func NewEmbeddedJson[T any](value T) EmbeddedJson[T] {
	return EmbeddedJson[T]{ref: value}
}

// Get returns the underlying value.
func (p EmbeddedJson[T]) Get() T {
	return p.ref
}

// Set updates the underlying value.
func (p *EmbeddedJson[T]) Set(value T) {
	p.ref = value
}

// Value returns a pointer to the underlying value.
func (p *EmbeddedJson[T]) Value() *T {
	return &p.ref
}

// MarshalJSON implements json.Marshaler. It marshals the inner value as JSON,
// and then serializes that result as an escaped JSON string.
func (p EmbeddedJson[T]) MarshalJSON() ([]byte, error) {
	r, err := json.Marshal(p.ref)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal EmbeddedJson[%T]: %w", p.ref, err)
	}
	r, err = json.Marshal(string(r))
	if err != nil {
		return nil, fmt.Errorf("cannot marshal EmbeddedJson[%T] as string: %w", p.ref, err)
	}
	return r, nil
}

// UnmarshalJSON implements json.Unmarshaler. It unmarshals the JSON string
// and then parses that string into the inner value.
func (p *EmbeddedJson[T]) UnmarshalJSON(data []byte) error {
	if p == nil {
		return fmt.Errorf("cannot unmarshal into nil *EmbeddedJson")
	}
	if string(data) == "null" {
		var zero T
		p.ref = zero
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal EmbeddedJson[%T] string: %w", p.ref, err)
	}
	if len(s) == 0 {
		var zero T
		p.ref = zero
		return nil
	}
	if err := json.Unmarshal([]byte(s), &p.ref); err != nil {
		return fmt.Errorf("cannot unmarshal inner EmbeddedJson[%T]: %w", p.ref, err)
	}
	return nil
}
