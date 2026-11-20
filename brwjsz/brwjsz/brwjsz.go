package brwjsz

import (
	"context"
	"errors"
	"fmt"

	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type Options struct {
	Command execz.Options
	Server  httpzserver.Options
}

type Browser struct {
	server *httpzserver.Server
	prc    *execz.Process
}

func New(ctx context.Context, opts Options) *Browser {
	ret := &Browser{
		server: httpzserver.New(ctx, opts.Server),
		prc:    execz.New(ctx, opts.Command),
	}
	handlers.AddHanlers(ctx, ret.server)
	return ret
}

func (b *Browser) Close() error {
	var prcErr, serverErr error
	if b.prc != nil {
		prcErr = b.prc.Close()
	}
	if b.server != nil {
		serverErr = b.server.Close()
	}
	err := errors.Join(prcErr, serverErr)
	if err != nil {
		return fmt.Errorf("error closing: %w", err)
	}
	return err
}

func (b *Browser) Start(ctx context.Context) error {
	b.server.Listen()
	err := b.prc.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starging process: %w", err)
	}
	return err
}
