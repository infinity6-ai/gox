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
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
)

// HeaderContentEncodingWas indicates that the incoming request body was compressed
// and decompressed by the gzip filter.
const HeaderContentEncodingWas = "X-I6-Content-Encoding-Was"

// HeaderAcceptsGzip indicates whether the client supports gzipped responses.
const HeaderAcceptsGzip = "X-I6-Accepts-Gzip"

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
	if status == http.StatusNoContent || status == http.StatusNotModified || status == http.StatusPartialContent {
		return false
	}
	if status >= 100 && status < 200 {
		return false
	}
	if headers != nil {
		if headers.Get("Content-Encoding") != "" {
			return false
		}
		if headers.Get("Content-Range") != "" {
			return false
		}
		if strings.Contains(strings.ToLower(headers.Get("Cache-Control")), "no-transform") {
			return false
		}
		if isAlreadyCompressed(headers.Get("Content-Type")) {
			return false
		}
	}
	return true
}

// isAlreadyCompressed returns true if the media type represents data that is already
// compressed (such as PNG, JPEG, audio, video, ZIP, PDF, or WOFF fonts) and should not
// be re-compressed with gzip.
func isAlreadyCompressed(contentType string) bool {
	if contentType == "" {
		return false
	}
	ct := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.IndexByte(ct, ';'); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	if strings.HasPrefix(ct, "image/") {
		switch ct {
		case "image/svg+xml", "image/x-icon", "image/vnd.microsoft.icon", "image/bmp", "image/x-ms-bmp":
			return false
		default:
			return true
		}
	}
	if strings.HasPrefix(ct, "video/") || strings.HasPrefix(ct, "audio/") {
		return true
	}
	switch ct {
	case "application/zip",
		"application/x-zip-compressed",
		"application/gzip",
		"application/x-gzip",
		"application/x-bzip2",
		"application/x-xz",
		"application/x-7z-compressed",
		"application/x-rar-compressed",
		"application/vnd.rar",
		"application/zstd",
		"application/x-compress",
		"application/pdf",
		"font/woff",
		"font/woff2",
		"application/font-woff",
		"application/font-woff2":
		return true
	}
	return false
}

func GzipFilter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	if req.Headers == nil {
		req.Headers = make(http.Header)
	}
	req.Headers.Del(HeaderContentEncodingWas)
	req.Headers.Del(HeaderAcceptsGzip)

	origEncoding := req.Headers.Get("Content-Encoding")
	if req.Body != nil && req.Body != http.NoBody && isGzipEncoding(origEncoding) {
		gzReader, err := gzip.NewReader(req.Body)
		if err != nil {
			resp(http.StatusBadRequest, nil).Write([]byte("invalid gzip body"))
			return
		}
		defer gzReader.Close()
		req.Headers.Set(HeaderContentEncodingWas, origEncoding)
		req.Headers.Del("Content-Encoding")
		req.Headers.Del("Content-Length")
		req.Body = gzReader
		audit := httpzserverfilteraudit.GetEntry(ctx)
		if audit != nil {
			audit.ReqCompressed = true
			req.Body = audit.CreateReqBody(req.Body)
		}
	}

	clientAcceptsGzip := acceptsGzip(req)
	if clientAcceptsGzip {
		req.Headers.Set(HeaderAcceptsGzip, "true")
	} else {
		req.Headers.Set(HeaderAcceptsGzip, "false")
	}

	if !clientAcceptsGzip {
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

		gzWriter = gzip.NewWriter(outWriter)

		audit := httpzserverfilteraudit.GetEntry(ctx)
		if audit != nil {
			audit.RespCompressed = true
			return audit.CreateRespBody(gzWriter)
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
