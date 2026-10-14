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

func acceptsGzip(req *httpzrequest.Req) bool {
	if req == nil || req.Headers == nil {
		return false
	}
	acceptEncoding := req.Headers.Get("Accept-Encoding")
	if acceptEncoding == "" {
		return false
	}
	hasGzip := false
	hasWildcard := false
	for _, part := range strings.Split(acceptEncoding, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if strings.HasPrefix(part, "gzip") {
			if strings.Contains(part, ";q=0") {
				return false
			}
			hasGzip = true
		} else if strings.HasPrefix(part, "*") {
			if !strings.Contains(part, ";q=0") {
				hasWildcard = true
			}
		}
	}
	return hasGzip || hasWildcard
}

func shouldCompress(status int, headers http.Header) bool {
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return false
	}
	if status >= 100 && status < 200 {
		return false
	}
	if headers != nil && headers.Get("Content-Encoding") != "" {
		return false
	}
	return true
}

func (f *filterImpl) filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	if req.Headers == nil {
		req.Headers = make(http.Header)
	}
	req.Headers.Del(HeaderContentEncodingWas)

	origEncoding := req.Headers.Get("Content-Encoding")
	if req.Body != nil && req.Body != http.NoBody && isGzipEncoding(origEncoding) {
		gzReader, err := gzip.NewReader(req.Body)
		if err != nil {
			resp(http.StatusBadRequest, nil).Write([]byte("invalid gzip body"))
			return
		}
		defer gzReader.Close()
		req.Body = gzReader
		req.Headers.Set(HeaderContentEncodingWas, origEncoding)
		req.Headers.Del("Content-Encoding")
		req.Headers.Del("Content-Length")
	}

	if !acceptsGzip(req) {
		next(ctx, resp, req)
		return
	}

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

		if !shouldCompress(nStatus, nHeaders) {
			return resp(nStatus, nHeaders)
		}

		nHeaders.Set("Content-Encoding", "gzip")
		nHeaders.Del("Content-Length")
		vary := nHeaders.Get("Vary")
		if vary == "" {
			nHeaders.Set("Vary", "Accept-Encoding")
		} else if !strings.Contains(strings.ToLower(vary), "accept-encoding") {
			nHeaders.Set("Vary", vary+", Accept-Encoding")
		}

		outWriter := resp(nStatus, nHeaders)

		var err error
		if f.level != gzip.DefaultCompression {
			gzWriter, err = gzip.NewWriterLevel(outWriter, f.level)
		}
		if gzWriter == nil || err != nil {
			gzWriter = gzip.NewWriter(outWriter)
		}
		return gzWriter
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
