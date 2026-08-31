package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

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
