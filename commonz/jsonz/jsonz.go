package jsonz

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func ParseReader[T any](r io.Reader, v T) (T, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	err := decoder.Decode(v)
	if err != nil {
		return v, fmt.Errorf("failed to parse json: %w", err)
	}
	return v, nil
}

func Parse[T any, I blobz.Data](data I, v T) (T, error) {
	return ParseReader(blobz.New(data).NewReader(), v)
}

func FormatWriter(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	return nil
}

func Format(v any) (blobz.Blob, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return blobz.New(b), nil
}

// Copy marshals the input to JSON and unmarshals it into the output.
// This is useful for deep copying or type conversion through JSON serialization.
// It returns the populated output object.
func Copy[I any, O any](input I, output O) (O, error) {
	data, err := Format(input)
	if err != nil {
		return output, err
	}
	return Parse(data.Bytes(), output)
}

// Clone creates a deep copy of the input object using JSON marshaling and unmarshaling.
// It returns a new instance of the same type as the input.
func Clone[T any](input T, output T) (T, error) {
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		var zero T
		return zero, nil
	}
	return Copy(input, output)
}
