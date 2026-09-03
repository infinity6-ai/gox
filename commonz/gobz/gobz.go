package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"
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
func Parse(data []byte, v any) error {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to gob decode: %w", err)
	}
	return nil
}
