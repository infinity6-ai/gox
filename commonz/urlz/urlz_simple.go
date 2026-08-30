package urlz

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

func (u *SimpleUrl) validate() error {
	panic("implement me")
}
