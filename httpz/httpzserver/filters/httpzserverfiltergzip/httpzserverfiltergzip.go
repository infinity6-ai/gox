// Package httpzserverfiltergzip provides HTTP gzip compression and decompression
// middleware for httpzserver, handling request decompression and response compression.
package httpzserverfiltergzip

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// HeaderContentEncodingWas indicates that the incoming request body was compressed
// and decompressed by the gzip filter.
const HeaderContentEncodingWas = "X-I6-Content-Encoding-Was"

// Options configures the gzip filter middleware.
type Options struct {
	// CompressionLevel specifies the gzip compression level
	// (e.g. gzip.DefaultCompression, gzip.BestSpeed, gzip.BestCompression).
	// If 0, gzip.DefaultCompression is used.
	CompressionLevel int
}

type filterImpl struct {
	level int
}

// New creates a new gzip filter middleware with the specified options.
func New(opts Options) httpzserver.Filter {
	level := opts.CompressionLevel
	if level == 0 {
		level = gzip.DefaultCompression
	}
	f := &filterImpl{
		level: level,
	}
	return f.filter
}

var defaultFilter = New(Options{})

// GzipFilter is the default gzip filter middleware that decompresses incoming
// gzip request bodies and compresses outgoing response bodies when requested by the client.
func GzipFilter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	defaultFilter(ctx, resp, req, next)
}

func (f *filterImpl) filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	// Sanitize any untrusted incoming X-I6-Content-Encoding-Was header for security
	req.Headers.Del(HeaderContentEncodingWas)

	// 1. Handle incoming request decompression if Content-Encoding is gzip
	origEncoding := req.Headers.Get("Content-Encoding")
	if isGzipEncoding(origEncoding) {
		if req.Body != nil && req.Body != http.NoBody {
			gzReader, err := gzip.NewReader(req.Body)
			if err != nil {
				w := resp(http.StatusBadRequest, nil)
				_, _ = w.Write([]byte("invalid gzip body"))
				return
			}
			req.Body = &gzipReadCloser{
				gzReader:   gzReader,
				underlying: req.Body,
			}
			req.Headers.Set(HeaderContentEncodingWas, origEncoding)
			req.Headers.Del("Content-Encoding")
			req.Headers.Del("Content-Length")
		}
	}

	// 2. Check if client accepts gzip for response compression
	clientAcceptsGzip := acceptsGzip(req.Headers.Get("Accept-Encoding"))

	var gzWriter *gzip.Writer
	defer func() {
		if gzWriter != nil {
			_ = gzWriter.Close()
		}
	}()

	nResp := func(nStatus int, nHeaders http.Header) io.Writer {
		if nHeaders == nil {
			nHeaders = make(http.Header)
		}

		shouldCompress := clientAcceptsGzip &&
			nStatus != http.StatusNoContent &&
			nStatus != http.StatusNotModified &&
			!(nStatus >= 100 && nStatus < 200) &&
			nHeaders.Get("Content-Encoding") == ""

		if shouldCompress {
			nHeaders.Set("Content-Encoding", "gzip")
			nHeaders.Del("Content-Length")
			vary := nHeaders.Get("Vary")
			if vary == "" {
				nHeaders.Set("Vary", "Accept-Encoding")
			} else if !strings.Contains(strings.ToLower(vary), "accept-encoding") {
				nHeaders.Set("Vary", vary+", Accept-Encoding")
			}
		}

		outWriter := resp(nStatus, nHeaders)

		if shouldCompress {
			var err error
			if f.level != gzip.DefaultCompression {
				gzWriter, err = gzip.NewWriterLevel(outWriter, f.level)
			}
			if gzWriter == nil || err != nil {
				gzWriter = gzip.NewWriter(outWriter)
			}
			return &gzipResponseWriter{
				gw: gzWriter,
				w:  outWriter,
			}
		}

		return outWriter
	}

	next(ctx, nResp, req)
}

func isGzipEncoding(contentEncoding string) bool {
	if contentEncoding == "" {
		return false
	}
	for _, part := range strings.Split(contentEncoding, ",") {
		if strings.TrimSpace(strings.ToLower(part)) == "gzip" {
			return true
		}
	}
	return false
}

func acceptsGzip(acceptEncoding string) bool {
	if acceptEncoding == "" {
		return false
	}
	for _, part := range strings.Split(acceptEncoding, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if strings.HasPrefix(part, "gzip") {
			if strings.Contains(part, ";q=0") {
				return false
			}
			return true
		}
		if part == "*" {
			return true
		}
	}
	return false
}

type gzipReadCloser struct {
	gzReader   *gzip.Reader
	underlying io.Reader
}

func (g *gzipReadCloser) Read(p []byte) (int, error) {
	return g.gzReader.Read(p)
}

func (g *gzipReadCloser) Close() error {
	err1 := g.gzReader.Close()
	var err2 error
	if c, ok := g.underlying.(io.Closer); ok {
		err2 = c.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

type gzipResponseWriter struct {
	gw *gzip.Writer
	w  io.Writer
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	return g.gw.Write(p)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.gw.Write([]byte(s))
}

func (g *gzipResponseWriter) Flush() {
	_ = g.gw.Flush()
	if f, ok := g.w.(http.Flusher); ok {
		f.Flush()
	}
}
