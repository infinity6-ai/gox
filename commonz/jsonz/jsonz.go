package jsonz

import (
	"encoding/json"
	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
)

func Parse[T any, I blobz.Data](data I) (*T, error) {
	var result T
	if err := json.Unmarshal(blobz.ToBytes(data), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func MustParse[T any, I blobz.Data](data I) *T {
	res, err := Parse[T](data)
	if err != nil {
		panic(err)
	}
	return res
}

func Format[T any](data *T) (blobz.Blob, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return blobz.New(b), nil
}

func MustFormat[T any](data *T) blobz.Blob {
	res, err := Format(data)
	if err != nil {
		panic(err)
	}
	return res
}
