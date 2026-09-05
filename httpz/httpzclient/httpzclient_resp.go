package httpzclient

import (
	"io"
	"net/http"
)

type Resp struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

func (r *Resp) fromHttpResponse(resp *http.Response) {
	r.Status = resp.Status
	r.StatusCode = resp.StatusCode
	r.Headers = resp.Header
	r.Body = resp.Body
}
