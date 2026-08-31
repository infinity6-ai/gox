package urlz

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

var ErrUnknownScheme = errors.New("unknown scheme")

type providerSpec struct {
	Parser   func(u *url.URL) (*Url, error)
	ToString func(u *Url) string
}

func getProvider(scheme string) (*providerSpec, error) {
	prv := &providerSpec{}
	switch scheme {
	case "http", "https":
		panic("implement")
	case "file":
		panic("implement")
	case "gs":
		panic("implement")
	case "unix":
		return prv.providerUnix()
	default:
		return nil, ErrUnknownScheme
	}
}

func (p *providerSpec) providerUnix() (*providerSpec, error) {
	p.Parser = func(u *url.URL) (*Url, error) {
		pt, err := pathz.Parse(u.Path)
		if err != nil {
			return nil, fmt.Errorf("error parsing unix url: %s, %w", u, err)
		}
		return &Url{Scheme: u.Scheme, Path: pt}, nil
	}
	p.ToString = func(u *Url) string {
		return u.Scheme + ":/" + u.Path.String()
	}
	return p, nil
}
