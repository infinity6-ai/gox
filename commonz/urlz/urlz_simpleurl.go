package urlz

import "fmt"

type SimpleUrl struct {
	schema string
	path   string
}

func (u *SimpleUrl) Schema() string {
	return u.schema
}

func (u *SimpleUrl) String() string {
	return u.schema + ":" + u.path
}

func NewSimpleUrl(schema string, path string) (*SimpleUrl, error) {
	ret := &SimpleUrl{
		schema: schema,
		path:   path,
	}
	if err := ret.validate(); err != nil {
		return nil, err
	}
	return ret, nil
}

func ParseSimpleUrl(urlStr string) (*SimpleUrl, error) {
	schema, path, err := Schema(urlStr)
	if err != nil {
		return nil, err
	}
	return NewSimpleUrl(schema, path)
}

func (u *SimpleUrl) validate() error {
	switch u.schema {
	case "file", "gs", "unix":
		return nil
	default:
		return fmt.Errorf("%w: invalid schema for SimpleUrl: %s", ErrUnknownSchema, u.schema)
	}
}
