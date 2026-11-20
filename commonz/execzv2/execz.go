package execzv2

import (
	"bytes"
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

const DefaultLineBufferSize = 64 * 1024

type Line struct {
	Line    string
	Partial bool
}

type Options struct {
	Bin            string
	Args           []string
	Envs           map[string]string
	Dir            string
	CaptureStdout  int
	CaptureStderr  int
	EnableStdin    int
	LineBufferSize int
}

type stateType string

const (
	stateCreated  stateType = "created"
	stateStarted  stateType = "started"
	stateFinished stateType = "finished"
)

type Process struct {
	Options
	cmd *exec.Cmd

	Err error

	mu             sync.Mutex
	state          stateType
	waitChan       chan struct{}
	killChan       chan struct{}
	closeKillOnce  sync.Once
	stdinPipe      io.WriteCloser
	stdoutPipe     io.ReadCloser
	stderrPipe     io.ReadCloser
	stdinChan      chan Line
	stdoutChan     chan Line
	stderrChan     chan Line
	closeStdinOnce sync.Once
	wg             sync.WaitGroup
}

func New(ctx context.Context, opts Options) *Process {
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.LineBufferSize <= 0 {
		opts.LineBufferSize = DefaultLineBufferSize
	}

	p := &Process{
		Options:  opts,
		cmd:      exec.CommandContext(ctx, opts.Bin, opts.Args...),
		state:    stateCreated,
		waitChan: make(chan struct{}),
		killChan: make(chan struct{}),
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

	if opts.EnableStdin > 0 {
		p.stdinChan = make(chan Line, opts.EnableStdin)
		stdinPipe, err := p.cmd.StdinPipe()
		if err != nil {
			p.Err = fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		p.stdinPipe = stdinPipe
	}

	if opts.CaptureStdout > 0 {
		p.stdoutChan = make(chan Line, opts.CaptureStdout)
		stdoutPipe, err := p.cmd.StdoutPipe()
		if err != nil && p.Err == nil {
			p.Err = fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		p.stdoutPipe = stdoutPipe
	}

	if opts.CaptureStderr > 0 {
		p.stderrChan = make(chan Line, opts.CaptureStderr)
		stderrPipe, err := p.cmd.StderrPipe()
		if err != nil && p.Err == nil {
			p.Err = fmt.Errorf("failed to create stderr pipe: %w", err)
		}
		p.stderrPipe = stderrPipe
	}

	return p
}

func (p *Process) GetStdout() <-chan Line {
	return p.stdoutChan
}

func (p *Process) GetStdOut() <-chan Line {
	return p.GetStdout()
}

func (p *Process) GetStderr() <-chan Line {
	return p.stderrChan
}

func (p *Process) GetStdErr() <-chan Line {
	return p.GetStderr()
}

func (p *Process) GetStdin() chan<- Line {
	return p.stdinChan
}

func (p *Process) GetStdIn() chan<- Line {
	return p.GetStdin()
}

func readStreamLines(r io.Reader, maxBuf int, killChan <-chan struct{}, outChan chan<- Line) {
	defer close(outChan)
	if r == nil {
		return
	}

	buf := make([]byte, 4096)
	var lineBuf []byte

	sendLine := func(l Line) bool {
		select {
		case <-killChan:
			return false
		case outChan <- l:
			return true
		}
	}

	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			for len(chunk) > 0 {
				idx := bytes.IndexByte(chunk, '\n')
				if idx >= 0 {
					lineBuf = append(lineBuf, chunk[:idx]...)
					chunk = chunk[idx+1:]
					for len(lineBuf) >= maxBuf {
						part := string(lineBuf[:maxBuf])
						lineBuf = append([]byte(nil), lineBuf[maxBuf:]...)
						if !sendLine(Line{Line: part, Partial: true}) {
							return
						}
					}
					str := string(lineBuf)
					str = strings.TrimSuffix(str, "\r")
					lineBuf = lineBuf[:0]
					if !sendLine(Line{Line: str, Partial: false}) {
						return
					}
				} else {
					lineBuf = append(lineBuf, chunk...)
					chunk = nil
					for len(lineBuf) >= maxBuf {
						part := string(lineBuf[:maxBuf])
						lineBuf = append([]byte(nil), lineBuf[maxBuf:]...)
						if !sendLine(Line{Line: part, Partial: true}) {
							return
						}
					}
				}
			}
		}
		if err != nil {
			for len(lineBuf) >= maxBuf {
				part := string(lineBuf[:maxBuf])
				lineBuf = append([]byte(nil), lineBuf[maxBuf:]...)
				if !sendLine(Line{Line: part, Partial: true}) {
					return
				}
			}
			if len(lineBuf) > 0 {
				str := string(lineBuf)
				str = strings.TrimSuffix(str, "\r")
				_ = sendLine(Line{Line: str, Partial: true})
			}
			return
		}
	}
}

func (p *Process) startGoroutines() {
	if p.stdinPipe != nil && p.stdinChan != nil {
		go func() {
			defer p.stdinPipe.Close()
			for {
				select {
				case <-p.killChan:
					return
				case <-p.waitChan:
					return
				case item, ok := <-p.stdinChan:
					if !ok {
						return
					}
					toWrite := item.Line
					if !item.Partial && !strings.HasSuffix(toWrite, "\n") {
						toWrite += "\n"
					}
					_, err := p.stdinPipe.Write([]byte(toWrite))
					if err != nil {
						return
					}
				}
			}
		}()
	}

	if p.stdoutPipe != nil && p.stdoutChan != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			readStreamLines(p.stdoutPipe, p.Options.LineBufferSize, p.killChan, p.stdoutChan)
		}()
	}

	if p.stderrPipe != nil && p.stderrChan != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			readStreamLines(p.stderrPipe, p.Options.LineBufferSize, p.killChan, p.stderrChan)
		}()
	}

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
		if p.stdoutChan != nil {
			close(p.stdoutChan)
		}
		if p.stderrChan != nil {
			close(p.stderrChan)
		}
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
	p.closeStdinOnce.Do(func() {
		if p.stdinChan != nil {
			close(p.stdinChan)
		}
	})
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
		p.closeKillOnce.Do(func() { close(p.killChan) })
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
		p.closeKillOnce.Do(func() { close(p.killChan) })
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
