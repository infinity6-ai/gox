package httpzserverfilterauditv2

import (
	"context"

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
		// head, err := ioz.ReadFull(req.Body, logzaudit.LargeBodyThreashold)
		// if err != nil {
		// 	logger.Error(ctx, "audit error", nil, fmt.Errorf("error reading body head: %w", err))
		// 	resp(http.StatusBadRequest, nil).Write([]byte("error reading body head"))
		// 	return
		// }
		// reqResp.Req.Body.SetPayload(head)
		// req.Body = io.MultiReader(bytes.NewBuffer(head), req.Body)
		// next(ctx, resp, req)
	}
}
