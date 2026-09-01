package httpzclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
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

func NewReq(method string, path string) *Req {
	return &Req{
		Method:  method,
		Path:    pathz.MustParse(path),
		Query:   url.Values{},
		Headers: http.Header{},
	}
}

func (r *Req) AddQuery(key, value string) *Req {
	r.Query.Add(key, value)
	return r
}

func (r *Req) AddHeader(key, value string) *Req {
	r.Headers.Add(key, value)
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
