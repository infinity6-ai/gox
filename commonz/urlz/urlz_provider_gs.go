package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *providerSpec) providerGs() (*providerSpec, error) {
	p.Validation = func(u *Url) error {
		return nil
	}
	p.Parser = func(u *url.URL) (*Url, error) {
		if u.Host == "" {
			return nil, fmt.Errorf("gs url is missing bucket: %s", u)
		}
		pt, err := pathz.Parse(u.Path)
		if err != nil {
			return nil, fmt.Errorf("error parsing gs url path: %s, %w", u.Path, err)
		}
		return &Url{
			Scheme: u.Scheme,
			Host:   u.Host, // bucket
			Path:   pt,
		}, nil
	}
	p.ToString = func(u *Url) string {
		out := &url.URL{
			Scheme: "gs",
			Host:   u.Host,
			Path:   u.Path.String(),
		}
		return out.String()
	}
	return p, nil
}
