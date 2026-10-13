// Package httpzrequest provides data structures and builder methods for constructing
// and configuring HTTP requests with paths, URLs, query parameters, headers, and bodies.
package httpzrequest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/pathz/patternpathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
)

// Req represents an HTTP request definition with structured path or URL, query parameters, headers, and body.
type Req struct {
	// Method is the HTTP method (e.g. GET, POST).
	Method string
	// Path is the parsed relative or absolute request path.
	Path *pathz.Path
	// Url is the fully qualified target URL, mutually exclusive with Path during resolution.
	Url *urlz.Url
	// Query holds URL query parameters.
	Query url.Values
	// Headers holds HTTP request headers.
	Headers http.Header
	// Body is the request payload stream.
	Body io.Reader
}

// Format creates a new Req by parsing path and substituting named pattern parameters.
// If params is not empty, path is parsed as a pattern and formatted with the given parameters.
func Format(method string, path string, params map[string]string) (*Req, error) {
	p := pathz.MustParse(path)
	if len(params) > 0 {
		pattern, err := patternpathz.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		p, err = pattern.Format(params)
		if err != nil {
			return nil, fmt.Errorf("failed to format path: %w", err)
		}
	}
	return &Req{
		Method:  method,
		Path:    p,
		Query:   url.Values{},
		Headers: http.Header{},
	}, nil
}

// MustFormat creates a new Req with pattern parameters formatted, panicking if formatting fails.
func MustFormat(method string, path string, params map[string]string) *Req {
	req, err := Format(method, path, params)
	errorz.Check(err)
	return req
}

// New creates a new Req initialized with the specified HTTP method and path string.
func New(method string, path string) *Req {
	return &Req{
		Method:  method,
		Path:    pathz.MustParse(path),
		Query:   url.Values{},
		Headers: http.Header{},
	}
}

// FromUrl creates a new Req initialized with the specified HTTP method and absolute URL.
func FromUrl(method string, u *urlz.Url) *Req {
	return &Req{
		Method:  method,
		Url:     u,
		Query:   url.Values{},
		Headers: http.Header{},
	}
}

// SetQuery sets the query parameter key to value, overwriting any previous values for key.
func (r *Req) SetQuery(key, value string) *Req {
	r.Query.Set(key, value)
	return r
}

// AddQuery appends a value to the query parameter key.
func (r *Req) AddQuery(key, value string) *Req {
	r.Query.Add(key, value)
	return r
}

// AddHeader appends a header key-value pair to the request.
func (r *Req) AddHeader(key, value string) *Req {
	r.Headers.Add(key, value)
	return r
}

// SetHeader sets the header key to value, overwriting any previous values for key.
func (r *Req) SetHeader(key, value string) *Req {
	r.Headers.Set(key, value)
	return r
}

// SetBody sets the request body reader.
func (r *Req) SetBody(body io.Reader) *Req {
	r.Body = body
	return r
}

// ResolveUrl resolves the effective request URL against an optional baseUrl.
// It panics if both Path and Url are specified or if neither is specified.
// If Url is set, it validates that baseUrl is a base of Url (if baseUrl is non-nil).
// If Path is set, baseUrl must not be nil and Path is joined onto baseUrl.
func (r *Req) ResolveUrl(baseUrl *urlz.Url) (*urlz.Url, error) {
	if r.Path == nil && r.Url == nil {
		panic("use either Path or Url")
	}
	if r.Path != nil && r.Url != nil {
		panic("do not use both Path and Url")
	}
	if r.Url != nil {
		if baseUrl != nil && !baseUrl.IsBaseOf(r.Url) {
			return nil, fmt.Errorf("base url mismatch: %s != %s", baseUrl, r.Url)
		}
		return r.Url, nil
	}
	if baseUrl == nil {
		return nil, fmt.Errorf("base url not found: %s", r.Path)
	}
	return baseUrl.JoinPath(r.Path)
}
