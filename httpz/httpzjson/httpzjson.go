package httpzjson

import (
	"context"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

type Options struct {
	Client   *httpzclient.Client
	Req      *httpzrequest.Req
	Input    func() any
	Validate func(resp *httpzclient.Resp) error
	Output   func(status int, headers http.Header) any
}

func MustDo(ctx context.Context, opts Options) {
	err := Do(ctx, opts)
	errorz.Check(err)
}

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
				return err
			}
			return nil
		}
	}
	return httpz.Do(ctx, o)
}
