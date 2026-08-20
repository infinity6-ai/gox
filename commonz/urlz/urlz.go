package urlz

import "net/url"

type Url struct {
	scheme   string
	host     string
	port     int
	path     string
	query    string
	fragment string

	parsedQuery    url.Values
	parsedFragment url.Values
}

func (u *Url) Parse(urlStr string) {
	// consider that schemes: http, https, file, gs, unix and panic if it is not one of them
	panic("implement it")
}

func (u *Url) String() {
	panic("implement it")
}
