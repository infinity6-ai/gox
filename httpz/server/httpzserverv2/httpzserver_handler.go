package httpzserverv2

import "net/http"

type Handler interface {
	HandleHttpz(resp *Resp, req *Req)
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
