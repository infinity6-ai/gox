package httpzsite

import (
	"context"
	"fmt"
	"net"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterauthbasic"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

type Site struct {
	Context      context.Context
	LocalAddress string
	Pack         any
	BasicUsers   map[string]string

	server *httpzserver.Server
}

// ExtractTo extracts the site's static pack assets to the specified destination path.
func (s *Site) ExtractTo(ctx context.Context, path *pathz.Path) error {
	err := staticz.ExtractTo(ctx, path, s.Pack)
	if err != nil {
		return fmt.Errorf("failed to extract site pack to %s: %w", path, err)
	}
	return nil
}

func (s *Site) Listen() {
	if s.server == nil {
		s.server = httpzserver.New(s.Context, httpzserver.Options{
			LocalAddress: s.LocalAddress,
		})
	}
	s.server.Listen()
}

func (s *Site) AddHandlers() {
	if s.BasicUsers != nil {
		s.server.AddFilter(httpzserverfilterauthbasic.New(s.Context, httpzserverfilterauthbasic.Options{
			Verify: func(ctx context.Context, user string, pass string) (bool, error) {
				expectedPass, ok := s.BasicUsers[user]
				if !ok {
					return false, nil
				}
				return expectedPass == pass, nil
			},
		}))
	}
	s.server.AddHandler("GET", "/*", httpzserverstatic.StaticHandler(s.Pack))
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
	s.server.Start()
}

func (s *Site) Serve() {
	s.server.Serve()
}
