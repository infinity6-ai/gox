package urlz

import (
	"errors"
	"fmt"
)

var ErrUnknownSchema = errors.New("unknown schema")

type URI interface {
	Schema() string
	String() string
	validate() error
}

func Schema(url string) (string, string, error) {
	for i := 0; i < len(url); i++ {
		if url[i] == ':' {
			return url[:i], url[i+1:], nil
		}
	}
	return "", "", ErrUnknownSchema
}

func Parse(urlStr string) (URI, error) {
	schema, p, err := Schema(urlStr)
	if err != nil {
		return nil, err
	}
	switch schema {
	case "file", "gs", "unix":
		return NewSimpleUrl(schema, p)
	case "http", "https":
		return NewHttpUrl(schema, p)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownSchema, urlStr)
	}
}
