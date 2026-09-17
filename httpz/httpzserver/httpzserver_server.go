// Package httpzserver provides an HTTP server featuring pattern-based URL routing,
// filter middleware pipelines, and structured lifecycle management.
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
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/internal/httpzhelper"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

// Options configures a Server instance.
type Options struct {
	// LocalAddress is the TCP address to bind the server to (e.g. "localhost:8080" or ":8080").
	// If empty, it checks the PORT environment variable or defaults to "localhost:0".
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

// Server is an HTTP server that routes requests using pattern matching and processes them
// through middleware filters.
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

// New creates a new Server instance with the specified Options.
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

// Base returns the base URL of the running server, using the listening address.
func (s *Server) Base() *urlz.Url {
	return &urlz.Url{
		Scheme: "http",
		Host:   s.Addr().String(),
		Path:   pathz.MustParse("/"),
	}
}

// Addr returns the network listener address, or nil if the server is not listening.
func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// Listen creates and starts the underlying TCP network listener.
// It panics if the server is already listening or if binding fails.
func (s *Server) Listen() {
	if s.listener != nil {
		panic(fmt.Sprintf("already configured: %s", s.listener.Addr()))
	}
	listener, err := net.Listen("tcp", s.Options.LocalAddress)
	errorz.Check(err)
	s.listener = listener
	s.dfz.AddCloserS(listener)
}

// Close gracefully stops the server and releases all bound resources.
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
		h = func(ctx context.Context, resp Resp, req *httpzrequest.Req) {
			filter(ctx, resp, req, next)
		}
	}

	s.httpServer = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := &httpzrequest.Req{}
			httpzhelper.FromHttpRequest(r, req)
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
