package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/internal/stzfiles"
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

// Site configures serving static files from a disk directory on the server.
type Site struct {
	Disk *pathz.Path
	Path *pathz.Path // default "/*"
}

// AddHandlers registers all internal handlers and an optional disk site on the server.
func AddHandlers(ctx context.Context, s *httpzserver.Server, hub *Hub, site *Site) {
	addStatic(s)
	addEvents(s, hub)
	addReply(s, hub)
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

func addSite(s *httpzserver.Server, site *Site) {
	if site == nil || site.Disk == nil {
		return
	}
	mountPath := "/*"
	prefix := ""
	if site.Path != nil {
		p := site.Path.String()
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		clean := strings.TrimSuffix(p, "/*")
		clean = strings.TrimSuffix(clean, "*")
		clean = strings.TrimSuffix(clean, "/")
		if clean == "" {
			mountPath = "/*"
			prefix = ""
		} else {
			mountPath = clean + "/*"
			prefix = clean
		}
	}
	s.AddHandler("GET", mountPath, httpzserverdisk.Handler(httpzserverdisk.Options{
		Base:    site.Disk,
		Prefix:  prefix,
		Listing: true,
	}))
}
