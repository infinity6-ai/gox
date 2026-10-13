package httpzserverfilteraudit

import (
	"context"
	"io"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

type kt int

const k kt = 1

type Entry struct {
	ReqResp *logzaudit.AuditEntryReqResp

	rawReqBody       *ioz.ReaderCounter
	rawRespBody      *ioz.WriterCounter
	headTailReqBody  *ioz.ReaderHeadTail
	headTailRespBody *ioz.WriterHeadTail
}

func (e *Entry) CreateReqBody(r io.Reader) io.Reader {
	checker.NotNil(r, "wrong")
	if e.rawReqBody == nil {
		e.rawReqBody = ioz.NewReaderCounter(r)
	}
	e.headTailReqBody = ioz.NewReaderHeadTail(e.rawReqBody, logzaudit.LargeBodyThreashold, logzaudit.TailBodyThreashold)
	return e.headTailReqBody
}

func (e *Entry) CreateRespBody(w io.Writer) io.Writer {
	checker.NotNil(w, "wrong")
	if e.rawRespBody == nil {
		e.rawRespBody = ioz.NewWriterCounter(w)
	}
	e.headTailRespBody = ioz.NewWriterHeadTail(e.rawRespBody, logzaudit.LargeBodyThreashold, logzaudit.TailBodyThreashold)
	return e.headTailRespBody
}

func GetEntry(ctx context.Context) *Entry {
	ret := ctx.Value(k)
	checker.NotNil(ret, "AuditEntryReqResp not found")
	return ret.(*Entry)
}
