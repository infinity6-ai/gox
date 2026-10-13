package httpzserverfilterauditv2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

type kt int

const k kt = 1

type AuditFilter struct {
	as *logzaudit.AuditService
}

func New(as *logzaudit.AuditService) *AuditFilter {
	return &AuditFilter{as: as}
}

func (a *AuditFilter) WriterFilter() httpzserver.Filter {
	return func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
		checker.Nil(ctx.Value(k), "wrong")
		reqResp := &logzaudit.AuditEntryReqResp{}
		defer func() {
			logger.(logz.AuditLogger).AuditReqResp(ctx, reqResp)
		}()
		ctx = context.WithValue(ctx, k, reqResp)
		head, err := ioz.ReadFull(req.Body, logzaudit.LargeBodyThreashold)
		if err != nil {
			logger.Error(ctx, "audit error", nil, fmt.Errorf("error reading body head: %w", err))
			resp(http.StatusBadRequest, nil).Write([]byte("error reading body head"))
			return
		}
		reqResp.Req.Body.SetPayload(head)
		req.Body = io.MultiReader(bytes.NewBuffer(head), req.Body)
		next(ctx, resp, req)
	}
}

func GetEntry(ctx context.Context) *logzaudit.AuditEntryReqResp {
	ret := ctx.Value(k)
	checker.NotNil(ret, "AuditEntryReqResp not found")
	return ret.(*logzaudit.AuditEntryReqResp)
}
