package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"reflect"
)

// Format encodes a value into a byte slice using gob.
func Format(v any) ([]byte, error) {
	if v != nil {
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr && val.IsNil() {
			v = nil // Convert typed nil to untyped nil to prevent panic
		}
	}
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("failed to gob encode: %w", err)
	}
	return buf.Bytes(), nil
}

// Parse decodes a gob-encoded byte slice into a value.
// The value 'v' must be a pointer to the target data structure.
func Parse[T any](data []byte, v T) (T, error) {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(v); err != nil {
		return v, fmt.Errorf("failed to gob decode: %w", err)
	}
	return v, nil
}

// ParseReader decodes a gob-encoded stream from an io.Reader into a value.
// The value 'v' must be a pointer to the target data structure.
func ParseReader[T any](r io.Reader, v T) (T, error) {
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(v); err != nil {
		return v, fmt.Errorf("failed to gob decode: %w", err)
	}
	return v, nil
}

// FormatWriter encodes a value into an io.Writer using gob.
func FormatWriter(w io.Writer, v any) error {
	encoder := gob.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("failed to gob encode: %w", err)
	}
	return nil
}

// Copy encodes the input to gob and decodes it into the output.
// This is useful for deep copying.
// It returns the populated output object.
func Copy[I any, O any](input I, output O) (O, error) {
	data, err := Format(input)
	if err != nil {
		return output, err
	}
	return Parse(data, output)
}

// Clone creates a deep copy of the input object using gob encoding and decoding.
// It returns a new instance of the same type as the input.
func Clone[T any](input T, output T) (T, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		var zero T
		return zero, nil
	}
	return Copy(input, output)
}
