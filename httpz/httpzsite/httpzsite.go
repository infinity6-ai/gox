package httpzsite

import (
	"context"
	"net"
	"sync"

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

type Site struct {
	Context      context.Context
	LocalAddress string
	Pack         any

	mu     sync.RWMutex
	server *httpzserver.Server
}

func (s *Site) initServer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server == nil {
		server := httpzserver.New(s.Context, httpzserver.Options{
			LocalAddress: s.LocalAddress,
		})
		server.Listen()
		server.AddHandler("GET", "/*", httpzserverstatic.StaticHandler(s.Pack))
		s.server = server
	}
}

// Addr returns the network address the server is listening on, or nil if not listening.
func (s *Site) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.server == nil {
		return nil
	}
	return s.server.Addr()
}

// Base returns the base URL of the running site, using the listening address.
func (s *Site) Base() *urlz.Url {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.server == nil {
		return nil
	}
	return s.server.Base()
}

func (s *Site) Close() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}

func (s *Site) Start() {
	s.initServer()
	s.mu.RLock()
	srv := s.server
	s.mu.RUnlock()
	srv.Start()
}

func (s *Site) Serve() {
	s.initServer()
	s.mu.RLock()
	srv := s.server
	s.mu.RUnlock()
	srv.Serve()
}
