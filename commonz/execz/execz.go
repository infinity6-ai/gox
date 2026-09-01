package execz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/infinity6-ai/gox/commonz/logz"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

type stateType string

const (
	created  stateType = "created"
	started  stateType = "started"
	finished stateType = "finished"
)

type Cmd struct {
	ctx      context.Context
	cmd      *exec.Cmd
	state    stateType
	mu       sync.Mutex
	waitChan chan struct{}
	err      error
}

func (c *Cmd) Error() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func New(ctx context.Context, program string, args ...string) *Cmd {
	ret := &Cmd{
		ctx:      ctx,
		cmd:      exec.CommandContext(ctx, program, args...),
		state:    created,
		waitChan: make(chan struct{}),
	}
	ret.cmd.Stderr = os.Stderr
	ret.cmd.Stdout = os.Stderr
	return ret
}

func (c *Cmd) SetDir(dir string) *Cmd {
	c.cmd.Dir = dir
	return c
}

func (c *Cmd) AddEnv(env map[string]string) *Cmd {
	for k, v := range env {
		c.SetEnv(k, v)
	}
	return c
}

func (c *Cmd) SetEnv(k string, v string) {
	c.cmd.Env = append(c.cmd.Env, fmt.Sprintf("%s=%s", k, v))
}

func (c *Cmd) Run() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != created {
		return errors.New("run has already been called")
	}
	c.state = started
	defer func() {
		close(c.waitChan)
		c.state = finished
	}()
	c.err = c.cmd.Run()
	return c.err
}

func (c *Cmd) Wait() error {
	<-c.waitChan
	return c.err
}

func (c *Cmd) Kill() {
	if c.cmd == nil || c.cmd.Process == nil {
		return
	}

	pid := c.cmd.Process.Pid
	logger.Info(c.ctx, "sending SIGTERM", map[string]any{"pid": pid})
	_ = c.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case <-time.After(3 * time.Second):
		logger.Info(c.ctx, "grace period expired, killing", map[string]any{"pid": pid})
		_ = c.cmd.Process.Kill()
	case <-done:
		logger.Info(c.ctx, "process exited gracefully", map[string]any{"pid": pid})
	}
}
