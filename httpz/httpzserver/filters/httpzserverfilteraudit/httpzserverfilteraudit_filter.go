package httpzserverfilteraudit

import (
	"context"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

func Collector() func() []*logzspec.Entry {
	return logger.(logz.Collector).Collector()
}

type AuditFilter struct {
	as *logzaudit.AuditService
}

func New(as *logzaudit.AuditService) *AuditFilter {
	return &AuditFilter{as: as}
}

func (a *AuditFilter) WriterFilter() httpzserver.Filter {
	return func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
		checker.Nil(ctx.Value(k), "wrong")
		entry := &Entry{}
		req.Body = entry.CreateReqBody(req.Body)
		defer func() {
			logger.(logz.AuditLogger).AuditReqResp(ctx, entry.BuildReqResp())
		}()
		ctx = context.WithValue(ctx, k, entry.ReqResp)
		next(ctx, func(status int, headers http.Header) io.Writer {
			w := resp(status, headers)
			w = entry.CreateRespBody(w)
			return w
		}, req)
	}
}
