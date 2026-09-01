package httpzserverv2

import (
	"context"
)

type Handler func(ctx context.Context, resp *Resp, req *Req)

type FilterX func(ctx context.Context, resp *Resp, req *Req, next Handler)

func (s *Server) AddX(filter FilterX) {
	s.Filters = append(s.Filters, func(next Handler) Handler {
		return func(ctx context.Context, resp *Resp, req *Req) {
			filter(ctx, resp, req, next)
		}
	})
}
