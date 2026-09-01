package httpzclient

import (
	"io"
	"net/http"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

type Req struct {
	Method  string
	Path    *pathz.Path
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
