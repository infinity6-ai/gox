// Package httpzjson provides JSON-oriented HTTP request helpers built on top of httpz,
// enabling automated serialization of request inputs and deserialization of response outputs.
package httpzjson

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

// Options configures a JSON HTTP request and response execution workflow.
type Options struct {
	// Client is the HTTP client used to send the request.
	Client *httpzclient.Client
	// Req is the HTTP request to execute. Req.Body must be nil.
	Req *httpzrequest.Req
	// Input optionally provides a value to be JSON-serialized into the request body.
	Input func() any
	// Validate optionally verifies the response status. If nil, httpz.ValidateRespSuccess is used.
	Validate func(resp *httpzclient.Resp) error
	// Output optionally returns a destination pointer into which the JSON response body will be decoded.
	Output func(status int, headers http.Header) any
}

// MustDo executes a JSON HTTP request according to the provided Options, panicking if an error occurs.
func MustDo(ctx context.Context, opts Options) {
	err := Do(ctx, opts)
	errorz.Check(err)
}

// Do executes a JSON HTTP request according to the provided Options.
//
// If opts.Input is non-nil, its returned value is formatted as a JSON ReadCloser for the request body.
// The request is executed and validated via httpz.Do. If opts.Output is non-nil and returns a destination
// pointer, the response body is JSON-decoded into that value.
func Do(ctx context.Context, opts Options) error {
	o := httpz.Options{
		Client:   opts.Client,
		Req:      opts.Req,
		Validate: opts.Validate,
	}
	if opts.Input != nil {
		o.Format = func(ctx context.Context) (io.ReadCloser, error) {
			return jsonz.FormatReadCloser(opts.Input()), nil
		}
	}
	if opts.Output != nil {
		o.Parse = func(status int, headers http.Header, r io.Reader) error {
			v := opts.Output(status, headers)
			if v != nil {
				_, err := jsonz.ParseReader(r, v)
				if err != nil {
					return fmt.Errorf("failed to parse json response: %w", err)
				}
				return nil
			}
			return nil
		}
	}
	err := httpz.Do(ctx, o)
	if err != nil {
		return fmt.Errorf("failed to execute json request: %w", err)
	}
	return nil
}
