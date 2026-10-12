package httpzserverfilteraudit

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

func AuditFilter(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
	l := logger.(logz.AuditLogger)
	l.AuditReqResp(ctx, nil)
	next(ctx, resp, req)
}
