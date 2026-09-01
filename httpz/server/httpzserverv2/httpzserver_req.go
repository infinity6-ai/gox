package httpzserverv2

import (
	"io"
	"net/http"
	"net/url"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

type Req struct {
	Path    *pathz.Path
	Query   url.Values
	Headers http.Header
	Body    io.Reader
}

func (r *Req) fromHttpRequest(req *http.Request) {
	r.Path = pathz.MustParse(req.URL.Path)
	r.Query = req.URL.Query()
	r.Headers = req.Header
	r.Body = req.Body
}
