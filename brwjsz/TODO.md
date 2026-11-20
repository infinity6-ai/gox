# brwjsz TODO

A lightweight Go-to-browser execution runtime and IPC bridge. Spawns a browser process with multiple tabs hosting a local controller page (`controller.html?tab=N`), connects them back to an internal `httpzserver`, and allows Go to dispatch JavaScript and manage remote object references (`Ref[T]`).

---

## 1. Architecture & Design Decisions

- [ ] **Tab Model & Spawning**:
  - Launch browser via `execz` with multiple tab URLs on CLI: `controller.html?tab=0`, `controller.html?tab=1`, ...
  - Add default Chromium flags to disable background tab throttling (`--disable-background-timer-throttling`, `--disable-backgrounding-occluded-windows`, `--disable-renderer-backgrounding`).
  - Add flag to allow headless mode configuration in `Options`.

- [ ] **Communication Transport**:
  - **Server-to-Tab (Push)**: SSE (Server-Sent Events) endpoint `GET /api/brwjs/events?tab={id}` (or WebSocket) for streaming command envelopes.
  - **Tab-to-Server (Reply)**: `POST /api/brwjs/reply` for returning evaluation results, return values, and JS error reports.
  - **Tab Handshake / Ready**: Initial connection from each tab registers tab readiness; `Browser.Start()` blocks until all `0 .. TabCount-1` tabs report ready.

- [ ] **Remote Reference Registry (`Ref[T]`)**:
  - `type Ref[T any] string` remains a lightweight string handle.
  - `MustCreate[T]() Ref[T]` runs locally with zero network RPC, generating a unique ID via `idgen.B64()`.
  - `MustSet[T](ctx, c, ref, v) io.Closer` binds/updates the value in the browser and returns an `io.Closer` that automatically calls delete on `Close()` (can be safely ignored if persistent).
  - `MustDelete[T](ctx, c, ref)` explicitly deletes the handle from the browser registry (`window.__refs.delete(ref)`).
  - In-browser registry (`window.__refs = new Map()`) stores JS objects, DOM nodes, functions, etc.
  - Serialization protocol for arguments: JSON primitives + reference envelope (e.g., `{"$ref": "ref_id"}`).

---

## 2. Implementation Tasks

### Phase 1: Controller & Static Assets
- [ ] Fix `handlers.BasePath` and `ControllerPath` in `brwjsz/internal/handlers/handlers.go` (ensure correct URL path slashes).
- [ ] Implement `brwjsz/staticz/brwjsz/index.js`:
  - Tab ID discovery via `URLSearchParams(window.location.search).get('tab')`.
  - Connect to SSE command stream (`/api/brwjs/events?tab=...`).
  - Command dispatcher (eval, call function, create ref, set ref, delete ref).
  - Promise resolution: await returned Promises before sending result reply.
  - Error capture (`try/catch` and `window.onerror`) with structured error response back to Go.
  - Reference store (`Map` of `id -> object`).
- [ ] Update static asset bundling (`stzfiles`) if code generation is required.

### Phase 2: Server Endpoints & Dispatcher (`internal/handlers`)
- [ ] Add SSE event handler (`GET /api/brwjs/events?tab={id}`):
  - Stream headers (`Content-Type: text/event-stream`, `Cache-Control: no-cache`).
  - Track connected tabs and signal readiness.
  - Command queue per tab with correlation IDs (`request_id`).
- [ ] Add reply endpoint (`POST /api/brwjs/reply`):
  - Correlate response with pending Go caller using channels / promises.
  - Handle success payload and error payload.

### Phase 3: Browser Lifecycle & Tab Management (`brwjsz/brwjsz_server.go`)
- [ ] Implement `(o *Options) Chrome(binPath string, headless bool, userDataDir string) *Options`:
  - Default `binPath` to `"google-chrome"` if empty (with fallback discovery for `chromium`/`chromium-browser`).
  - Configure flags for headless mode (`--headless=new` or headed).
  - Add anti-throttling flags (`--disable-background-timer-throttling`, `--disable-backgrounding-occluded-windows`, `--disable-renderer-backgrounding`).
  - User profile handling: if `userDataDir == ""`, create isolated temporary user data dir, set `--user-data-dir=...`, and record it in private `tempUserDataDir` for deletion on close. If `userDataDir != ""`, use it and do NOT delete on close.
  - Disable first-run, sync prompts, and background updates.
- [ ] Extend `Options`:
  - `TabCount int` (defaults to 1).
  - `tempUserDataDir string` (internal tracked temporary profile for automatic deletion on close).
- [ ] Update `Browser.Close()`:
  - Clean up `tempUserDataDir` if populated.
- [ ] Update `fixArgs()`:
  - Generate $N$ tab URLs (`controller.html?tab=0`, ..., `tab=N-1`).
  - Inject required Chromium flags.
- [ ] Update `Browser.Start()`:
  - Wait for all tabs to connect and complete handshake with timeout context.
- [ ] Provide tab-scoped accessor:
  - `func (b *Browser) Tab(index int) *brwclient.Client` (or `b.Client().ForTab(index)`).

### Phase 4: Client API (`brwclient/brwclient.go`)
- [ ] Implement command dispatching with correlation IDs and timeouts.
- [ ] Implement `MustCreate[T any]() Ref[T]`:
  - Generate unique handle locally via `idgen.B64()` with zero network RPC.
- [ ] Implement `MustSet[T any](ctx context.Context, c *Client, ref Ref[T], v T) io.Closer`:
  - Send value to browser registry under `ref`, returning an `io.Closer` that deletes `ref` on `Close()`.
- [ ] Implement `MustDelete[T any](ctx context.Context, c *Client, ref Ref[T])`:
  - Instruct browser to delete handle from reference registry.
- [ ] Implement `MustProc(ctx context.Context, c *Client, fn string, args ...any)`:
  - Execute statement or function without expecting a return value.
- [ ] Implement `MustFunc[T any](ctx context.Context, c *Client, fn string, args ...any) Ref[T]`:
  - Execute function, store result in tab's reference registry with a generated `Ref[T]`, and return it.
- [ ] Implement direct evaluation / value retrieval:
  - `Eval[T any](ctx context.Context, c *Client, expr string) (T, error)`.

### Phase 5: Testing & Verification
- [ ] Unit tests for `internal/handlers` (mock SSE and replies without launching a real browser).
- [ ] Unit tests for `brwclient` encoding/decoding and request-reply coordination.
- [ ] Integration / E2E test with headless Chromium (or mock process) validating multi-tab execution, `MustProc`, `MustFunc`, and `Ref[T]` lifecycle.
