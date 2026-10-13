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
	ReqResp        logzaudit.AuditEntryReqResp
	ReqCompressed  bool
	RespCompressed bool

	rawReqBody  *ioz.ReaderCounter
	rawRespBody *ioz.WriterCounter

	headTailReqBody  *ioz.ReaderHeadTail
	headTailRespBody *ioz.WriterHeadTail
}

func setBody(auditBody *logzaudit.AuditBody, compressed bool, raw ioz.Countable, body ioz.HeadTail) {
	auditBody.Payload = body.HeadBytes()
	auditBody.Tail = body.TailBytes()
	auditBody.Size = int(body.GetCount())
	auditBody.Large = body.GetCount() >= logzaudit.LargeBodyThreashold
	if compressed {
		auditBody.CompressedSize = int(raw.GetCount())
	}
}

func (e *Entry) BuildReqResp() *logzaudit.AuditEntryReqResp {
	e.ReqResp.Req.Body = &logzaudit.AuditBody{}
	e.ReqResp.Resp.Body = &logzaudit.AuditBody{}
	setBody(e.ReqResp.Req.Body, e.ReqCompressed, e.rawReqBody, e.headTailReqBody)
	setBody(e.ReqResp.Resp.Body, e.RespCompressed, e.rawRespBody, e.headTailRespBody)

	return &e.ReqResp
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
