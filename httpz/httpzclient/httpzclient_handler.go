package httpzclient

import (
	"context"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
)

// Handler defines the function signature for a client-side request handler.
// It takes a request and returns a response and an error.
type Handler func(ctx context.Context, req *httpzrequest.Req) (*Resp, error)

// Filter is a middleware function that intercepts a request, allowing for modification
// of the request or response. It receives the current request and the next handler in the chain.
type Filter func(ctx context.Context, req *httpzrequest.Req, next Handler) (*Resp, error)
