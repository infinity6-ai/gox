package httpzclient

import (
	"io"
	"net/http"
)

// Resp represents an HTTP response returned by a Client request.
type Resp struct {
	// Status is the HTTP status text (e.g. "200 OK").
	Status string
	// StatusCode is the HTTP response status code (e.g. 200).
	StatusCode int
	// Headers contains the HTTP response headers.
	Headers http.Header
	// Body is the response payload stream, which must be closed by the caller.
	Body io.ReadCloser
}

func (r *Resp) fromHttpResponse(resp *http.Response) {
	r.Status = resp.Status
	r.StatusCode = resp.StatusCode
	r.Headers = resp.Header
	r.Body = resp.Body
}
