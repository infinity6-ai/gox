// Package httpz coordinates end-to-end HTTP request and response lifecycles,
// including body formatting, request execution, response status validation,
// and response payload parsing.
package httpz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

// ErrStatus is returned when an HTTP response status code indicates an error (outside the 200-299 range).
var ErrStatus = errors.New("Http Status Error")

var errStatus = ErrStatus

// ValidateRespSuccess verifies that the HTTP response has a 2xx status code.
// It returns ErrStatus if the status code is less than 200 or greater than or equal to 300.
func ValidateRespSuccess(resp *httpzclient.Resp) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrStatus
	}
	return nil
}

// Options specifies the configuration and lifecycle callbacks for executing an HTTP request with Do or MustDo.
type Options struct {
	// Client is the HTTP client used to perform the request.
	Client *httpzclient.Client
	// Req is the HTTP request to execute. Req.Body must be nil before Do is called.
	Req *httpzrequest.Req
	// Format optionally prepares the request body stream.
	Format func(ctx context.Context) (io.ReadCloser, error)
	// Validate checks whether the response is acceptable. If nil, ValidateRespSuccess is used.
	Validate func(resp *httpzclient.Resp) error
	// Parse optionally parses the response status, headers, and body.
	Parse func(status int, headers http.Header, r io.Reader) error
}

// MustDo executes an HTTP request according to the provided Options and panics if any error occurs.
func MustDo(ctx context.Context, opts Options) {
	err := Do(ctx, opts)
	errorz.Check(err)
}

// Do executes an HTTP request according to the provided Options.
//
// It verifies that opts.Req.Body is initially nil (panicking otherwise).
// If opts.Format is provided, it is invoked to produce a request body ReadCloser which is
// automatically closed after execution.
// The request is dispatched via opts.Client.
// The response is validated using opts.Validate (defaulting to ValidateRespSuccess).
// If validation succeeds and opts.Parse is provided, the response is parsed.
func Do(ctx context.Context, opts Options) error {
	checker.Nil(opts.Req.Body, "request body must be nil")
	if opts.Format != nil {
		r, err := opts.Format(ctx)
		if err != nil {
			return fmt.Errorf("error creating request body: %w", err)
		}
		defer r.Close()
		opts.Req.Body = r
	}
	resp, err := opts.Client.Do(ctx, opts.Req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	validate := opts.Validate
	if validate == nil {
		validate = ValidateRespSuccess
	}
	err = validate(resp)
	if err != nil {
		return fmt.Errorf("http validating error: %d, head body: %s, err: %w", resp.StatusCode, blobz.MustReadAll(resp.Body, 1024).Ascii(), err)
	}
	if opts.Parse != nil {
		err = opts.Parse(resp.StatusCode, resp.Headers, resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}
	}
	return nil
}
