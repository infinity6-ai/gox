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

type contextFilters struct {
	outer []Filter
	inner []Filter
}

type contextKeyFilters struct{}

var filtersKey = contextKeyFilters{}

// WithFilters returns a new context containing the specified outer and inner filters.
// Outer filters execute before client-level filters; inner filters execute after client-level filters.
// If both outer and inner are empty, ctx is returned unchanged.
// Subsequent calls to WithFilters append to any filters already present in ctx.
func WithFilters(ctx context.Context, outer []Filter, inner []Filter) context.Context {
	if len(outer) == 0 && len(inner) == 0 {
		return ctx
	}

	existing, _ := ctx.Value(filtersKey).(*contextFilters)

	var next contextFilters
	if existing != nil {
		next.outer = append(append([]Filter(nil), existing.outer...), outer...)
		next.inner = append(append([]Filter(nil), existing.inner...), inner...)
	} else {
		next.outer = append([]Filter(nil), outer...)
		next.inner = append([]Filter(nil), inner...)
	}

	return context.WithValue(ctx, filtersKey, &next)
}

