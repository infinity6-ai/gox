package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverdisk"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

const BasePath = "/api/brwjs/f"
const ControllerPath = BasePath + "/controller.html"
const EventsPath = "/api/brwjs/events"
const ReplyPath = "/api/brwjs/reply"
const ExecPath = "/api/brwjs/exec"
const StopPath = "/api/brwjs/stop"

// Site configures serving static files from a disk directory on the server.
type Site struct {
	Disk *pathz.Path
	Path string // default "/*"
}

// AddHandlers registers all internal handlers and an optional disk site on the server.
func AddHandlers(ctx context.Context, s *httpzserver.Server, hub *Hub, site *Site) {
	addStatic(s)
	addEvents(s, hub)
	addReply(s, hub)
	addExec(s, hub)
	addStop(s, hub)
	addSite(s, site)
}

func addStatic(s *httpzserver.Server) {
	s.AddHandler("GET", BasePath+"/*", httpzserverstatic.Handler(stzfiles.Name, httpzserverstatic.Options{
		Prefix:  BasePath,
		NoIndex: false,
	}))
}

func addEvents(s *httpzserver.Server, hub *Hub) {
	s.AddHandler("GET", EventsPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		conn, cleanup := hub.RegisterController()
		defer cleanup()

		headers := make(http.Header)
		headers.Set("Content-Type", "text/event-stream")
		headers.Set("Cache-Control", "no-cache")
		headers.Set("Connection", "keep-alive")
		headers.Set("Access-Control-Allow-Origin", "*")

		w := resp(http.StatusOK, headers)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		for {
			select {
			case cmd, ok := <-conn.cmdChan:
				if !ok {
					return
				}
				data, err := json.Marshal(cmd)
				if err != nil {
					continue
				}
				_, err = fmt.Fprintf(w, "event: cmd\ndata: %s\n\n", string(data))
				if err != nil {
					return
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			case <-ctx.Done():
				return
			}
		}
	})
}

func addReply(s *httpzserver.Server, hub *Hub) {
	s.AddHandler("POST", ReplyPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		if req.Body == nil {
			w := resp(http.StatusBadRequest, nil)
			_, _ = w.Write([]byte("missing body"))
			return
		}
		var reply brwclient.Reply
		if err := json.NewDecoder(req.Body).Decode(&reply); err != nil {
			w := resp(http.StatusBadRequest, nil)
			_, _ = w.Write([]byte(fmt.Sprintf("invalid json: %s", err)))
			return
		}
		hub.HandleReply(&reply)
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("ok"))
	})
}

// ExecRequest represents the JSON payload sent to /api/brwjs/exec.
type ExecRequest struct {
	Code string `json:"code"`
	Args []any  `json:"args,omitempty"`
}

// ExecResponse represents the JSON response returned by /api/brwjs/exec.
type ExecResponse struct {
	Ok     bool   `json:"ok"`
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func addExec(s *httpzserver.Server, hub *Hub) {
	s.AddHandler("POST", ExecPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		headers := make(http.Header)
		headers.Set("Content-Type", "application/json")

		if req.Body == nil {
			w := resp(http.StatusBadRequest, headers)
			_ = json.NewEncoder(w).Encode(ExecResponse{Ok: false, Error: "missing body"})
			return
		}
		var execReq ExecRequest
		if err := json.NewDecoder(req.Body).Decode(&execReq); err != nil {
			w := resp(http.StatusBadRequest, headers)
			_ = json.NewEncoder(w).Encode(ExecResponse{Ok: false, Error: fmt.Sprintf("invalid json: %s", err)})
			return
		}

		wrapper := `async (code, args) => {
			window.args = args;
			let target;
			try {
				target = (0, eval)('(' + code + ')');
			} catch (_) {
				target = (0, eval)(code);
			}
			if (typeof target === 'function') {
				let res = target.apply(window, args);
				if (res && typeof res.then === 'function') {
					res = await res;
				}
				return res;
			}
			if (target && typeof target.then === 'function') {
				target = await target;
			}
			return target;
		}`

		cmd := &brwclient.Command{
			ID:     idgen.B64(),
			Action: "call",
			Fn:     wrapper,
			Args:   []any{execReq.Code, execReq.Args},
		}

		reply, err := hub.Dispatch(ctx, cmd)
		if err != nil {
			w := resp(http.StatusInternalServerError, headers)
			_ = json.NewEncoder(w).Encode(ExecResponse{Ok: false, Error: err.Error()})
			return
		}
		if !reply.Ok {
			w := resp(http.StatusOK, headers)
			_ = json.NewEncoder(w).Encode(ExecResponse{Ok: false, Error: reply.Error})
			return
		}

		var resVal any
		if len(reply.Result) > 0 {
			_ = json.Unmarshal(reply.Result, &resVal)
		}
		w := resp(http.StatusOK, headers)
		_ = json.NewEncoder(w).Encode(ExecResponse{Ok: true, Result: resVal})
	})
}

func addStop(s *httpzserver.Server, hub *Hub) {
	s.AddHandler("POST", StopPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		headers := make(http.Header)
		headers.Set("Content-Type", "application/json")
		w := resp(http.StatusOK, headers)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		go func() {
			time.Sleep(50 * time.Millisecond)
			hub.RequestStop()
		}()
	})
}

func addSite(s *httpzserver.Server, site *Site) {
	if site == nil || site.Disk == nil {
		return
	}
	path := site.Path
	if path == "" {
		path = "/*"
	}
	prefix := strings.TrimSuffix(path, "/*")
	prefix = strings.TrimSuffix(prefix, "*")
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	s.AddHandler("GET", path, httpzserverdisk.Handler(httpzserverdisk.Options{
		Base:    site.Disk,
		Prefix:  prefix,
		Listing: true,
	}))
}
