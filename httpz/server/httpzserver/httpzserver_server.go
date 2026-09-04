package httpzserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/infinity6-ai/gox/commonz/constraintz"
	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/syncz/promise"
	"github.com/infinity6-ai/gox/commonz/urlz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

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
		o.LocalAddress = "localhost:0"
	}
}

type Server struct {
	Context         context.Context
	Options         Options
	listener        net.Listener
	filters         []Filter
	patternHandlers []PatternHandler
	dfz             *deferz.Deferz
	servePromise    *promise.Promise[constraintz.Void]
	httpServer      *http.Server
}

func New(ctx context.Context, opts Options) *Server {
	opts.fix()
	ret := &Server{
		Context: ctx,
		Options: opts,
		dfz:     deferz.New(ctx),
	}
	ret.dfz.Add(func() {
		if ret.servePromise != nil {
			ret.servePromise.GetV()
		}
	})
	return ret
}

func (s *Server) Base() *urlz.Url {
	return &urlz.Url{
		Scheme: "http",
		Host:   s.Addr().String(),
		Path:   pathz.MustParse("/"),
	}
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
	if s.dfz != nil {
		return s.dfz.Close()
	}
	return nil
}

func (s *Server) internalServe() {
	err := s.httpServer.Serve(s.listener)
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Errorf("http server failed: %w", err))
	}
}

func (s *Server) startServer() {
	if s.listener == nil {
		panic("not configured")
	}
	if s.httpServer != nil {
		panic("already configured")
	}
	var h Handler
	h = s.route
	for i := len(s.filters) - 1; i >= 0; i-- {
		filter := s.filters[i]
		next := h
		h = func(ctx context.Context, resp Resp, req *Req) {
			filter(ctx, resp, req, next)
		}
	}

	s.httpServer = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &Req{}
			req.fromHttpRequest(r)
			resp := func(status int, headers http.Header) io.Writer {
				for k, v := range headers {
					w.Header()[k] = v
				}
				w.WriteHeader(status)
				return w
			}
			h(r.Context(), resp, req)
		}),
		BaseContext: func(l net.Listener) context.Context {
			return s.Context
		},
	}

	s.dfz.Add(func() {
		_ = s.httpServer.Shutdown(s.Context)
	})
}

// Serve runs the server in the current goroutine, blocking until the server is
// shut down.
func (s *Server) Serve() {
	if s.listener == nil {
		panic("not configured")
	}
	s.startServer()
	logger.Info(s.Context, "Server Started", map[string]any{"addr": s.listener.Addr().String()})
	s.internalServe()
}

// Start runs the server in a new goroutine, making it non-blocking.
func (s *Server) Start() {
	if s.listener == nil {
		panic("not configured")
	}
	s.startServer()
	s.servePromise = promise.AsyncV(s.Context, s.internalServe)
}
