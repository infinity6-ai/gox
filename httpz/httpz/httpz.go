package httpz

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

func ValidateRespSuccess(resp *httpzclient.Resp) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http client error: %d, head body: %s", resp.StatusCode, blobz.MustReadAll(resp.Body, 1024).Ascii())
	}
	return nil
}

type Options struct {
	Client   *httpzclient.Client
	Req      *httpzrequest.Req
	Format   func(ctx context.Context) (io.ReadCloser, error)
	Validate func(resp *httpzclient.Resp) error
	Parse    func(status int, headers http.Header, r io.Reader) error
}

func MustDo(ctx context.Context, opts Options) {
	err := Do(ctx, opts)
	errorz.Check(err)
}

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
		return err
	}
	defer resp.Body.Close()
	if opts.Validate != nil {
		err := opts.Validate(resp)
		if err != nil {
			return fmt.Errorf("error validating response: %w", err)
		}
	}
	if opts.Parse != nil {
		err = opts.Parse(resp.StatusCode, resp.Headers, resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}
	}
	return nil
}
