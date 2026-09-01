package httpzserverv2

import (
	"context"
)

type Handler func(ctx context.Context, resp *Resp, req *Req)

type FilterX func(ctx context.Context, resp *Resp, req *Req, next Handler)

func (s *Server) AddX(filter FilterX) {
	s.Filters = append(s.Filters, filter)
}
