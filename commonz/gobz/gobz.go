package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/constraintz/parserz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func Parse[T any](data []byte) (*T, error) {
	var result T
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode gob: %w", err)
	}
	return &result, nil
}

func MustParse[T any](data []byte) *T {
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

type gobReader[T any] struct {
	decoder *gob.Decoder
}

// NewReader creates a new ItemReader for gob-encoded data, compliant with the parserz.ItemReader interface.
func NewReader[T any](r io.Reader) parserz.ItemReader[T] {
	return &gobReader[T]{
		decoder: gob.NewDecoder(r),
	}
}

// ReadItem decodes one gob-encoded item from the reader.
// It returns the item and any decoding error. If io.EOF is encountered, it returns nil, nil.
func (r *gobReader[T]) ReadItem() (*T, error) {
	var item T
	if err := r.decoder.Decode(&item); err != nil {
		if err == io.EOF {
			return nil, nil // Return nil, nil for EOF as requested
		}
		return nil, err // For other errors, return nil and the error
	}
	return &item, nil
}

type gobWriter[T any] struct {
	encoder *gob.Encoder
}

// NewWriter creates a new ItemWriter for gob-encoded data, compliant with the parserz.ItemWriter interface.
func NewWriter[T any](w io.Writer) parserz.ItemWriter[T] {
	return &gobWriter[T]{
		encoder: gob.NewEncoder(w),
	}
}

// WriteItem encodes one item to the writer as gob-encoded data.
func (w *gobWriter[T]) WriteItem(item *T) error {
	return w.encoder.Encode(item)
}
