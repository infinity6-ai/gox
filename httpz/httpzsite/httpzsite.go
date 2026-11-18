package httpzsite

import (
	"context"

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
		s.server.AddHandler("GET", "/", httpzserverstatic.StaticHandler(s.Pack))
		s.server.Start()
	}
}

func (s *Site) Close() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}

func (s *Site) Start() {
	s.initServer()
}

func (s *Site) Serve() {
	s.initServer()
	s.server.Serve()
}
