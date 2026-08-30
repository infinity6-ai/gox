package urlz

import (
	"fmt"
	"net/url"
	"strings"
)

type HttpUrl struct {
	schema   string
	host     string
	port     string
	path     string
	query    string
	fragment string
	userInfo *url.Userinfo
}

func NewHttpUrl(schema, host, port, path, query, fragment string, userInfo *url.Userinfo) (*HttpUrl, error) {
	ret := &HttpUrl{
		schema:   schema,
		host:     host,
		port:     port,
		path:     path,
		query:    query,
		fragment: fragment,
		userInfo: userInfo,
	}
	if err := ret.validate(); err != nil {
		return nil, err
	}
	return ret, nil
}

func newHttpUrlFromParts(schema string, urlPart string) (*HttpUrl, error) {
	// Prepend // to the urlPart if it's missing, so url.Parse can handle it correctly.
	// e.g. for "google.com/search", it becomes "//google.com/search"
	if !strings.HasPrefix(urlPart, "//") {
		urlPart = "//" + urlPart
	}

	fullURL := schema + ":" + urlPart
	return ParseHttpUrl(fullURL)
}

func ParseHttpUrl(urlStr string) (*HttpUrl, error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse http url: %w", err)
	}

	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid schema for HttpUrl: %s", parsed.Scheme)
	}

	if parsed.Scheme == "" {
		// try parsing with a default schema
		parsed, err = url.Parse("https://" + urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse http url: %w", err)
		}
	}

	ret := &HttpUrl{
		schema:   parsed.Scheme,
		host:     parsed.Hostname(),
		port:     parsed.Port(),
		path:     parsed.Path,
		query:    parsed.RawQuery,
		fragment: parsed.Fragment,
		userInfo: parsed.User,
	}
	if err := ret.validate(); err != nil {
		return nil, err
	}
	return ret, nil
}

func (u *HttpUrl) validate() error {
	if u.schema != "http" && u.schema != "https" {
		return fmt.Errorf("invalid schema for HttpUrl: %s", u.schema)
	}
	if u.host == "" {
		return fmt.Errorf("host cannot be empty for HttpUrl")
	}
	return nil
}

func (u *HttpUrl) Schema() string {
	return u.schema
}

func (u *HttpUrl) String() string {
	var uRI url.URL
	uRI.Scheme = u.schema
	uRI.Host = u.host
	if u.port != "" {
		uRI.Host += ":" + u.port
	}
	uRI.Path = u.path
	uRI.RawQuery = u.query
	uRI.Fragment = u.fragment
	uRI.User = u.userInfo
	return uRI.String()
}
