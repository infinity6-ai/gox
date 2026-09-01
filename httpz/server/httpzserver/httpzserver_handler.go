package httpzserver

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
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
	Pattern *pathz.Path
	Prefix  bool
	Handler HandlerPattern
}

func (s *Server) AddHandler(method string, pattern string, handler HandlerPattern) {
	checker.OneOf(allowedMethods, method, "unknown method")
	p := pathz.MustParse(pattern)
	p.Check(pathz.ValidateOptions{
		Absolute:    new(true),
		Wildchar:    true,
		EndingSlash: new(false),
	})
	ph := &PatternHandler{
		Method:  method,
		Pattern: p,
		Prefix:  false,
		Handler: handler,
	}
	dir, base, _ := p.Parent()
	if base == "*" {
		ph.Pattern = dir
		ph.Prefix = true
	}
	s.patternHandlers = append(s.patternHandlers, *ph)
}

func (s *Server) route(ctx context.Context, resp Resp, req *Req) {
	actualParts := req.Path.Parts()
	for _, ph := range s.patternHandlers {
		if ph.Method != "*" && ph.Method != req.Method {
			continue
		}
		params, ok := match(ph.Pattern.Parts(), actualParts, ph.Prefix)
		if ok {
			ph.Handler(ctx, resp, req, params)
			return
		}
	}
	w := resp(http.StatusNotFound, nil)
	_, err := w.Write([]byte("Not Found"))
	errorz.Check(err)
}

func match(patternParts, actualParts []string, prefix bool) (map[string]string, bool) {
	if !prefix && len(patternParts) != len(actualParts) {
		return nil, false
	}
	if prefix && len(actualParts) < len(patternParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, patternPart := range patternParts {
		actualPart := actualParts[i]

		if strings.HasPrefix(patternPart, "{") && strings.HasSuffix(patternPart, "}") {
			paramName := patternPart[1 : len(patternPart)-1]
			params[paramName] = actualPart
		} else if patternPart != actualPart {
			return nil, false
		}
	}

	if !prefix && len(patternParts) != len(actualParts) {
		return nil, false
	}

	return params, true
}
