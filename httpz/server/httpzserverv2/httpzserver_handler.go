package httpzserverv2

import (
	"context"
	"net/http"
)

type Handler func(ctx context.Context, resp *Resp, req *Req)

type Filter func(next Handler) Handler

type FilterX func(ctx context.Context, resp *Resp, req *Req, next Handler)

func (s *Server) AddX(filter FilterX) {
	s.Add(func(next Handler) Handler {
		return func(ctx context.Context, resp *Resp, req *Req) {
			filter(ctx, resp, req, next)
		}
	})
}

func (s *Server) Add(filter Filter) {
	s.Filters = append(s.Filters, filter)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Req{}
	req.fromHttpRequest(r)
	// ctx := r.Context()
	resp := &Resp{}
	resp.fromHttpResponseWriter(w)

	// s.Handler(ctx, resp, req)

	// resp.Status = http.StatusBadRequest
	// resp.Headers.Set("x", "a")
	// resp.Write([]byte("nok"))

	// resp.fromHttpResponseWriter(ctx, w)
	// w.WriteHeader(http.StatusNoContent)
}
