package execzv2

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/infinity6-ai/gox/commonz/logz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

var ErrKilled = errors.New("ErrKilled")

type Options struct {
	Bin    string
	Args   []string
	Envs   map[string]string
	Dir    string
	StdOut func() (string, bool)
	StdErr func() (string, bool)
	StdIn  func(str string)
}

type Process struct {
	cmd *exec.Cmd

	Err error
}

func New(ctx context.Context, opts Options) *Process {
	panic("implement it")
}

func (p *Process) Close() error {
	panic("implement it")
}

func (p *Process) WaitFor() error {
	panic("implement it")
}

func (p *Process) Start() error {
	panic("implement it")
}

func (p *Process) Run() error {
	panic("implement it")
}

func (p *Process) Kill(ctx context.Context, deadline time.Duration) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	pid := p.cmd.Process.Pid
	logger.Info(ctx, "sending SIGTERM", map[string]any{"pid": pid})
	_ = p.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case <-time.After(deadline):
		logger.Info(ctx, "grace period expired, killing", map[string]any{"pid": pid})
		_ = p.cmd.Process.Kill()
		return fmt.Errorf("Killed %d: %s [%s], %w", p.cmd.Process.Pid, p.cmd.Path, p.cmd.Args, ErrKilled)
	case err := <-done:
		logger.Info(ctx, "process exited gracefully", map[string]any{"pid": pid})
		return err
	}
}
