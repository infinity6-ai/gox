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

func supportGzip(req *httpzrequest.Req) bool {
	acceptEncodings := strings.Split(req.Headers.Get("Accept-Encoding"), ",")
	for _, acceptEncoding := range acceptEncodings {
		if strings.TrimSpace(acceptEncoding) == "gzip" {
			return true
		}
	}
	return false
}

func isCompressible(respHeaders http.Header) bool {
	ct := respHeaders.Get("Content-Type")
	return strings.Contains(ct, "text/") || strings.Contains(ct, "json") || strings.Contains(ct, "xml")
}

func shouldCompress(status int) bool {
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return false
	}
	if status >= 100 && status < 200 {
		return false
	}
	return true
}

func (f *filterImpl) filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
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

	if !supportGzip(req) {
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

		if !shouldCompress(nStatus) {
			return resp(nStatus, nHeaders)
		}

		if !isCompressible(nHeaders) {
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
