// Package httpzserverdisk provides a static file server handler for httpzserver,
// serving files directly from a disk directory using pathz.Path.
package httpzserverdisk

import (
	"context"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// Options configures the behavior of the disk static file handler.
type Options struct {
	// Base is the root directory path on disk from which to serve files. Must not be nil.
	Base *pathz.Path
	// Prefix is the URL path prefix to strip before looking up disk files (e.g. "/static" or "/assets").
	Prefix string
	// Index specifies the default file to serve for directory requests. Defaults to "index.html".
	Index string
	// NoIndex disables serving an index file when a directory or empty path is requested.
	NoIndex bool
	// Headers are additional HTTP headers added to all static responses (e.g., Cache-Control).
	Headers http.Header
}

func (o *Options) fix() {
	checker.NotNil(o.Base, "Base path cannot be nil")
	if !o.NoIndex && o.Index == "" {
		o.Index = "index.html"
	}
	if o.Index != "" {
		o.Index = strings.TrimPrefix(o.Index, "/")
	}
	if o.Prefix != "" {
		o.Prefix = strings.TrimSuffix(o.Prefix, "/")
		if !strings.HasPrefix(o.Prefix, "/") {
			o.Prefix = "/" + o.Prefix
		}
	}
}

// Handler returns a pattern handler that serves static files from disk.
// It is identical to DiskHandler.
func Handler(opts Options) httpzserver.HandlerPattern {
	return DiskHandler(opts)
}

// DiskHandler returns a pattern handler that serves static files from disk.
func DiskHandler(opts Options) httpzserver.HandlerPattern {
	opts.fix()

	var prefixPath *pathz.Path
	if opts.Prefix != "" {
		prefixPath = pathz.MustParse(opts.Prefix)
	}

	var indexPath *pathz.Path
	if opts.Index != "" {
		indexPath = pathz.MustParse(opts.Index)
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

		if err := relPath.Validate(pathz.ValidateOptions{
			MaxParents: new(0),
			Wildchar:   false,
		}); err != nil {
			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
		}

		targetPath, err := opts.Base.Join(relPath)
		if err != nil {
			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
		}

		fileInfo, err := os.Stat(targetPath.String())
		if err != nil {
			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
		}

		if fileInfo.IsDir() {
			if indexPath != nil {
				indexTarget, err := targetPath.Join(indexPath)
				if err != nil {
					w := resp(http.StatusNotFound, nil)
					_, err = w.Write([]byte("Not Found"))
					errorz.Check(err)
					return
				}
				idxInfo, err := os.Stat(indexTarget.String())
				if err != nil || idxInfo.IsDir() {
					w := resp(http.StatusNotFound, nil)
					_, err = w.Write([]byte("Not Found"))
					errorz.Check(err)
					return
				}
				targetPath = indexTarget
				fileInfo = idxInfo
			} else {
				w := resp(http.StatusNotFound, nil)
				_, err = w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}
		}

		f, err := os.Open(targetPath.String())
		if err != nil {
			w := resp(http.StatusInternalServerError, nil)
			_, err = w.Write([]byte("Internal Server Error"))
			errorz.Check(err)
			return
		}
		defer f.Close()

		headers := make(http.Header)
		if opts.Headers != nil {
			for k, v := range opts.Headers {
				headers[k] = v
			}
		}

		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", detectContentType(targetPath.String()))
		}
		if fileInfo.Size() >= 0 && headers.Get("Content-Length") == "" {
			headers.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
		}

		w := resp(http.StatusOK, headers)
		if req.Method == http.MethodHead {
			return
		}

		_, err = io.Copy(w, f)
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
