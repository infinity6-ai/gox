package urlz

import (
	"errors"
	"fmt"
)

var ErrUnknownSchema = errors.New("unknown schema")

type Url interface {
	Schema() string
	String() string
	validate() error
}

func Schema(url string) (string, error) {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' {
			return url[:i], nil
		}
	}
	return "", ErrUnknownSchema
}

func Parse(urlStr string) (Url, error) {
	schema, err := Schema(urlStr)
	if err != nil {
		return nil, err
	}
	switch schema {
	case "file", "gs", "unix":
		panic("implement it")
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownSchema, urlStr)
	}
}
