package urlz

import (
	"errors"
	"net/url"
)

var ErrUnknownScheme = errors.New("unknown scheme")

type providerSpec struct {
	Parser   func(u *url.URL) (*Url, error)
	ToString func(u *Url) string
}

func getProvider(scheme string) (*providerSpec, error) {
	switch scheme {
	case "http", "https":
		panic("implement")
	case "file":
		panic("implement")
	case "gs":
		panic("implement")
	case "unix":
		panic("implement")
	default:
		return nil, ErrUnknownScheme
	}
}
