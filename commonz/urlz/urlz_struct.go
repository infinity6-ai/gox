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
	return ret, err
}

func (u *Url) String() string {
	provider, err := getProvider(u.Scheme)
	errorz.Check(err)
	return provider.ToString(u)
}
