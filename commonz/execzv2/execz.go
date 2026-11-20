package execzv2

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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

type stateType string

const (
	stateCreated  stateType = "created"
	stateStarted  stateType = "started"
	stateFinished stateType = "finished"
)

type fifoQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []string
	head   int
	closed bool
}

func newFIFOQueue() *fifoQueue {
	q := &fifoQueue{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *fifoQueue) Push(item string) {
	q.mu.Lock()
	q.items = append(q.items, item)
	q.cond.Signal()
	q.mu.Unlock()
}

func (q *fifoQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.cond.Broadcast()
	q.mu.Unlock()
}

func (q *fifoQueue) Pop() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.head >= len(q.items) && !q.closed {
		q.cond.Wait()
	}
	if q.head >= len(q.items) {
		return "", false
	}
	item := q.items[q.head]
	q.head++
	if q.head >= len(q.items) {
		q.items = nil
		q.head = 0
	} else if q.head > 1024 {
		q.items = q.items[q.head:]
		q.head = 0
	}
	return item, true
}

type Process struct {
	Options
	cmd *exec.Cmd

	Err error

	mu         sync.Mutex
	state      stateType
	waitChan   chan struct{}
	stdinPipe  io.WriteCloser
	stdoutPipe io.ReadCloser
	stderrPipe io.ReadCloser
	stdoutQ    *fifoQueue
	stderrQ    *fifoQueue
	wg         sync.WaitGroup
}

func New(ctx context.Context, opts Options) *Process {
	if ctx == nil {
		ctx = context.Background()
	}

	p := &Process{
		Options:  opts,
		cmd:      exec.CommandContext(ctx, opts.Bin, opts.Args...),
		state:    stateCreated,
		waitChan: make(chan struct{}),
		stdoutQ:  newFIFOQueue(),
		stderrQ:  newFIFOQueue(),
	}
	p.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if len(opts.Envs) > 0 {
		env := os.Environ()
		for k, v := range opts.Envs {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		p.cmd.Env = env
	}

	if opts.Dir != "" {
		p.cmd.Dir = opts.Dir
	}

	stdinPipe, err := p.cmd.StdinPipe()
	if err != nil {
		p.Err = fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	p.stdinPipe = stdinPipe

	stdoutPipe, err := p.cmd.StdoutPipe()
	if err != nil && p.Err == nil {
		p.Err = fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	p.stdoutPipe = stdoutPipe

	stderrPipe, err := p.cmd.StderrPipe()
	if err != nil && p.Err == nil {
		p.Err = fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	p.stderrPipe = stderrPipe

	if opts.StdOut != nil {
		p.StdOut = opts.StdOut
	} else {
		p.StdOut = p.stdoutQ.Pop
	}

	if opts.StdErr != nil {
		p.StdErr = opts.StdErr
	} else {
		p.StdErr = p.stderrQ.Pop
	}

	if opts.StdIn != nil {
		p.StdIn = func(str string) {
			opts.StdIn(str)
			p.mu.Lock()
			defer p.mu.Unlock()
			if p.stdinPipe != nil {
				_, _ = p.stdinPipe.Write([]byte(str))
			}
		}
	} else {
		p.StdIn = func(str string) {
			p.mu.Lock()
			defer p.mu.Unlock()
			if p.stdinPipe != nil {
				_, _ = p.stdinPipe.Write([]byte(str))
			}
		}
	}

	return p
}

func (p *Process) startGoroutines() {
	p.wg.Add(2)

	go func() {
		defer p.wg.Done()
		defer p.stdoutQ.Close()
		if p.stdoutPipe == nil {
			return
		}
		reader := bufio.NewReader(p.stdoutPipe)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				line = strings.TrimSuffix(line, "\n")
				line = strings.TrimSuffix(line, "\r")
				p.stdoutQ.Push(line)
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer p.wg.Done()
		defer p.stderrQ.Close()
		if p.stderrPipe == nil {
			return
		}
		reader := bufio.NewReader(p.stderrPipe)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				line = strings.TrimSuffix(line, "\n")
				line = strings.TrimSuffix(line, "\r")
				p.stderrQ.Push(line)
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		p.wg.Wait()
		waitErr := p.cmd.Wait()
		p.mu.Lock()
		if p.Err == nil && waitErr != nil {
			p.Err = fmt.Errorf("process execution failed: %w", waitErr)
		}
		p.state = stateFinished
		close(p.waitChan)
		p.mu.Unlock()
	}()
}

func (p *Process) Start() error {
	p.mu.Lock()
	if p.state != stateCreated {
		p.mu.Unlock()
		return fmt.Errorf("process already started")
	}
	if p.Err != nil {
		p.mu.Unlock()
		return fmt.Errorf("process initialization error: %w", p.Err)
	}
	p.state = stateStarted
	err := p.cmd.Start()
	if err != nil {
		p.state = stateFinished
		p.Err = fmt.Errorf("failed to start process: %w", err)
		p.stdoutQ.Close()
		p.stderrQ.Close()
		close(p.waitChan)
		p.mu.Unlock()
		return p.Err
	}
	p.mu.Unlock()

	p.startGoroutines()
	return nil
}

func (p *Process) Run() error {
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	if err := p.WaitFor(); err != nil {
		return fmt.Errorf("failed to wait for process: %w", err)
	}
	return nil
}

func (p *Process) WaitFor() error {
	<-p.waitChan
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Err != nil {
		return p.Err
	}
	return nil
}

func (p *Process) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stdinPipe != nil {
		err := p.stdinPipe.Close()
		p.stdinPipe = nil
		if err != nil && !errors.Is(err, os.ErrClosed) {
			return fmt.Errorf("failed to close stdin: %w", err)
		}
	}
	return nil
}

func (p *Process) Kill(ctx context.Context, deadline time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	p.mu.Lock()
	if p.cmd == nil || p.cmd.Process == nil {
		p.mu.Unlock()
		return nil
	}
	pid := p.cmd.Process.Pid
	p.mu.Unlock()

	logger.Info(ctx, "sending SIGTERM", map[string]any{"pid": pid})
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
	}

	done := make(chan error, 1)
	go func() {
		done <- p.WaitFor()
	}()

	select {
	case <-ctx.Done():
		logger.Info(ctx, "context cancelled, killing", map[string]any{"pid": pid})
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			_ = p.cmd.Process.Kill()
		}
		if p.stdoutPipe != nil {
			_ = p.stdoutPipe.Close()
		}
		if p.stderrPipe != nil {
			_ = p.stderrPipe.Close()
		}
		killErr := fmt.Errorf("Killed %d: %s [%s], %w", pid, p.cmd.Path, p.cmd.Args, ErrKilled)
		p.mu.Lock()
		p.Err = killErr
		p.mu.Unlock()
		<-p.waitChan
		return killErr
	case <-time.After(deadline):
		logger.Info(ctx, "grace period expired, killing", map[string]any{"pid": pid})
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			_ = p.cmd.Process.Kill()
		}
		if p.stdoutPipe != nil {
			_ = p.stdoutPipe.Close()
		}
		if p.stderrPipe != nil {
			_ = p.stderrPipe.Close()
		}
		killErr := fmt.Errorf("Killed %d: %s [%s], %w", pid, p.cmd.Path, p.cmd.Args, ErrKilled)
		p.mu.Lock()
		p.Err = killErr
		p.mu.Unlock()
		<-p.waitChan
		return killErr
	case err := <-done:
		logger.Info(ctx, "process exited gracefully", map[string]any{"pid": pid})
		if err != nil {
			return fmt.Errorf("process exited with error: %w", err)
		}
		return nil
	}
}
