# `httpexecz` Design & Decision Plan

## Overview
`httpexecz` provides an `http.RoundTripper` implementation that delegates HTTP execution to an external CLI tool (by default `curl`). This allows transparent integration with standard `http.Client` and `httpzclient` while taking advantage of OS-level CLI tools (e.g., custom network namespaces, TLS fingerprint spoofing via `curl-impersonate`, SPNEGO/Kerberos authentication, or complex SOCKS/VPN routing).

---

## Key Decisions to Make

### 1. Package Location & Dependencies
* **Option A (`commonz/httpexecz`):**
  * Sits in `commonz`, depending only on Go standard library (`net/http`, `os/exec`, `bufio`, `io`) and `commonz` error handling.
  * Ensures zero cyclical dependencies with higher-level packages.
* **Option B (`httpz/httpzexec`):**
  * Sits in `httpz`, directly integrating with `httpzclient.Filter` or `httpzclient.Options`.
* **Recommendation:** `commonz/httpexecz` providing `http.RoundTripper` and `NewClient(opts) *http.Client`. `httpzclient` can then consume it without dependency cycles.

---

### 2. Command Abstraction: Hardcoded `curl` vs CLI Driver Interface
* **Option A (Fixed `curl` generator):**
  * Struct configuration specifies `Command` (defaults to `"curl"`) and extra arguments. Argument generation logic is hardcoded for `curl` conventions (`-s`, `-i`, `-X`, `--data-binary`, etc.).
* **Option B (Pluggable `Driver` interface):**
  * Provide a driver interface:
    ```go
    type Driver interface {
        BuildCommand(ctx context.Context, req *http.Request) (*exec.Cmd, error)
    }
    ```
  * Supply `CurlDriver` as the default implementation.
* **Questions:**
  * Do we only need `curl` and `curl-impersonate` (which share identical CLI flags)?
  * If yes, a dedicated `CurlDriver` / `Options` struct is simpler and less indirect.

---

### 3. Argument Construction & Security (/proc inspection)
* **Risk:** Passing sensitive headers (e.g., `Authorization: Bearer <secret>`, `Cookie: ...`) via command-line arguments (`-H "Authorization: ..."`) makes them readable by any local user via `/proc/<pid>/cmdline` or `ps aux`.
* **Options:**
  1. **Argv-only (`-H` flags):** Simple, standard curl usage. Safe in isolated container/sandbox environments, unsafe on multi-tenant hosts.
  2. **Config file via stdin (`--config -`):** `curl` supports reading config files from stdin via `--config -`. However, this conflicts with streaming the HTTP request body through stdin (`--data-binary @-`).
  3. **Temporary secure config files or named pipes (`mkfifo` / `/dev/fd/...`):** Allows writing sensitive headers out-of-band without leaking to argv.
* **Decision Needed:** Should we accept argv passing for simplicity by default, or provide an option to route headers safely via config files?

---

### 4. HTTP Version & Status Line Compatibility
* **Issue:** `curl -i` against HTTP/2 servers outputs `HTTP/2 200` without a minor version. Go's standard library `http.ReadResponse(bufio.NewReader(stdout), req)` fails on `HTTP/2 200` with:
  ```
  malformed HTTP version "HTTP/2"
  ```
* **Options:**
  1. **Enforce `--http1.1` in default curl flags:** Forces curl to negotiate HTTP/1.1. Simple and ensures standard compatibility with `http.ReadResponse`.
  2. **Status-line normalizer wrapper:** Read the first line from stdout. If it begins with `HTTP/2`, rewrite it into `HTTP/1.1` (or prepend `HTTP/1.1`) before handing the stream over to `http.ReadResponse`.
* **Decision Needed:** Should `--http1.1` be forced by default, or should we support HTTP/2 responses via line normalization?

---

### 5. Request & Response Body Streaming vs Process Lifecycle
* **Request Body (stdin):**
  * When `req.Body != nil` and `req.ContentLength != 0`:
    * Pipe `req.Body` directly to `cmd.Stdin`.
    * Supply `--data-binary @-` to `curl`.
  * For bodyless requests (`GET`, `HEAD`):
    * Do not attach stdin or pass body flags; otherwise curl may hang waiting for EOF or convert the request to `POST`.
* **Response Body (stdout) & Process Cleanup:**
  * Parse response status & headers using `http.ReadResponse(bufio.NewReader(stdoutPipe), req)`.
  * The returned `resp.Body` must wrap `stdoutPipe` and ensure `cmd.Wait()` is executed when `resp.Body.Close()` is called.
  * **Failure mode:** If client code does not close `resp.Body`, the process could leak as a zombie until context cancellation or garbage collection.
  * Ensure context cancellation (`exec.CommandContext`) terminates the process group via SIGKILL / SIGTERM.

---

### 6. Redirect Handling Strategy
* **Standard `http.RoundTripper` Contract:**
  * A `RoundTripper` must execute a single HTTP transaction and return the exact response received, including `3xx` redirect responses, leaving redirect policy to `http.Client.CheckRedirect`.
* **Decision:**
  * Always omit `-L`.
  * Curl returns the raw 3xx redirect response to `RoundTrip`.
  * If the caller uses `http.Client`, Go's standard redirect logic (or custom `CheckRedirect`) manages following redirects as desired.

---

### 7. Stderr Capture & Diagnostic Errors
* When curl fails (e.g. connection refused, DNS resolution failed, SSL handshake failure):
  * `stdout` is empty; `http.ReadResponse` returns `io.EOF` or an unexpected EOF error.
  * `curl` prints human-readable errors to `stderr` and exits with a non-zero code.
* **Design:**
  * Buffer `stderr` in memory (`bytes.Buffer`).
  * If `cmd.Wait()` returns an exit error or `ReadResponse` fails, combine the exit code and stderr output into the returned error:
    ```go
    fmt.Errorf("httpexecz: curl exited with status %d: %s: %w", exitCode, stderrStr, err)
    ```

---

### 8. Proposed API Definition

```go
package httpexecz

import (
	"context"
	"net/http"
	"time"
)

// Options configures the exec-based HTTP Transport.
type Options struct {
	// Command specifies the binary to execute (defaults to "curl").
	Command string

	// ExtraArgs specifies additional CLI arguments to prepend or append to curl.
	ExtraArgs []string

	// ForceHTTP11 forces curl to negotiate HTTP/1.1 (defaults to true).
	ForceHTTP11 bool

	// Timeout specifies a process execution timeout if request context has no deadline.
	Timeout time.Duration

	// Env specifies additional environment variables for the command execution.
	Env []string

	// Trace specifies an optional directory path where wire trace logs will be dumped
	// via curl's --trace flag using the pattern dump-[ts]-[host]-[path]-[rnd].log.
	Trace string

	// OutScript specifies an optional directory path where a runnable shell script (.sh)
	// reproducing the curl command (split across lines with \) will be written.
	// If the request has a body, the payload is also saved alongside it as .payload.
	OutScript string

	// BeforeExec optionally inspects or mutates the executable binary and arguments
	// before process execution. If it returns an error, the round-trip aborts immediately.
	BeforeExec func(cmd string, args []string) (string, []string, error)
}

// Transport implements http.RoundTripper using an external process.
type Transport struct {
	opts Options
}

// NewTransport creates an initialized Transport with default options.
func NewTransport(opts Options) *Transport

// RoundTrip executes a single HTTP transaction via process execution.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error)

// NewClient returns an *http.Client configured with this Transport.
func NewClient(opts Options) *http.Client
```

---

## Verification & Testing Strategy
1. **Mock HTTP Server Tests (`httptest.Server`):**
   * Standard `GET`, `POST`, `PUT`, `DELETE` with payload validation.
   * Header forwarding (request headers -> curl `-H` -> server -> response headers).
   * Status code propagation (`200`, `201`, `400`, `404`, `500`).
   * Large body streaming (input and output).
2. **Context & Cancellation Tests:**
   * Aborted request context verifies the subprocess is killed immediately without hanging.
3. **Error Reporting Tests:**
   * Connection to closed port or invalid host verifies stderr capture and exit code error wrapping.
