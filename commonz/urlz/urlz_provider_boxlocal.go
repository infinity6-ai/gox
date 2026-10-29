package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation"
)

func (p *providerSpec) providerBoxlocal() (*providerSpec, error) {
	p.Validation = func(u *Url) error {
		err := validation.Equal("boxlocal", u.Scheme, "scheme")
		if err != nil {
			return err
		}
		err = validation.StrNotEmpty(u.Host, "host(box)")
		if err != nil {
			return err
		}
		if u.Path == nil {
			return fmt.Errorf("boxlocal url path cannot be nil")
		}
		err = u.Path.ValidateAbsoluteFile()
		if err != nil {
			return err
		}
		return nil
	}
	p.Parser = func(u *url.URL) (*Url, error) {
		if u.Host == "" {
			return nil, fmt.Errorf("boxlocal url is missing box: %s", u)
		}
		pt, err := pathz.Parse(u.Path)
		if err != nil {
			return nil, fmt.Errorf("error parsing boxlocal url path: %s, %w", u.Path, err)
		}
		return &Url{
			Scheme: u.Scheme,
			Host:   u.Host,
			Path:   pt,
		}, nil
	}
	p.ToString = func(u *Url) string {
		var pathStr string
		if u.Path != nil {
			pathStr = u.Path.String()
		}
		out := &url.URL{
			Scheme: "boxlocal",
			Host:   u.Host,
			Path:   pathStr,
		}
		return out.String()
	}
	return p, nil
}
