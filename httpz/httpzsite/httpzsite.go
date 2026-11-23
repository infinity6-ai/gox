package httpzsite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/trokerz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterauthbasic"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

type Site struct {
	Context       context.Context
	LocalAddress  string
	Pack          any
	BasicUsers    map[string]string
	TrokerOptions []trokerz.Options

	server *httpzserver.Server
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
	if len(s.TrokerOptions) > 0 {
		s.server.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
			var buf bytes.Buffer
			var origWriter io.Writer

			nResp := func(status int, headers http.Header) io.Writer {
				if headers != nil {
					headers.Del("Content-Length")
				}
				origWriter = resp(status, headers)
				return &buf
			}

			next(ctx, nResp, req)

			if origWriter != nil {
				content := buf.String()
				var err error
				for _, opt := range s.TrokerOptions {
					content, err = trokerz.Troke(content, opt)
					if err != nil {
						errorz.Check(fmt.Errorf("failed to apply troker: %w", err))
					}
				}
				_, err = origWriter.Write([]byte(content))
				errorz.Check(err)
			}
		})
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
