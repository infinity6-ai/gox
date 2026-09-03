package jsonz

import (
	"encoding/json"
	"fmt"
	"io" // Added import

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/constraintz/parserz" // Added import
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func Parse[T any, I blobz.Data](data I) (T, error) {
	var result T
	decoder := json.NewDecoder(blobz.New(data).NewReader())
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode json: %w", err)
	}
	return result, nil
}

func MustParse[T any, I blobz.Data](data I) T {
	res, err := Parse[T](data)
	errorz.Check(err)
	return res
}

func Format(data any) (blobz.Blob, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return blobz.New(b), nil
}

func MustFormat(data any) blobz.Blob {
	res, err := Format(data)
	errorz.Check(err)
	return res
}

// NewReader creates a new ItemReader for JSON-encoded data, compliant with the parserz.ItemReader interface.
func NewReader[T any](r io.Reader) parserz.ItemReader[T] {
	decoder := json.NewDecoder(r)
	decoder.UseNumber() // Maintain consistent decoding behavior with Parse

	return parserz.NewItemReaderWriter(
		func() (T, error) {
			var item T
			if err := decoder.Decode(&item); err != nil {
				if err == io.EOF {
					return item, io.EOF // Return zero value, io.EOF for EOF
				}
				return item, fmt.Errorf("failed to decode json item: %w", err)
			}
			return item, nil
		},
		nil,
	)
}

// NewWriter creates a new ItemWriter for JSON-encoded data, compliant with the parserz.ItemWriter interface.
func NewWriter[T any](w io.Writer) parserz.ItemWriter[T] {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false) // Preserve original behavior if any

	return parserz.NewItemReaderWriter(
		nil,
		func(item T) error {
			if err := encoder.Encode(item); err != nil {
				return fmt.Errorf("failed to encode json item: %w", err)
			}
			return nil
		},
	)
}

// NewReaderWriter creates a new ItemReaderWriter for JSON-encoded data, compliant with the parserz.ItemReaderWriter interface.
func NewReaderWriter[T any](rw io.ReadWriter) parserz.ItemReaderWriter[T] {
	decoder := json.NewDecoder(rw)
	decoder.UseNumber()
	encoder := json.NewEncoder(rw)
	encoder.SetEscapeHTML(false)

	return parserz.NewItemReaderWriter(
		func() (T, error) {
			var item T
			if err := decoder.Decode(&item); err != nil {
				if err == io.EOF {
					return item, io.EOF // Return zero value, io.EOF for EOF
				}
				return item, fmt.Errorf("failed to decode json item: %w", err)
			}
			return item, nil
		},
		func(item T) error {
			if err := encoder.Encode(item); err != nil {
				return fmt.Errorf("failed to encode json item: %w", err)
			}
			return nil
		},
	)
}

// Copy marshals the input to JSON and unmarshals it into the output.
// This is useful for deep copying or type conversion through JSON serialization.
func Copy[I any, O any](input I, output O) error {
	b, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}
	if err := json.Unmarshal(b, output); err != nil {
		return fmt.Errorf("failed to unmarshal into output: %w", err)
	}
	return nil
}

// MustCopy is a convenience function that calls Copy and panics if an error occurs.
func MustCopy[I any, O any](input I, output O) {
	errorz.Check(Copy(input, output))
}

// Clone creates a deep copy of the input object using JSON marshaling and unmarshaling.
// It returns a new instance of the same type as the input.
func Clone[T any](input T) (T, error) {
	var cloned T
	b, err := json.Marshal(input)
	if err != nil {
		return cloned, fmt.Errorf("failed to marshal input for cloning: %w", err)
	}
	if err := json.Unmarshal(b, &cloned); err != nil {
		return cloned, fmt.Errorf("failed to unmarshal into cloned object: %w", err)
	}
	return cloned, nil
}

// MustClone is a convenience function that calls Clone and panics if an error occurs.
func MustClone[T any](input T) T {
	res, err := Clone(input)
	errorz.Check(err)
	return res
}
