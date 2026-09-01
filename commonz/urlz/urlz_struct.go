package urlz

import (
	"fmt"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
)

// http(s)://aaa....
// file:/tmp/x , file:///tmp/x
// unix:/tmp/socket
// gs://bla/ble

type Url struct {
	Scheme   string
	User     string
	Password string
	Host     string
	Port     string
	Path     *pathz.Path
	Query    string
	Fragment string
}

func Parse(urlStr string) (*Url, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("error parseing url: %s, %w", urlStr, err)
	}
	provider, err := getProvider(u.Scheme)
	if err != nil {
		return nil, err
	}
	ret, err := provider.Parser(u)
	if err != nil {
		return nil, err
	}
	err = provider.Validation(ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (u *Url) String() string {
	provider, err := getProvider(u.Scheme)
	errorz.Check(err)
	err = provider.Validation(u)
	errorz.Check(err)
	return provider.ToString(u)
}

func (u *Url) Clone() *Url {
	return &Url{
		Scheme:   u.Scheme,
		User:     u.User,
		Password: u.Password,
		Host:     u.Host,
		Port:     u.Port,
		Path:     u.Path.Clone(),
		Query:    u.Query,
		Fragment: u.Fragment,
	}
}

func (u *Url) Append(others ...*pathz.Path) (*Url, error) {
	p, err := u.Path.Join(others...)
	if err != nil {
		err = fmt.Errorf("%w: error appending path to url %s: %s", err, u, others)
	}
	var ret *Url
	if p != nil {
		ret = u.Clone()
		ret.Path = p
	}
	return ret, err
}
