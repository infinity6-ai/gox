package httpzjson

import (
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

type Options struct {
	Client   *httpzclient.Client
	Req      *httpzrequest.Req
	Input    any
	Validate func(resp *httpzclient.Resp) error
	Output   any
}

// func Do(ctx context.Context, opts Options) error {

// }
