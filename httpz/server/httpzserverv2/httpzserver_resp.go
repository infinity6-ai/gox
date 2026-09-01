package httpzserverv2

import (
	"net/http"
)

type Resp struct {
	Status  int
	Headers http.Header

	headersSent bool
	w           http.ResponseWriter
}

func (r *Resp) Flush() {
	r.sendHeaders()
	r.w.(http.Flusher).Flush()
}

func (r *Resp) Write(p []byte) (n int, err error) {
	r.sendHeaders()
	return r.w.Write(p)
}

func (r *Resp) sendHeaders() {
	if r.headersSent {
		return
	}
	r.headersSent = true
	r.w.WriteHeader(r.Status)
}

func (r *Resp) fromHttpResponseWriter(w http.ResponseWriter) {
	r.w = w
	r.Status = -1
	r.Headers = w.Header()
}
