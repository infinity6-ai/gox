package httpzsite

import (
	"context"
	"net"

	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

type Site struct {
	Context      context.Context
	LocalAddress string
	Pack         any

	server *httpzserver.Server
}

func (s *Site) initServer() {
	if s.server == nil {
		s.server = httpzserver.New(s.Context, httpzserver.Options{
			LocalAddress: s.LocalAddress,
		})
		s.server.Listen()
		s.server.AddHandler("GET", "/*", httpzserverstatic.StaticHandler(s.Pack))
	}
}

// Addr returns the network address the server is listening on, or nil if not listening.
func (s *Site) Addr() net.Addr {
	if s.server == nil {
		return nil
	}
	return s.server.Addr()
}

// Base returns the base URL of the running site, using the listening address.
func (s *Site) Base() *urlz.Url {
	if s.server == nil {
		return nil
	}
	return s.server.Base()
}

func (s *Site) Close() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}

func (s *Site) Start() {
	s.initServer()
	s.server.Start()
}

func (s *Site) Serve() {
	s.initServer()
	s.server.Serve()
}
