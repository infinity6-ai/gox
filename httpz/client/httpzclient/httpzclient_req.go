package httpzclient

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

type Req struct {
	Method  string
	Path    *pathz.Path
	Url     *urlz.Url
	Query   url.Values
	Headers http.Header
	Body    io.Reader
}

func FormatReq(method string, path string, params map[string]string) (*Req, error) {
	p := pathz.MustParse(path)
	if params != nil {
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

func MustFormatReq(method string, path string, params map[string]string) *Req {
	req, err := FormatReq(method, path, params)
	errorz.Check(err)
	return req
}

func NewReq(method string, path string) *Req {
	return &Req{
		Method:  method,
		Path:    pathz.MustParse(path),
		Query:   url.Values{},
		Headers: http.Header{},
	}
}

func (r *Req) SetQuery(key, value string) *Req {
	r.Query.Set(key, value)
	return r
}

func (r *Req) AddQuery(key, value string) *Req {
	r.Query.Add(key, value)
	return r
}

func (r *Req) AddHeader(key, value string) *Req {
	r.Headers.Add(key, value)
	return r
}

func (r *Req) SetHeader(key, value string) *Req {
	r.Headers.Set(key, value)
	return r
}

func (r *Req) SetBody(body io.Reader) *Req {
	r.Body = body
	return r
}

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
