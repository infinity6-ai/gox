package httpzserverv2

import (
	"context"
)

type Handler func(ctx context.Context, resp *Resp, req *Req)

type Filter func(ctx context.Context, resp *Resp, req *Req, next Handler)

func (s *Server) AddFilter(filter Filter) {
	s.Filters = append(s.Filters, filter)
}
