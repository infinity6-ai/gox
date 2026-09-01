package httpzserverv2

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/pathz"
)

type Handler func(ctx context.Context, resp *Resp, req *Req)

type HandlerPattern func(ctx context.Context, resp *Resp, req *Req, params map[string]string)

type Filter func(ctx context.Context, resp *Resp, req *Req, next Handler)

func (s *Server) AddFilter(filter Filter) {
	s.filters = append(s.filters, filter)
}

type PatternHandler struct {
	Pattern *pathz.Path
	Prefix  bool
	Handler Handler
}

func (s *Server) AddHandler(method string, pattern string, handler Handler) {
	p := pathz.MustParse(pattern)
	p.Check(pathz.ValidateOptions{
		Absolute:    new(true),
		Wildchar:    true,
		EndingSlash: new(false),
	})
	ph := &PatternHandler{
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
