package httpzserverfilteraudit

import (
	"context"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

type AuditFilter struct {
	as *logzaudit.AuditService
}

func New(as *logzaudit.AuditService) *AuditFilter {
	return &AuditFilter{as: as}
}

func (a *AuditFilter) WriterFilter() httpzserver.Filter {
	return func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
		checker.Nil(ctx.Value(k), "wrong")
		entry := &Entry{
			ReqResp: &logzaudit.AuditEntryReqResp{},
		}
		defer func() {
			logger.(logz.AuditLogger).AuditReqResp(ctx, entry.ReqResp)
		}()
		ctx = context.WithValue(ctx, k, entry.ReqResp)
		req.Body = AuditRequestBody(ctx, req.Body)
		next(ctx, func(status int, headers http.Header) io.Writer {
			w := resp(status, headers)
			w = AuditResponseBody(ctx, w)
			return w
		}, req)
	}
}
