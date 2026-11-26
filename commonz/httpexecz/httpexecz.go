// Package httpexecz provides an http.RoundTripper implementation that executes
// HTTP transactions via an external command-line process (defaulting to curl).
package httpexecz

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/infinity6-ai/gox/commonz/datez"
)

// Options configures an exec-based HTTP Transport.
type Options struct {
	// Command specifies the executable binary to run (defaults to "curl").
	Command string

	// ExtraArgs specifies additional command-line arguments passed to the command.
	ExtraArgs []string

	// DisableHTTP11 specifies whether to disable forcing HTTP/1.1.
	// When false (default), --http1.1 is passed to curl to guarantee compatibility
	// with net/http.ReadResponse.
	DisableHTTP11 bool

	// Timeout specifies an optional process execution timeout if the request context has no deadline.
	Timeout time.Duration

	// Env specifies additional environment variables for the command execution.
	Env []string

	// Trace specifies an optional directory path where wire trace logs will be written.
	// When non-empty, curl is invoked with --trace <dir>/dump-[ts]-[host]-[path]-[rnd].log.
	// The directory is created automatically if it does not exist.
	Trace string

	// OutScript specifies an optional directory path where a runnable shell script (.sh)
	// reproducing the curl command (split across lines with \) will be written.
	// If the request has a body, the payload is also saved alongside it as .payload.
	// The directory is created automatically if it does not exist.
	OutScript string

	// BeforeExec optionally inspects or mutates the executable binary and arguments
	// before process execution. If it returns an error, the round-trip aborts immediately.
	BeforeExec func(cmd string, args []string) (string, []string, error)
}

func (o *Options) fix() {
	if o.Command == "" {
		o.Command = "curl"
	}
}

// Transport implements http.RoundTripper by executing an external CLI process.
type Transport struct {
	opts Options
}

// NewTransport returns a new Transport configured with opts.
func NewTransport(opts Options) *Transport {
	opts.fix()
	return &Transport{
		opts: opts,
	}
}

// NewClient returns a new *http.Client configured with an exec-based Transport.
func NewClient(opts Options) *http.Client {
	return &http.Client{
		Transport: NewTransport(opts),
	}
}

// RoundTrip executes a single HTTP transaction via process execution.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL == nil {
		return nil, fmt.Errorf("httpexecz: request URL is nil")
	}
	if req.URL.String() == "" {
		return nil, fmt.Errorf("httpexecz: request URL is empty")
	}

	hasBody := req.Body != nil && req.Body != http.NoBody

	baseName := generateDumpBaseName(req)

	args, tracePath, err := t.buildArgs(req, hasBody, baseName)
	if err != nil {
		return nil, fmt.Errorf("httpexecz: failed to build command arguments: %w", err)
	}

	command := t.opts.Command
	if t.opts.BeforeExec != nil {
		command, args, err = t.opts.BeforeExec(command, args)
		if err != nil {
			return nil, fmt.Errorf("httpexecz: BeforeExec failed: %w", err)
		}
	}

	var payloadPath string
	if t.opts.OutScript != "" {
		if err := os.MkdirAll(t.opts.OutScript, 0755); err != nil {
			return nil, fmt.Errorf("httpexecz: failed to create out-script directory: %w", err)
		}
		scriptPath := filepath.Join(t.opts.OutScript, baseName+".sh")
		if hasBody {
			payloadPath = filepath.Join(t.opts.OutScript, baseName+".payload")
		}
		scriptContent := generateScriptContent(command, args, hasBody, baseName)
		if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			return nil, fmt.Errorf("httpexecz: failed to write out-script file: %w", err)
		}
	}

	ctx := req.Context()
	var cancel context.CancelFunc
	if t.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, t.opts.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	cmd := exec.CommandContext(ctx, command, args...)
	if len(t.opts.Env) > 0 {
		cmd.Env = append(os.Environ(), t.opts.Env...)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("httpexecz: failed to create stdout pipe: %w", err)
	}

	var stderrBuf safeBuffer
	cmd.Stderr = &stderrBuf

	if hasBody {
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			_ = stdoutPipe.Close()
			cancel()
			return nil, fmt.Errorf("httpexecz: failed to create stdin pipe: %w", err)
		}

		var payloadWriter io.WriteCloser
		if payloadPath != "" {
			pf, err := os.Create(payloadPath)
			if err != nil {
				_ = stdoutPipe.Close()
				cancel()
				return nil, fmt.Errorf("httpexecz: failed to create payload file: %w", err)
			}
			payloadWriter = pf
		}

		go func() {
			defer stdinPipe.Close()
			defer req.Body.Close()
			if payloadWriter != nil {
				defer payloadWriter.Close()
				tee := io.TeeReader(req.Body, payloadWriter)
				_, _ = io.Copy(stdinPipe, tee)
			} else {
				_, _ = io.Copy(stdinPipe, req.Body)
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		_ = stdoutPipe.Close()
		cancel()
		return nil, fmt.Errorf("httpexecz: failed to start command %q: %w", command, err)
	}

	bufReader := bufio.NewReader(stdoutPipe)
	var resp *http.Response
	for {
		var readErr error
		resp, readErr = http.ReadResponse(bufReader, req)
		if readErr != nil {
			_ = stdoutPipe.Close()
			waitErr := cmd.Wait()
			cancel()

			stderrStr := strings.TrimSpace(stderrBuf.String())
			var traceInfo string
			if tracePath != "" {
				traceInfo = fmt.Sprintf(" (trace: %s)", tracePath)
			}
			if stderrStr != "" {
				return nil, fmt.Errorf("httpexecz: %s: %s (exit: %v)%s: %w", command, stderrStr, waitErr, traceInfo, readErr)
			}
			if waitErr != nil {
				return nil, fmt.Errorf("httpexecz: %s failed (%v)%s: %w", command, waitErr, traceInfo, readErr)
			}
			return nil, fmt.Errorf("httpexecz: failed to read HTTP response from %s%s: %w", command, traceInfo, readErr)
		}
		if resp.StatusCode < 200 {
			// Skip 1xx informational responses (e.g. 100 Continue)
			continue
		}
		break
	}

	resp.Body = &responseBody{
		reader:    resp.Body,
		stdout:    stdoutPipe,
		cmd:       cmd,
		stderr:    &stderrBuf,
		tracePath: tracePath,
		cancel:    cancel,
	}

	return resp, nil
}

func (t *Transport) buildArgs(req *http.Request, hasBody bool, baseName string) ([]string, string, error) {
	var args []string

	args = append(args, "-s")

	if !t.opts.DisableHTTP11 {
		args = append(args, "--http1.1")
	}

	method := strings.ToUpper(req.Method)
	if method == "" {
		method = http.MethodGet
	}

	if method == http.MethodHead {
		args = append(args, "-I")
	} else {
		args = append(args, "-i")
		if method != http.MethodGet {
			args = append(args, "-X", method)
		}
	}

	if req.Host != "" && (req.URL == nil || req.Host != req.URL.Host) {
		args = append(args, "-H", fmt.Sprintf("Host: %s", req.Host))
	}

	var headerKeys []string
	for k := range req.Header {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	for _, k := range headerKeys {
		if strings.EqualFold(k, "Host") && req.Host != "" {
			continue
		}
		for _, v := range req.Header[k] {
			args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
		}
	}

	if req.Close && req.Header.Get("Connection") == "" {
		args = append(args, "-H", "Connection: close")
	}

	if hasBody {
		args = append(args, "--data-binary", "@-")
	}

	if len(t.opts.ExtraArgs) > 0 {
		args = append(args, t.opts.ExtraArgs...)
	}

	var tracePath string
	if t.opts.Trace != "" {
		if err := os.MkdirAll(t.opts.Trace, 0755); err != nil {
			return nil, "", fmt.Errorf("httpexecz: failed to create trace directory: %w", err)
		}
		tracePath = filepath.Join(t.opts.Trace, baseName+".log")
		args = append(args, "--trace", tracePath)
	}

	args = append(args, req.URL.String())

	return args, tracePath, nil
}

func sanitizePrefix(s string, maxLen int) string {
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	res := strings.Trim(sb.String(), "_")
	if res == "" {
		return "none"
	}
	if len(res) > maxLen {
		res = res[:maxLen]
	}
	return strings.TrimRight(res, "_-")
}

func generateDumpBaseName(req *http.Request) string {
	ts := datez.TimeFormatTZ(time.Now())

	host := "nohost"
	if req.URL != nil && req.URL.Hostname() != "" {
		host = sanitizePrefix(req.URL.Hostname(), 24)
	}

	path := "root"
	if req.URL != nil && req.URL.Path != "" {
		p := strings.TrimPrefix(req.URL.Path, "/")
		if p != "" {
			path = sanitizePrefix(p, 24)
		}
	}

	b := make([]byte, 2)
	_, _ = rand.Read(b)
	rnd := hex.EncodeToString(b)

	return fmt.Sprintf("dump-%s-%s-%s-%s", ts, host, path, rnd)
}

func quoteBashArg(arg string) string {
	if arg == "" {
		return `""`
	}
	safe := true
	for _, r := range arg {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r == '=' || r == '@') {
			safe = false
			break
		}
	}
	if safe {
		return arg
	}
	escaped := strings.ReplaceAll(arg, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, `$`, `\$`)
	escaped = strings.ReplaceAll(escaped, "`", "\\`")
	return `"` + escaped + `"`
}

func generateScriptContent(command string, args []string, hasBody bool, baseName string) string {
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("set -e\n\n")
	if hasBody {
		sb.WriteString("DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"\n\n")
	}

	sb.WriteString(quoteBashArg(command))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if hasBody && arg == "--data-binary" && i+1 < len(args) && args[i+1] == "@-" {
			sb.WriteString(" \\\n  --data-binary @\"$DIR/" + baseName + ".payload\"")
			i++
			continue
		}
		if (arg == "-X" || arg == "-H" || arg == "--trace" || arg == "--retry") && i+1 < len(args) {
			sb.WriteString(" \\\n  " + arg + " " + quoteBashArg(args[i+1]))
			i++
			continue
		}
		sb.WriteString(" \\\n  " + quoteBashArg(arg))
	}
	sb.WriteString("\n")
	return sb.String()
}

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

type responseBody struct {
	reader    io.ReadCloser
	stdout    io.Closer
	cmd       *exec.Cmd
	stderr    *safeBuffer
	tracePath string
	cancel    context.CancelFunc
	closeOnce sync.Once
	closeErr  error
}

func (b *responseBody) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *responseBody) Close() error {
	b.closeOnce.Do(func() {
		var readErr error
		if b.reader != nil {
			readErr = b.reader.Close()
		}
		var stdoutErr error
		if b.stdout != nil {
			stdoutErr = b.stdout.Close()
		}

		waitErr := b.cmd.Wait()

		if b.cancel != nil {
			b.cancel()
		}

		if waitErr != nil {
			stderrStr := strings.TrimSpace(b.stderr.String())
			var traceInfo string
			if b.tracePath != "" {
				traceInfo = fmt.Sprintf(" (trace: %s)", b.tracePath)
			}
			if stderrStr != "" {
				b.closeErr = fmt.Errorf("httpexecz: command %s failed%s: %w: %s", b.cmd.Path, traceInfo, waitErr, stderrStr)
			} else {
				b.closeErr = fmt.Errorf("httpexecz: command %s failed%s: %w", b.cmd.Path, traceInfo, waitErr)
			}
			return
		}

		if readErr != nil {
			b.closeErr = fmt.Errorf("httpexecz: failed to close response reader: %w", readErr)
			return
		}
		if stdoutErr != nil {
			b.closeErr = fmt.Errorf("httpexecz: failed to close stdout pipe: %w", stdoutErr)
			return
		}
	})
	return b.closeErr
}
