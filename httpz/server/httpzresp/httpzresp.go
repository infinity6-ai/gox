package httpzresp

import (
	"io"
	"net/http"
)

// The generic wrapper that handles all the nasty HTTP edge cases
type StreamInterceptor struct {
	http.ResponseWriter
	Writer      io.Writer
	wroteHeader bool
	// Optional hook to modify headers before the first byte is written
	BeforeHeader func(int) int
}

func (w *StreamInterceptor) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	if w.BeforeHeader != nil {
		code = w.BeforeHeader(code)
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *StreamInterceptor) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

func (w *StreamInterceptor) Flush() {
	if f, ok := w.Writer.(interface{ Flush() }); ok {
		f.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// type Options struct {
// 	Header      func(w Resp) http.Header
// 	Write       func(w Resp, data []byte) (int, error)
// 	WriteHeader func(w Resp, statusCode int)
// }

// type Resp struct {
// 	Original http.ResponseWriter
// 	opts     Options
// }

// func New(w http.ResponseWriter, opts Options) *Resp {
// 	return &Resp{
// 		Original: w,
// 		opts:     opts,
// 	}
// }

// func (r *Resp) Write(data []byte) (int, error) {
// 	if r.opts.Write == nil {
// 		return r.Original.Write(data)
// 	}
// 	return r.opts.Write(*r, data)
// }

// func (r *Resp) WriteHeader(statusCode int) {
// 	if r.opts.WriteHeader != nil {
// 		r.opts.WriteHeader(*r, statusCode)
// 		return
// 	}
// 	r.Original.WriteHeader(statusCode)
// }

// func (r *Resp) Header() http.Header {
// 	if r.opts.Header == nil {
// 		return r.Original.Header()
// 	}
// 	return r.opts.Header(*r)
// }
