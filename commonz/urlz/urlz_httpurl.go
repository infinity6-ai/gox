package urlz

type HttpUrl struct {
	schema   string
	host     string
	port     string
	path     string
	query    string
	fragment string
}

func NewHttpUrl(schema string, path string) (*SimpleUrl, error) {
	ret := &SimpleUrl{
		schema: schema,
		path:   path,
	}
	ret.validate()
	return ret, nil
}

func (u *HttpUrl) validate() error {
	panic("implement me")
}
