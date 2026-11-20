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
	originalArgs []string
	cmdOpts      execz.Options
	server       *httpzserver.Server
	prc          *execz.Process
}

func New(ctx context.Context, opts Options) *Browser {
	ret := &Browser{
		server:       httpzserver.New(ctx, opts.Server),
		cmdOpts:      opts.Command,
		originalArgs: opts.Command.Args,
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

	b.fixArgs()
	b.prc = execz.New(ctx, b.cmdOpts)

	err := b.prc.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starging process: %w", err)
	}
	return err
}

func (b *Browser) fixArgs() {
	u := b.server.Base().MustJoinPathString(handlers.ControllerPath)
	b.cmdOpts.Args = make([]string, len(b.originalArgs), len(b.originalArgs)+1)
	d := false
	for idx, arg := range b.originalArgs {
		if arg == "URL" {
			b.cmdOpts.Args[idx] = u.String()
			d = true
		} else {
			b.cmdOpts.Args[idx] = arg
		}
	}
	if !d {
		b.cmdOpts.Args = append(b.cmdOpts.Args, u.String())
	}
}
