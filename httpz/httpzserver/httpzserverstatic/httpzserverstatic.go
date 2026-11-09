// Package httpzserverstatic provides a static file server handler for httpzserver,
// utilizing the commonz/staticz package for static asset lookups.
package httpzserverstatic

import (
	"context"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// Options configures the behavior of the static file handler.
type Options struct {
	// Prefix is the URL path prefix to strip before looking up static files (e.g. "/static").
	Prefix string
	// Index specifies the default file to serve for directory requests. Defaults to "index.html".
	Index string
	// NoIndex disables serving an index file when a directory or empty path is requested.
	NoIndex bool
	// Headers are additional HTTP headers added to all static responses (e.g., Cache-Control).
	Headers http.Header
}

func (o *Options) fix() {
	if !o.NoIndex && o.Index == "" {
		o.Index = "index.html"
	}
}

// Handler returns a pattern handler that serves static files from the specified staticz pack.
// It is identical to StaticHandler.
func Handler(name any, opts ...Options) httpzserver.HandlerPattern {
	return StaticHandler(name, opts...)
}

// StaticHandler returns a pattern handler that serves static files from the specified staticz pack.
func StaticHandler(name any, opts ...Options) httpzserver.HandlerPattern {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.fix()

	var prefixPath *pathz.Path
	if opt.Prefix != "" {
		prefixPath = pathz.MustParse(opt.Prefix)
	}

	var indexPath *pathz.Path
	if opt.Index != "" {
		indexPath = pathz.MustParse(opt.Index)
	}

	return func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			w := resp(http.StatusMethodNotAllowed, nil)
			_, err := w.Write([]byte("Method Not Allowed"))
			errorz.Check(err)
			return
		}

		relPath := req.Path
		if prefixPath != nil {
			if !prefixPath.IsBaseOf(req.Path) {
				w := resp(http.StatusNotFound, nil)
				_, err := w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}
			var err error
			relPath, err = prefixPath.ExtractRelative(req.Path)
			if err != nil {
				w := resp(http.StatusNotFound, nil)
				_, err = w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}
		} else {
			relPath = relPath.ForceCurrent()
		}

		if relPath.PartsLen() == 0 || relPath.HasEndingSlash() {
			if indexPath != nil {
				if relPath.PartsLen() == 0 {
					relPath = indexPath
				} else {
					var err error
					relPath, err = relPath.Join(indexPath)
					if err != nil {
						w := resp(http.StatusNotFound, nil)
						_, err = w.Write([]byte("Not Found"))
						errorz.Check(err)
						return
					}
				}
			} else {
				w := resp(http.StatusNotFound, nil)
				_, err := w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}
		}

		entry, err := staticz.Lookup(ctx, name, relPath)
		if err != nil || entry == nil {
			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
		}

		r, err := entry.Open()
		if err != nil {
			w := resp(http.StatusInternalServerError, nil)
			_, err = w.Write([]byte("Internal Server Error"))
			errorz.Check(err)
			return
		}
		defer r.Close()

		headers := make(http.Header)
		if opt.Headers != nil {
			for k, v := range opt.Headers {
				headers[k] = v
			}
		}

		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", detectContentType(entry.Name().String()))
		}
		if entry.Size() >= 0 && headers.Get("Content-Length") == "" {
			headers.Set("Content-Length", strconv.FormatInt(entry.Size(), 10))
		}

		w := resp(http.StatusOK, headers)
		if req.Method == http.MethodHead {
			return
		}

		_, err = io.Copy(w, r)
		errorz.Check(err)
	}
}

func detectContentType(name string) string {
	ext := filepath.Ext(name)
	if ext != "" {
		ct := mime.TypeByExtension(ext)
		if ct != "" {
			return ct
		}
		switch ext {
		case ".js", ".mjs":
			return "text/javascript; charset=utf-8"
		case ".css":
			return "text/css; charset=utf-8"
		case ".json":
			return "application/json"
		case ".wasm":
			return "application/wasm"
		case ".svg":
			return "image/svg+xml"
		}
	}
	return "application/octet-stream"
}
