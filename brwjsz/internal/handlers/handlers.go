package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/internal/stzfiles"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

const BasePath = "/api/brwjs/f"
const ControllerPath = BasePath + "/controller.html"
const EventsPath = "/api/brwjs/events"
const ReplyPath = "/api/brwjs/reply"

// AddHandlers registers the static file handler, SSE event stream, and reply endpoint on the server.
func AddHandlers(ctx context.Context, s *httpzserver.Server, hub *Hub) {
	// Debug log endpoint from JS
	s.AddHandler("GET", "/api/brwjs/log", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		msg := req.Query.Get("msg")
		fmt.Printf("[JS-LOG] %s\n", msg)
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("ok"))
	})

	// Serve static assets from embedded stzfiles
	s.AddHandler("GET", BasePath+"/*", httpzserverstatic.Handler(stzfiles.Name, httpzserverstatic.Options{
		Prefix:  BasePath,
		NoIndex: false,
	}))

	// SSE Events stream
	s.AddHandler("GET", EventsPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		tabStr := req.Query.Get("tab")
		tab, err := strconv.Atoi(tabStr)
		if err != nil {
			tab = 0
		}
		fmt.Printf("[SERVER] SSE connected for tab %d\n", tab)

		conn, cleanup := hub.RegisterTab(tab)
		defer func() {
			fmt.Printf("[SERVER] SSE disconnected for tab %d\n", tab)
			cleanup()
		}()

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
				fmt.Printf("[SERVER] Sending cmd to tab %d: %s\n", tab, string(data))
				_, err = fmt.Fprintf(w, "event: cmd\ndata: %s\n\n", string(data))
				if err != nil {
					fmt.Printf("[SERVER] Write cmd error to tab %d: %v\n", tab, err)
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

	// Reply endpoint
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
		fmt.Printf("[SERVER] Reply received: ID=%s, Ok=%v, Err=%s, Result=%s\n", reply.ID, reply.Ok, reply.Error, string(reply.Result))
		hub.HandleReply(&reply)
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("ok"))
	})
}

// AddHanlers is an alias for backwards compatibility.
func AddHanlers(ctx context.Context, s *httpzserver.Server) *Hub {
	h := NewHub()
	AddHandlers(ctx, s, h)
	return h
}
