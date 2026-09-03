package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
)

// Format encodes a value into a byte slice using gob.
func Format(v any) ([]byte, error) {
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
