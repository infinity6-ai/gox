package urlz

import (
	"errors"
	"net/url"
)

var ErrUnknownScheme = errors.New("unknown scheme")

type providerSpec struct {
	Parser     func(u *url.URL) (*Url, error)
	ToString   func(u *Url) string
	Validation func(u *Url) error
}

func getProvider(scheme string) (*providerSpec, error) {
	prv := &providerSpec{}
	switch scheme {
	case "http", "https":
		return prv.providerHttp()
	case "file":
		return prv.providerFile()
	case "gs":
		return prv.providerGs()
	case "unix":
		return prv.providerUnix()
	default:
		return nil, ErrUnknownScheme
	}
}
