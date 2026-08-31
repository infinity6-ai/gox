package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation"
)

func (p *providerSpec) providerUnix() (*providerSpec, error) {
	p.Validation = func(u *Url) error {
		validation.Equal("unix", u.Scheme, "scheme")
		u.Path.Validate(pathz.ValidateOptions{Absolute: new(true), Empty: new(false)})
		return nil
	}
	p.Parser = func(u *url.URL) (*Url, error) {
		// For unix sockets, the path can be in Opaque or Path
		pathStr := u.Path
		if pathStr == "" {
			pathStr = u.Opaque
		}
		pt, err := pathz.Parse(pathStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing unix url: %s, %w", u, err)
		}
		return &Url{Scheme: u.Scheme, Path: pt}, nil
	}
	p.ToString = func(u *Url) string {
		return u.Scheme + ":" + u.Path.String()
	}
	return p, nil
}
