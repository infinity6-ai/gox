package httpzserverfilteraudit

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzctx"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

func Collector() func() []*logzspec.Entry {
	return logger.(logz.Collector).Collector()
}

func Filter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	checker.Nil(ctx.Value(k), "audit filter already registered in context")
	entry := &Entry{}

	if req.Headers != nil {
		entry.ReqResp.Id = req.Headers.Get(httpz.HeaderReqId)
	}
	entry.ReqResp.Service = getAuditService(ctx)

	var host string
	if req.Headers != nil {
		host = req.Headers.Get("Host")
	}
	if host == "" && req.Url != nil {
		host = req.Url.Host
	}

	var pathStr string
	if req.Path != nil {
		pathStr = req.Path.String()
	} else if req.Url != nil && req.Url.Path != nil {
		pathStr = req.Url.Path.String()
	}

	entry.ReqResp.Req = &logzaudit.AuditRequest{
		CreatedAt: time.Now().UTC().UnixMilli(),
		Host:      host,
		Method:    req.Method,
		Path:      pathStr,
		Query:     req.Query,
		Headers:   req.Headers,
	}
	entry.ReqResp.Resp = &logzaudit.AuditResponse{}

	if req.Body != nil {
		req.Body = entry.CreateReqBody(req.Body)
	}

	defer func() {
		if entry.ReqResp.Resp.CreatedAt == 0 {
			entry.ReqResp.Resp.CreatedAt = time.Now().UTC().UnixMilli()
		}
		if entry.ReqResp.Id == "" && req.Headers != nil {
			entry.ReqResp.Id = req.Headers.Get(httpz.HeaderReqId)
		}
		logger.(logz.AuditLogger).AuditReqResp(ctx, entry.BuildReqResp())
	}()

	ctx = context.WithValue(ctx, k, entry)
	next(ctx, func(status int, headers http.Header) io.Writer {
		entry.ReqResp.Resp.CreatedAt = time.Now().UTC().UnixMilli()
		entry.ReqResp.Resp.Status = status
		entry.ReqResp.Resp.Headers = headers
		w := resp(status, headers)
		w = entry.CreateRespBody(w)
		return w
	}, req)
}

func getAuditService(ctx context.Context) *logzaudit.AuditService {
	ret := &logzaudit.AuditService{}
	m := logzctx.Get(ctx)
	ret.Name = m[logz.CtxKeyAppName]
	ret.Version = m[logz.CtxKeyAppVersion]
	return ret
}
