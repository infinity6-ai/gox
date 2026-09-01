package httpzserverv2

import (
	"io"
	"net/http"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

type Req struct {
	Path    *pathz.Path
	Query   url.Values
	Headers http.Header
	Body    io.Reader
}

func (r *Req) fromHttpRequest(req *http.Request) {
	r.Path = pathz.MustParse(req.URL.Path)
	checker.True(r.Path.IsAbsolute(), "must be absolute: %s", r.Path)
	r.Query = req.URL.Query()
	r.Headers = req.Header
	r.Body = req.Body
}
