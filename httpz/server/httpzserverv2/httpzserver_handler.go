package httpzserverv2

import (
	"context"
	"net/http"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request)

type Filter func(next Handler) Handler

func (s *Server) Push(filter Filter) {
	s.Handler = filter(s.Handler)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Req{}
	req.fromHttpRequest(r)
	// ctx := r.Context()
	resp := &Resp{}
	resp.fromHttpResponseWriter(w)

	resp.Status = http.StatusBadRequest
	resp.Headers.Set("x", "a")
	resp.Write([]byte("nok"))

	// resp.fromHttpResponseWriter(ctx, w)
	// w.WriteHeader(http.StatusNoContent)
}
