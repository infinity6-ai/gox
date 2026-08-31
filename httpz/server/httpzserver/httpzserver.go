package httpzserver

import (
	"context"
	"fmt"
	"net"
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
	Context context.Context
	Options Options

	dfz *deferz.Deferz
}

func New(ctx context.Context, opts Options) *Server {
	opts.fix()
	return &Server{
		Context: ctx,
		Options: opts,
		dfz:     deferz.New(ctx),
	}
}

func (s *Server) Listen() {
	dfz := deferz.New(s.Context)
	defer dfz.Close()

	listener, err := net.Listen("tcp", s.Options.LocalAddress)
	errorz.Check(err)
	dfz.AddCloserS(listener)

	s.dfz.Add(dfz.Detach().Do)
}

func (s *Server) Close() error {
	return s.dfz.Close()
}

func (s *Server) Start() {
	panic("implement it")
}
