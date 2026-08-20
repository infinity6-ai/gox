package jsonz

import (
	"encoding/json"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

func Parse[T any, I blobz.Data](data I) (*T, error) {
	var result T
	decoder := json.NewDecoder(blobz.New(data).NewReader())
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	return &result, nil
}

func MustParse[T any, I blobz.Data](data I) *T {
	res, err := Parse[T](data)
	errorz.Check(err)
	return res
}

func Format[T any](data *T) (blobz.Blob, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return blobz.New(b), nil
}

func MustFormat[T any](data *T) blobz.Blob {
	res, err := Format(data)
	errorz.Check(err)
	return res
}
