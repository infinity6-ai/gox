package httpzserver

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/pathz/patternpathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

var allowedMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
	"*",
}

type Handler func(ctx context.Context, resp Resp, req *Req)

func (h Handler) WrapResponse(ctx context.Context, resp Resp, req *Req, fixHeaders func(outStatus int, outHeaders http.Header) int, fixWriter func(outWriter io.Writer) io.Writer) io.Writer {
	var rWriter io.Writer
	nResp := func(nStatus int, nHeaders http.Header) io.Writer {
		status := fixHeaders(nStatus, nHeaders)
		rWriter = resp(status, nHeaders)
		rWriter = fixWriter(rWriter)
		return rWriter
	}
	h(ctx, nResp, req)
	return rWriter
}

type HandlerPattern func(ctx context.Context, resp Resp, req *Req, params map[string]string)

type Filter func(ctx context.Context, resp Resp, req *Req, next Handler)

func (s *Server) AddFilter(filter Filter) {
	s.filters = append(s.filters, filter)
}

type PatternHandler struct {
	Method  string
	Pattern *patternpathz.PathPattern
	Prefix  bool
	Handler HandlerPattern
}

func (s *Server) AddHandler(method string, pattern string, handler HandlerPattern) {
	checker.OneOf(allowedMethods, method, "unknown method")
	ph := &PatternHandler{}
	p := pathz.MustParse(pattern)
	p.Check(pathz.ValidateOptions{
		Absolute:    new(true),
		Wildchar:    true,
		EndingSlash: new(false),
	})
	dir, base, _ := p.Parent()
	if base == "*" {
		p = dir
		ph.Prefix = true
	}
	pt := patternpathz.MustParse(p)
	ph.Method = method
	ph.Pattern = pt
	ph.Handler = handler
	s.patternHandlers = append(s.patternHandlers, *ph)
}

func (s *Server) route(ctx context.Context, resp Resp, req *Req) {
	for _, ph := range s.patternHandlers {
		if ph.Method != "*" && ph.Method != req.Method {
			continue
		}
		params, ok := match(ph.Pattern, req.Path, ph.Prefix)
		if ok {
			ph.Handler(ctx, resp, req, params)
			return
		}
	}
	w := resp(http.StatusNotFound, nil)
	_, err := w.Write([]byte("Not Found"))
	errorz.Check(err)
}

func match(pattern *patternpathz.PathPattern, p *pathz.Path, prefix bool) (map[string]string, bool) {
	params, suffix, err := pattern.Parse(p)
	if err != nil {
		if errors.Is(err, patternpathz.ErrMismatch) {
			return nil, false
		}
		errorz.Check(err)
	}
	if !prefix && suffix != nil {
		return nil, false
	}
	return params, true

}
