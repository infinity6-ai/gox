package httpzserverv2

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Options struct {
	LocalAddress string
}

func (o *Options) fix() {
	if o.LocalAddress == "" {
		port := os.Getenv("PORT")
		if port != "" {
			o.LocalAddress = fmt.Sprintf("0.0.0.0:%s", port)
		}
	}
	if o.LocalAddress == "" {
		o.LocalAddress = "localhost:8080"
	}
}

type Server struct {
	Context         context.Context
	Options         Options
	listener        net.Listener
	filters         []Filter
	patternHandlers []PatternHandler
	dfz             *deferz.Deferz
}

func New(ctx context.Context, opts Options) *Server {
	opts.fix()
	return &Server{
		Context: ctx,
		Options: opts,
		dfz:     deferz.New(ctx),
	}
}

func (s *Server) Base() string {
	return fmt.Sprintf("http://%s", s.Addr())
}

func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) Listen() {
	if s.listener != nil {
		panic(fmt.Sprintf("already configured: %s", s.listener.Addr()))
	}
	listener, err := net.Listen("tcp", s.Options.LocalAddress)
	errorz.Check(err)
	s.listener = listener
	s.dfz.AddCloserS(listener)
}

func (s *Server) Close() error {
	return s.dfz.Close()
}

func (s *Server) serve() {
	if s.listener == nil {
		panic("not configured")
	}

	var h Handler
	h = s.route
	for i := len(s.filters) - 1; i >= 0; i-- {
		filter := s.filters[i]
		next := h
		h = func(ctx context.Context, resp *Resp, req *Req) {
			println(i)
			filter(ctx, resp, req, next)
		}
	}

	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &Req{}
			req.fromHttpRequest(r)
			resp := &Resp{}
			resp.fromHttpResponseWriter(w)
			h(r.Context(), resp, req)
		}),
		BaseContext: func(l net.Listener) context.Context {
			return s.Context
		},
	}

	s.dfz.Add(func() {
		_ = httpServer.Shutdown(s.Context)
	})

	err := httpServer.Serve(s.listener)
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("http server failed: %w", err))
	}
}

// Serve runs the server in the current goroutine, blocking until the server is
// shut down.
func (s *Server) Serve() {
	if s.listener == nil {
		panic("not configured")
	}
	s.serve()
}

// Start runs the server in a new goroutine, making it non-blocking.
func (s *Server) Start() {
	if s.listener == nil {
		panic("not configured")
	}
	go s.Serve()
}
