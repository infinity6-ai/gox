package httpzclient

import (
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

type CodeError struct {
	Code    int
	Payload string
}

// Error implements [error].
func (c CodeError) Error() string {
	return fmt.Sprintf("http response error: %d, head body: '%s'", c.Code, c.Payload)
}

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

func (r *Resp) Success() error {
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return CodeError{
			Code:    r.StatusCode,
			Payload: blobz.MustReadAll(r.Body, 1024).Ascii(),
		}
	}
	return nil
}

func (r *Resp) MustSuccess() {
	errorz.Check(r.Success())
}
