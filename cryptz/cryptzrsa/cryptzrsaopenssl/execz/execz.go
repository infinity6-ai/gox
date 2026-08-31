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
)

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

	// pid := c.cmd.Process.Pid
	// logger.Info(c.ctx, "sending SIGTERM", map[string]any{"pid": pid})
	_ = c.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case <-time.After(3 * time.Second):
		// logger.Info(c.ctx, "grace period expired, killing", map[string]any{"pid": pid})
		_ = c.cmd.Process.Kill()
	case <-done:
		// logger.Info(c.ctx, "process exited gracefully", map[string]any{"pid": pid})
	}
}

// func (c *Cmd) Run() error {
// 	if err := c.start(); err != nil {
// 		return err
// 	}
// 	return c.cmd.Wait()
// }

// // Proc represents a managed OS process
// type Proc struct {
// 	ctx  context.Context
// 	cmd  *exec.Cmd
// 	done chan any
// 	err  error
// }

// func (me *Proc) start(out func(fd int, line string)) {
// 	success := false
// 	stdout := make(chan string, 1024)
// 	stderr := make(chan string, 1024)

// 	wout := ioz.NewChanLineWriter(stdout, 512)
// 	werr := ioz.NewChanLineWriter(stderr, 512)

// 	defer func() {
// 		if !success {
// 			close(stdout)
// 			close(stderr)
// 			wout.Close()
// 			werr.Close()
// 		}
// 	}()

// 	me.cmd.Stdout = wout
// 	me.cmd.Stderr = werr

// 	err := me.cmd.Start()
// 	util.Check(err)

// 	me.done = make(chan any, 1)

// 	// Stream monitoring goroutine
// 	go func() {
// 		defer close(me.done)
// 		stdoutClosed, stderrClosed := false, false
// 		for {
// 			select {
// 			case l, ok := <-stdout:
// 				if !ok {
// 					stdoutClosed = true
// 				} else {
// 					out(1, l)
// 				}
// 			case l, ok := <-stderr:
// 				if !ok {
// 					stderrClosed = true
// 				} else {
// 					out(2, l)
// 				}
// 			}
// 			if stdoutClosed && stderrClosed {
// 				return
// 			}
// 		}
// 	}()

// 	// Waiter goroutine
// 	go func() {
// 		defer close(stdout)
// 		defer close(stderr)
// 		defer wout.Close()
// 		defer werr.Close()
// 		me.err = me.cmd.Wait()
// 	}()

// 	success = true
// }

// func (me *Proc) StartWithLogs() {
// 	me.Start(func(fd int, line string) {
// 		msg := fmt.Sprintf("[%d] FD%d: %s", me.Pid(), fd, line)
// 		logger.Info(me.ctx, "process output", map[string]any{"line": msg})
// 	})
// }

// func (me *Proc) Pid() int {
// 	if me.cmd != nil && me.cmd.Process != nil {
// 		return me.cmd.Process.Pid
// 	}
// 	return 0
// }

// func (me *Proc) Kill() {
// 	Kill(me.ctx, me.cmd)
// }

// func (me *Proc) Wait() {
// 	<-me.done
// 	util.Check(me.err)
// }

// func Kill(ctx context.Context, cmd *exec.Cmd) {
// 	if cmd == nil || cmd.Process == nil {
// 		return
// 	}

// 	pid := cmd.Process.Pid
// 	logger.Info(ctx, "sending SIGTERM", map[string]any{"pid": pid})
// 	_ = cmd.Process.Signal(syscall.SIGTERM)

// 	done := make(chan error, 1)
// 	go func() {
// 		done <- cmd.Wait()
// 	}()

// 	select {
// 	case <-time.After(3 * time.Second):
// 		logger.Info(ctx, "grace period expired, killing", map[string]any{"pid": pid})
// 		_ = cmd.Process.Kill()
// 	case <-done:
// 		logger.Info(ctx, "process exited gracefully", map[string]any{"pid": pid})
// 	}
// }
