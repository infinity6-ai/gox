// Package httpzserverdisk provides a static file server handler for httpzserver,
// serving files directly from a disk directory using pathz.Path.
package httpzserverdisk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
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
	// Listing enables generating a directory listing (HTML or JSON) if an index file is not served or present.
	Listing bool
	// Headers are additional HTTP headers added to all static responses (e.g., Cache-Control).
	Headers http.Header
}

// ListingEntry represents a single file or directory item in a listing response.
type ListingEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
	Dir  bool   `json:"dir"`
}

// ListingResponse represents the JSON response structure for directory listings.
type ListingResponse struct {
	Path      string         `json:"path"`
	Recursive bool           `json:"recursive"`
	Entries   []ListingEntry `json:"entries"`
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

		explicitListing := opts.Listing && (req.Query.Has("t") || req.Query.Has("r"))

		if relPath.PartsLen() == 0 || relPath.HasEndingSlash() {
			if err := relPath.Validate(pathz.ValidateOptions{
				MaxParents: new(0),
				Wildchar:   false,
			}); err != nil {
				w := resp(http.StatusNotFound, nil)
				_, err = w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}

			dirTarget, err := opts.Base.Join(relPath)
			if err != nil {
				w := resp(http.StatusNotFound, nil)
				_, err = w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}

			dirInfo, err := os.Stat(dirTarget.String())
			if err != nil || !dirInfo.IsDir() {
				w := resp(http.StatusNotFound, nil)
				_, err = w.Write([]byte("Not Found"))
				errorz.Check(err)
				return
			}

			if !opts.NoIndex && indexPath != nil && !explicitListing {
				indexTarget, err := dirTarget.Join(indexPath)
				if err == nil {
					idxInfo, err := os.Stat(indexTarget.String())
					if err == nil && !idxInfo.IsDir() {
						serveFile(resp, req, indexTarget, idxInfo, opts)
						return
					}
				}
			}

			if opts.Listing {
				serveListing(resp, req, dirTarget, opts)
				return
			}

			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
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
			if !opts.NoIndex && indexPath != nil && !explicitListing {
				indexTarget, err := targetPath.Join(indexPath)
				if err == nil {
					idxInfo, err := os.Stat(indexTarget.String())
					if err == nil && !idxInfo.IsDir() {
						serveFile(resp, req, indexTarget, idxInfo, opts)
						return
					}
				}
			}

			if opts.Listing {
				serveListing(resp, req, targetPath, opts)
				return
			}

			w := resp(http.StatusNotFound, nil)
			_, err = w.Write([]byte("Not Found"))
			errorz.Check(err)
			return
		}

		serveFile(resp, req, targetPath, fileInfo, opts)
	}
}

func serveFile(resp httpzserver.Resp, req *httpzrequest.Req, targetPath *pathz.Path, fileInfo os.FileInfo, opts Options) {
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

func serveListing(resp httpzserver.Resp, req *httpzrequest.Req, dirTarget *pathz.Path, opts Options) {
	format := strings.ToLower(strings.TrimSpace(req.Query.Get("t")))
	if format == "" {
		format = "html"
	}
	if format != "html" && format != "json" {
		w := resp(http.StatusBadRequest, nil)
		_, err := w.Write([]byte("invalid format parameter 't', expected 'html' or 'json'"))
		errorz.Check(err)
		return
	}

	recursive := false
	if rStr := strings.TrimSpace(req.Query.Get("r")); rStr != "" {
		var err error
		recursive, err = strconv.ParseBool(rStr)
		if err != nil {
			w := resp(http.StatusBadRequest, nil)
			_, err = w.Write([]byte("invalid recursive parameter 'r', expected boolean"))
			errorz.Check(err)
			return
		}
	}

	dirURLPath := req.Path.String()
	if !strings.HasSuffix(dirURLPath, "/") {
		dirURLPath += "/"
	}

	var entries []ListingEntry
	targetStr := dirTarget.String()

	if recursive {
		err := filepath.WalkDir(targetStr, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == targetStr {
				return nil
			}
			rel, err := filepath.Rel(targetStr, path)
			if err != nil {
				return err
			}
			relSlash := filepath.ToSlash(rel)
			isDir := d.IsDir()
			urlPath := dirURLPath + relSlash
			if isDir {
				urlPath += "/"
			}
			var size int64
			if !isDir {
				if info, err := d.Info(); err == nil {
					size = info.Size()
				}
			}
			entries = append(entries, ListingEntry{
				Name: relSlash,
				Path: urlPath,
				Size: size,
				Dir:  isDir,
			})
			return nil
		})
		if err != nil {
			w := resp(http.StatusInternalServerError, nil)
			_, err = w.Write([]byte("Internal Server Error"))
			errorz.Check(err)
			return
		}
	} else {
		dirEntries, err := os.ReadDir(targetStr)
		if err != nil {
			w := resp(http.StatusInternalServerError, nil)
			_, err = w.Write([]byte("Internal Server Error"))
			errorz.Check(err)
			return
		}
		for _, entry := range dirEntries {
			name := entry.Name()
			isDir := entry.IsDir()
			urlPath := dirURLPath + name
			if isDir {
				urlPath += "/"
			}
			var size int64
			if !isDir {
				if info, err := entry.Info(); err == nil {
					size = info.Size()
				}
			}
			entries = append(entries, ListingEntry{
				Name: name,
				Path: urlPath,
				Size: size,
				Dir:  isDir,
			})
		}
	}

	if format == "json" {
		serveJSONListing(resp, req, dirURLPath, recursive, entries, opts)
	} else {
		serveHTMLListing(resp, req, dirURLPath, recursive, entries, opts)
	}
}

func serveJSONListing(resp httpzserver.Resp, req *httpzrequest.Req, dirURLPath string, recursive bool, entries []ListingEntry, opts Options) {
	if entries == nil {
		entries = []ListingEntry{}
	}
	responseObj := ListingResponse{
		Path:      dirURLPath,
		Recursive: recursive,
		Entries:   entries,
	}

	jsonBytes, err := json.MarshalIndent(responseObj, "", "  ")
	if err != nil {
		w := resp(http.StatusInternalServerError, nil)
		_, err = w.Write([]byte("Internal Server Error"))
		errorz.Check(err)
		return
	}

	headers := make(http.Header)
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			headers[k] = v
		}
	}
	headers.Set("Content-Type", "application/json; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

	w := resp(http.StatusOK, headers)
	if req.Method == http.MethodHead {
		return
	}

	_, err = w.Write(jsonBytes)
	errorz.Check(err)
}

func serveHTMLListing(resp httpzserver.Resp, req *httpzrequest.Req, dirURLPath string, recursive bool, entries []ListingEntry, opts Options) {
	var buf bytes.Buffer
	titleSuffix := ""
	if recursive {
		titleSuffix = " (recursive)"
	}
	escapedPath := html.EscapeString(dirURLPath) + titleSuffix

	buf.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	buf.WriteString("<meta charset=\"utf-8\">\n")
	buf.WriteString(fmt.Sprintf("<title>Index of %s</title>\n", escapedPath))
	buf.WriteString("<style>\n")
	buf.WriteString("body { font-family: monospace; padding: 20px; }\n")
	buf.WriteString("table { border-collapse: collapse; width: 100%; max-width: 800px; }\n")
	buf.WriteString("th, td { text-align: left; padding: 6px 12px; }\n")
	buf.WriteString("th { border-bottom: 2px solid #ccc; }\n")
	buf.WriteString("tr:hover { background-color: #f5f5f5; }\n")
	buf.WriteString("a { text-decoration: none; color: #0366d6; }\n")
	buf.WriteString("a:hover { text-decoration: underline; }\n")
	buf.WriteString(".size { text-align: right; }\n")
	buf.WriteString("</style>\n</head>\n<body>\n")
	buf.WriteString(fmt.Sprintf("<h1>Index of %s</h1>\n<hr>\n", escapedPath))
	buf.WriteString("<table>\n<thead>\n<tr><th>Name</th><th class=\"size\">Size</th><th>Type</th></tr>\n</thead>\n<tbody>\n")

	if req.Path.PartsLen() > 0 {
		parentURL := req.Path.Dir().String()
		if !strings.HasSuffix(parentURL, "/") {
			parentURL += "/"
		}
		buf.WriteString(fmt.Sprintf("<tr><td><a href=\"%s\">../</a></td><td class=\"size\">-</td><td>Directory</td></tr>\n", html.EscapeString(parentURL)))
	}

	for _, entry := range entries {
		displayName := entry.Name
		sizeStr := "-"
		typeStr := "Directory"
		if entry.Dir {
			displayName += "/"
		} else {
			typeStr = "File"
			sizeStr = formatSize(entry.Size)
		}
		buf.WriteString(fmt.Sprintf(
			"<tr><td><a href=\"%s\">%s</a></td><td class=\"size\">%s</td><td>%s</td></tr>\n",
			html.EscapeString(entry.Path),
			html.EscapeString(displayName),
			html.EscapeString(sizeStr),
			typeStr,
		))
	}

	buf.WriteString("</tbody>\n</table>\n<hr>\n</body>\n</html>\n")

	htmlBytes := buf.Bytes()
	headers := make(http.Header)
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			headers[k] = v
		}
	}
	headers.Set("Content-Type", "text/html; charset=utf-8")
	headers.Set("Content-Length", strconv.Itoa(len(htmlBytes)))

	w := resp(http.StatusOK, headers)
	if req.Method == http.MethodHead {
		return
	}

	_, err := w.Write(htmlBytes)
	errorz.Check(err)
}

func formatSize(size int64) string {
	const (
		b  = 1
		kb = 1024 * b
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case size >= gb:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(gb))
	case size >= mb:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(mb))
	case size >= kb:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(kb))
	default:
		return fmt.Sprintf("%d B", size)
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
