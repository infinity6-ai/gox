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

func (p *providerSpec) providerHttp() (*providerSpec, error) {
	p.Parser = func(u *url.URL) (*Url, error) {
		pt, err := pathz.Parse(u.Path)
		if err != nil {
			return nil, fmt.Errorf("error parsing http url path: %s, %w", u.Path, err)
		}
		password, _ := u.User.Password()
		return &Url{
			Scheme:   u.Scheme,
			User:     u.User.Username(),
			Password: password,
			Host:     u.Hostname(),
			Port:     u.Port(),
			Path:     pt,
			Query:    u.RawQuery,
			Fragment: u.Fragment,
		}, nil
	}
	p.ToString = func(u *Url) string {
		var user *url.Userinfo
		if u.User != "" {
			if u.Password != "" {
				user = url.UserPassword(u.User, u.Password)
			} else {
				user = url.User(u.User)
			}
		}

		host := u.Host
		if u.Port != "" {
			host = host + ":" + u.Port
		}

		ret := &url.URL{
			Scheme:   u.Scheme,
			User:     user,
			Host:     host,
			Path:     u.Path.String(),
			RawQuery: u.Query,
			Fragment: u.Fragment,
		}
		return ret.String()
	}
	return p, nil
}

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

func (p *providerSpec) providerGs() (*providerSpec, error) {
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

func (p *providerSpec) providerUnix() (*providerSpec, error) {
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
