// Package execzv2 provides structured OS process execution with channel-based
// streaming for standard I/O (stdin, stdout, stderr), process group signaling,
// and graceful termination with fallback to SIGKILL.
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

// ErrKilled is returned when a process does not exit within the grace period deadline and is terminated via SIGKILL.
var ErrKilled = errors.New("ErrKilled")

// DefaultLineBufferSize is the default maximum byte size of a line fragment (64 KB).
const DefaultLineBufferSize = 64 * 1024

// Line represents a streamed input or output line.
// For output:
//   - Partial is false when the line was completed by a newline (\n or \r\n), with the newline stripped.
//   - Partial is true when the line reached LineBufferSize without a newline, or ended at EOF without a newline.
//
// For input:
//   - If Partial is false, a newline is automatically appended if not already present.
//   - If Partial is true, the line is written as-is without appending a newline.
type Line struct {
	Line    string
	Partial bool
}

// Options configures the creation and I/O behavior of a Process.
type Options struct {
	// Bin specifies the path or name of the executable binary.
	Bin string
	// Args holds command-line arguments passed to the binary.
	Args []string
	// Envs provides environment variables to set or override (merged with os.Environ()).
	Envs map[string]string
	// Dir sets the working directory of the process.
	Dir string
	// CaptureStdout defines the buffer size of the stdout channel.
	// When 0 or negative, stdout capture is disabled and output is discarded.
	CaptureStdout int
	// CaptureStderr defines the buffer size of the stderr channel.
	// When 0 or negative, stderr capture is disabled and output is discarded.
	CaptureStderr int
	// EnableStdin defines the buffer size of the stdin channel.
	// When 0 or negative, stdin is disabled.
	EnableStdin int
	// LineBufferSize sets the maximum byte size of a single line chunk before emitting a partial line.
	// Defaults to DefaultLineBufferSize (64 KB) when 0 or negative.
	LineBufferSize int
}

type stateType string

const (
	stateCreated  stateType = "created"
	stateStarted  stateType = "started"
	stateFinished stateType = "finished"
)

// Process manages the execution and streaming lifecycle of an OS process.
type Process struct {
	Options
	cmd *exec.Cmd

	// Err holds the process exit error or failure reason after termination.
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

// New creates an unstarted Process configured with the given Options.
func New(ctx context.Context, opts Options) *Process {
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

// GetStdout returns a receive-only channel streaming lines read from the process's standard output.
// Returns nil if CaptureStdout was not enabled (> 0).
func (p *Process) GetStdout() <-chan Line {
	return p.stdoutChan
}

// GetStderr returns a receive-only channel streaming lines read from the process's standard error.
// Returns nil if CaptureStderr was not enabled (> 0).
func (p *Process) GetStderr() <-chan Line {
	return p.stderrChan
}

// GetStdin returns a send-only channel for streaming lines into the process's standard input.
// Returns nil if EnableStdin was not enabled (> 0).
// Closing this channel or calling Close() sends EOF to the process's stdin.
func (p *Process) GetStdin() chan<- Line {
	return p.stdinChan
}

// readStreamLines reads bytes from r, splits into lines by '\n', enforces maxBuf chunking,
// and delivers Line structs to outChan until EOF or killChan is closed.
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

// Start begins execution of the process asynchronously.
// If ctx is cancelled before or during process execution, the process is terminated.
// Returns an error if the process has already been started or if startup fails.
func (p *Process) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.state != stateCreated {
		p.mu.Unlock()
		return fmt.Errorf("process already started")
	}
	if p.Err != nil {
		p.mu.Unlock()
		return fmt.Errorf("process initialization error: %w", p.Err)
	}
	if err := ctx.Err(); err != nil {
		p.mu.Unlock()
		return fmt.Errorf("context cancelled before starting process: %w", err)
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

	go func() {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			if p.cmd.Process != nil && p.state == stateStarted {
				_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
			}
			p.mu.Unlock()
		case <-p.waitChan:
		}
	}()

	p.startGoroutines()
	return nil
}

// Run starts the process and blocks until it completes or ctx is cancelled / deadline expires,
// returning any exit error.
func (p *Process) Run(ctx context.Context, deadline time.Duration) error {
	if deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(deadline))
		defer cancel()
	}

	if err := p.Start(ctx); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	if err := p.WaitFor(ctx, 0); err != nil {
		return fmt.Errorf("failed to wait for process: %w", err)
	}
	return nil
}

// WaitFor blocks until the process exits and all captured streams have closed,
// or until ctx is cancelled / deadline expires. It is safe to call concurrently and repeatedly.
func (p *Process) WaitFor(ctx context.Context, deadline time.Duration) error {
	if deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(deadline))
		defer cancel()
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for process: %w", ctx.Err())
	case <-p.waitChan:
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.Err != nil {
			return p.Err
		}
		return nil
	}
}

// IsRunning returns true if the process has started and has not yet terminated.
func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state == stateStarted
}

// Close closes the standard input stream idempotently, signaling EOF to the child process.
func (p *Process) Close() error {
	p.closeStdinOnce.Do(func() {
		if p.stdinChan != nil {
			close(p.stdinChan)
		}
	})
	return nil
}

// Kill attempts to gracefully stop the process group via SIGTERM. It creates a deadline context
// (if deadline > 0) after sending the initial signal. If the process does not terminate within the deadline
// or if ctx is cancelled, it forcefully terminates the group via SIGKILL and returns an error wrapping
// both ErrKilled and the context error (context.DeadlineExceeded or context.Canceled).
func (p *Process) Kill(ctx context.Context, deadline time.Duration) error {
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

	if deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(deadline))
		defer cancel()
	}

	done := make(chan error, 1)
	go func() {
		<-p.waitChan
		p.mu.Lock()
		err := p.Err
		p.mu.Unlock()
		done <- err
	}()

	select {
	case <-ctx.Done():
		logger.Info(ctx, "grace period expired or context cancelled, killing", map[string]any{"pid": pid})
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
		killErr := fmt.Errorf("Killed %d: %s [%s], %w: %w", pid, p.cmd.Path, p.cmd.Args, ErrKilled, ctx.Err())
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
