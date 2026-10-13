package httpzserverfilteraudit

import (
	"context"

	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

type kt int

const k kt = 1

type Entry struct {
	ReqResp *logzaudit.AuditEntryReqResp
}

func GetEntry(ctx context.Context) *Entry {
	ret := ctx.Value(k)
	checker.NotNil(ret, "AuditEntryReqResp not found")
	return ret.(*Entry)
}
