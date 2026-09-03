package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/parserz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func Parse[T any](data []byte) (T, error) {
	var result T
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode gob: %w", err)
	}
	return result, nil
}

func MustParse[T any](data []byte) T {
	res, err := Parse[T](data)
	errorz.Check(err)
	return res
}

func Format[T any](data *T) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}
	return buf.Bytes(), nil
}

func MustFormat[T any](data *T) []byte {
	res, err := Format(data)
	errorz.Check(err)
	return res
}

// NewReader creates a new ItemReader for gob-encoded data, compliant with the parserz.ItemReader interface.
func NewReader[T any](r io.Reader) parserz.ItemReader[T] {
	decoder := gob.NewDecoder(r)
	return parserz.NewItemReaderWriter[T](
		func() (T, error) {
			var item T
			if err := decoder.Decode(&item); err != nil {
				return item, err // Return zero value and the error
			}
			return item, nil
		},
		nil,
	)
}

// NewWriter creates a new ItemWriter for gob-encoded data, compliant with the parserz.ItemWriter interface.
func NewWriter[T any](w io.Writer) parserz.ItemWriter[T] {
	encoder := gob.NewEncoder(w)
	return parserz.NewItemReaderWriter[T](
		nil,
		func(item T) error {
			return encoder.Encode(item)
		},
	)
}

// NewReaderWriter creates a new ItemReaderWriter for gob-encoded data, compliant with the parserz.ItemReaderWriter interface.
func NewReaderWriter[T any](rw io.ReadWriter) parserz.ItemReaderWriter[T] {
	decoder := gob.NewDecoder(rw)
	encoder := gob.NewEncoder(rw)

	return parserz.NewItemReaderWriter(
		func() (T, error) {
			var item T
			if err := decoder.Decode(&item); err != nil {
				return item, err // Return zero value and the error
			}
			return item, nil
		},
		func(item T) error {
			return encoder.Encode(item)
		},
	)
}
