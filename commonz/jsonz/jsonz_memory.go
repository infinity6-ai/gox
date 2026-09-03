package jsonz

import (
	"encoding/json"
	"fmt"
	"io"

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

func Format(data any) (blobz.Blob, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return blobz.New(b), nil
}
