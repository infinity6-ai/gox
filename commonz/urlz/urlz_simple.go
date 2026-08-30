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
	return u.path
}

func NewSimpleUrl(schema string, path string) (*SimpleUrl, error) {
	ret := &SimpleUrl{
		schema: schema,
		path:   path,
	}
	ret.validate()
	return ret, nil
}

func ParseSimpleUrl(urlStr string) (*SimpleUrl, error) {
	schema, _, err := Schema(urlStr)
	if err != nil {
		return nil, err
	}
	if schema != "file" && schema != "gs" && schema != "unix" {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSchema, urlStr)
	}
	panic("imeplement it")
}

func (u *SimpleUrl) validate() error {
	panic("implement me")
}
