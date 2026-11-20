# brwjsz TODO

A lightweight Go-to-browser execution runtime and IPC bridge. Spawns a browser process hosting a local controller page (`controller.html`), connects it back to an internal `httpzserver`, and allows Go to dispatch JavaScript and manage remote object references (`Ref[T]`).

---

## 1. Architecture & Design Decisions

- [x] **Process Model & Browser Spawning**:
  - Launch browser via `execz` pointing to a single controller URL: `controller.html`.
  - Provide `(o *Options) Chrome(binPath string, headless bool, userDataDir string) *Options` for setting up Chrome/Chromium flags.
  - If `userDataDir == ""`, generate an isolated temporary user data directory and clean it up on `Browser.Close()`. If provided, use persistent path without deletion.
  - Set flags for sandbox / headless operation (`--headless`, `--ozone-platform=headless`, `--enable-logging=stderr`).

- [x] **Communication Transport**:
  - **Server-to-Browser (Push)**: SSE (Server-Sent Events) endpoint `GET /api/brwjs/events` for streaming command envelopes (`{ id, action, fn, ref, args, value }`).
  - **Browser-to-Server (Reply)**: `POST /api/brwjs/reply` for returning execution results (`{ id, ok, result, ref, error }`).
  - **Handshake / Ready**: Controller's SSE connection signals browser readiness; `Browser.Start()` blocks until the controller is connected.

- [x] **Remote Reference Registry (`Ref[T]`)**:
  - `type Ref[T any] string` is a lightweight string handle.
  - `MustCreate[T]() Ref[T]` is purely local (zero network RPC), generating a unique ID via `idgen.B64()`.
  - `MustSet[T](ctx, c, ref, v) io.Closer` sets the value in the browser and returns an `io.Closer` that deletes the reference on `Close()`.
  - `MustDelete[T](ctx, c, ref)` explicitly deletes the handle from the browser registry.
  - `MustGet[T](ctx, c, ref) T` retrieves the value from the browser registry.
  - In-browser registry (`window.__refs = new Map()`) stores JS objects, DOM nodes, functions, etc.
  - Serialization protocol for arguments: JSON primitives + reference envelope (`{"$ref": "ref_id"}`).

---

## 2. Implementation Tasks

### Phase 1: Controller & Static Assets
- [x] Fix paths in `brwjsz/internal/handlers/handlers.go`:
  - `BasePath = "/api/brwjs/f"`
  - `ControllerPath = BasePath + "/controller.html"`
- [x] Implement `brwjsz/staticz/brwjsz/index.js`:
  - Connect to SSE stream `/api/brwjs/events`.
  - Command dispatcher (`proc`, `eval`, `func`, `set`, `get`, `delete`).
  - Promise resolution: await returned Promises before sending reply.
  - Error capture with error responses back to Go via `/api/brwjs/reply`.
  - Reference store (`Map` of `id -> object`).
- [x] Regenerate embedded static asset bundle (`stzfiles.gen.go`).

### Phase 2: Server Endpoints & Dispatcher (`internal/handlers`)
- [x] Add SSE event handler (`GET /api/brwjs/events`):
  - Stream headers (`Content-Type: text/event-stream`, `Cache-Control: no-cache`).
  - Track controller connection and signal readiness.
  - Command queue with correlation IDs (`id`).
- [x] Add reply endpoint (`POST /api/brwjs/reply`):
  - Correlate reply with pending Go caller using channels.
  - Handle success payload and error payload.

### Phase 3: Browser Lifecycle (`brwjsz/brwjsz_server.go`)
- [x] Implement `(o *Options) Chrome(binPath string, headless bool, userDataDir string) *Options`:
  - Default `binPath` to `"google-chrome"` if empty (with fallback discovery for `chromium`/`chromium-browser`).
  - Configure flags for headless mode (`--headless`, `--ozone-platform=headless`).
  - Configure logging (`--enable-logging=stderr`).
  - Handle user profile: if `userDataDir == ""`, create isolated temporary user data dir and track in private `tempUserDataDir`; if set, use persistent path.
- [x] Update `Browser.Close()`:
  - Clean up `tempUserDataDir` if populated.
- [x] Update `fixArgs()`:
  - Inject single `controller.html` URL.
- [x] Update `Browser.Start()`:
  - Wait for controller to connect and complete handshake with timeout context.
- [x] Provide client accessor:
  - `func (b *Browser) Client() *brwclient.Client`.

### Phase 4: Client API (`brwclient/brwclient.go`)
- [x] Implement command dispatching with correlation IDs and timeouts.
- [x] Implement `MustCreate[T any]() Ref[T]` (local `idgen.B64()`).
- [x] Implement `MustSet[T any](ctx context.Context, c *Client, ref Ref[T], v T) io.Closer`.
- [x] Implement `MustDelete[T any](ctx context.Context, c *Client, ref Ref[T])`.
- [x] Implement `MustGet[T any](ctx context.Context, c *Client, ref Ref[T]) T`.
- [x] Implement `MustProc(ctx context.Context, c *Client, fn string, args ...any)`.
- [x] Implement `MustFunc[T any](ctx context.Context, c *Client, fn string, args ...any) Ref[T]`.
- [x] Implement `MustEval[T any](ctx context.Context, c *Client, expr string) T`.

### Phase 5: Testing & Verification
- [x] Unit tests for `internal/handlers` (mock SSE and replies without launching a real browser).
- [x] Unit tests for `brwclient` encoding/decoding and request-reply coordination.
- [x] E2E test with headless Chromium validating `MustProc`, `MustFunc`, `MustEval`, and `Ref[T]` lifecycle.
