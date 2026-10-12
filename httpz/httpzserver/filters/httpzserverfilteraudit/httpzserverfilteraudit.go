// Package httpzserverfilteraudit provides audit logging middleware for httpzserver,
// capturing detailed HTTP request and response information including streaming bodies
// without loading entire payloads into memory.
package httpzserverfilteraudit

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

// Collector returns the log collector for capturing and asserting audit entries in tests.
func Collector() func() []*logzspec.Entry {
	return logger.(logz.Collector).Collector()
}

// Options configures the audit filter middleware.
type Options struct {
	// Service identifies the service name and version for the audit log entry.
	Service *logzaudit.AuditService
	// Logger is an optional custom audit logger. If nil, package Logger is used.
	Logger logz.AuditLogger
}

type filterImpl struct {
	service *logzaudit.AuditService
	logger  logz.AuditLogger
}

// New creates a new audit filter middleware with the specified options.
func New(opts Options) httpzserver.Filter {
	f := &filterImpl{
		service: opts.Service,
		logger:  opts.Logger,
	}
	return f.filter
}

var defaultFilter = New(Options{})

// AuditFilter is the default audit filter middleware that audits HTTP request and response
// pairs, logging them via the package logger without loading full body payloads into memory.
func AuditFilter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	defaultFilter(ctx, resp, req, next)
}

func (f *filterImpl) filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	reqCreatedAt := time.Now().UTC().UnixMilli()

	reqAuditor := &bodyAuditor{}
	if req.Body != nil && req.Body != http.NoBody {
		reqAuditor.hasBody = true
		req.Body = &auditReader{
			r:       req.Body,
			auditor: reqAuditor,
		}
	}

	var respStatus int
	var respHeaders http.Header
	var respCalled bool
	var respCreatedAt int64
	respAuditor := &bodyAuditor{}

	defer func() {
		// Drain any remaining unread bytes from req.Body without loading into memory.
		if req.Body != nil {
			var buf [4096]byte
			for {
				if ctx.Err() != nil {
					break
				}
				n, err := req.Body.Read(buf[:])
				if n == 0 || err != nil {
					break
				}
			}
		}

		var reqPath string
		if req.Path != nil {
			reqPath = req.Path.String()
		} else if req.Url != nil && req.Url.Path != nil {
			reqPath = req.Url.Path.String()
		}

		reqHost := req.Headers.Get("Host")
		if reqHost == "" && req.Url != nil {
			reqHost = req.Url.Host
		}

		var reqQuery url.Values
		if req.Query != nil {
			reqQuery = make(url.Values, len(req.Query))
			for k, v := range req.Query {
				reqQuery[k] = append([]string(nil), v...)
			}
		}

		var reqHeaders http.Header
		if req.Headers != nil {
			reqHeaders = req.Headers.Clone()
		}

		reqId := req.Headers.Get("X-Request-Id")
		if reqId == "" {
			reqId = idgen.String()
		}

		entry := &logzaudit.AuditEntryReqResp{
			Id:      reqId,
			Service: f.service,
			Req: &logzaudit.AuditRequest{
				CreatedAt: reqCreatedAt,
				Host:      reqHost,
				Method:    req.Method,
				Path:      reqPath,
				Query:     reqQuery,
				Headers:   reqHeaders,
				Body:      reqAuditor.toAuditBody(),
			},
		}

		if respCalled {
			entry.Resp = &logzaudit.AuditResponse{
				CreatedAt: respCreatedAt,
				Status:    respStatus,
				Headers:   respHeaders,
				Body:      respAuditor.toAuditBody(),
			}
		}

		l := f.logger
		if l == nil {
			l = logger.(logz.AuditLogger)
		}
		l.AuditReqResp(ctx, entry)
	}()

	fixHeaders := func(outStatus int, outHeaders http.Header) int {
		respCalled = true
		respCreatedAt = time.Now().UTC().UnixMilli()
		respStatus = outStatus
		if outHeaders != nil {
			respHeaders = outHeaders.Clone()
		}
		return outStatus
	}

	fixWriter := func(outWriter io.Writer) io.Writer {
		if respStatus != http.StatusNoContent && respStatus != http.StatusNotModified {
			respAuditor.hasBody = true
		}
		return &auditWriter{
			w:       outWriter,
			auditor: respAuditor,
		}
	}

	next.WrapResponse(ctx, resp, req, fixHeaders, fixWriter)
}

type bodyAuditor struct {
	mu      sync.Mutex
	hasBody bool
	size    int
	payload []byte
	tail    []byte
}

func (b *bodyAuditor) process(p []byte) {
	if len(p) == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.hasBody = true
	prevSize := b.size
	b.size += len(p)

	// 1. Capture head portion up to LargeBodyThreashold
	if prevSize < logzaudit.LargeBodyThreashold {
		headNeeded := logzaudit.LargeBodyThreashold - prevSize
		headTake := min(len(p), headNeeded)
		b.payload = append(b.payload, p[:headTake]...)
		p = p[headTake:]
	}

	// 2. Capture tail portion from remaining non-head content
	if len(p) > 0 {
		if cap(b.tail) < logzaudit.TailBodyThreashold*2 {
			b.tail = make([]byte, 0, logzaudit.TailBodyThreashold*2)
		}
		if len(p) >= logzaudit.TailBodyThreashold {
			tailChunk := p[len(p)-logzaudit.TailBodyThreashold:]
			b.tail = append(b.tail[:0], tailChunk...)
		} else {
			b.tail = append(b.tail, p...)
			if len(b.tail) > logzaudit.TailBodyThreashold {
				overflow := len(b.tail) - logzaudit.TailBodyThreashold
				copy(b.tail, b.tail[overflow:])
				b.tail = b.tail[:logzaudit.TailBodyThreashold]
			}
		}
	}
}

func (b *bodyAuditor) toAuditBody() *logzaudit.AuditBody {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.hasBody {
		return nil
	}

	body := &logzaudit.AuditBody{
		Size:           b.size,
		Large:          b.size > logzaudit.LargeBodyThreashold,
		CompressedSize: 0,
		Payload:        b.payload,
	}

	if body.Large && len(b.tail) > 0 {
		tailCopy := make([]byte, len(b.tail))
		copy(tailCopy, b.tail)
		body.Tail = tailCopy
	}

	return body
}

type auditReader struct {
	r       io.Reader
	auditor *bodyAuditor
	closed  bool
}

func (ar *auditReader) Read(p []byte) (int, error) {
	if ar.closed {
		return 0, io.EOF
	}
	n, err := ar.r.Read(p)
	if n > 0 {
		ar.auditor.process(p[:n])
	}
	return n, err
}

func (ar *auditReader) Close() error {
	ar.closed = true
	if closer, ok := ar.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type auditWriter struct {
	w       io.Writer
	auditor *bodyAuditor
}

func (aw *auditWriter) Write(p []byte) (int, error) {
	n, err := aw.w.Write(p)
	if n > 0 {
		aw.auditor.process(p[:n])
	}
	return n, err
}

func (aw *auditWriter) WriteString(s string) (int, error) {
	return aw.Write([]byte(s))
}

func (aw *auditWriter) Flush() {
	if f, ok := aw.w.(http.Flusher); ok {
		f.Flush()
	}
}
