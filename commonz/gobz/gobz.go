package gobz

import (
	"bytes"
	"encoding/gob"
	"fmt"

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
