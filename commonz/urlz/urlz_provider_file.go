package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

func (p *providerSpec) providerFile() (*providerSpec, error) {
	p.Parser = func(u *url.URL) (*Url, error) {
		pathStr := u.Path
		if u.Opaque != "" {
			pathStr = u.Opaque
		}

		pt, err := pathz.Parse(pathStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing file url path: %q, %w", pathStr, err)
		}
		return &Url{Scheme: u.Scheme, Path: pt}, nil
	}
	p.ToString = func(u *Url) string {
		out := &url.URL{
			Scheme: "file",
			Path:   u.Path.String(),
		}
		return out.String()
	}
	return p, nil
}
