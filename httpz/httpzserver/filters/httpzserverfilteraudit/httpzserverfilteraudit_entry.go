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
	if auditBody == nil || body == nil {
		return
	}
	auditBody.Payload = body.HeadBytes()
	auditBody.Tail = body.TailBytes()
	auditBody.Size = int(body.GetCount())
	auditBody.Large = body.GetCount() >= logzaudit.LargeBodyThreashold
	if compressed && raw != nil {
		auditBody.CompressedSize = int(raw.GetCount())
	}
}

func (e *Entry) BuildReqResp() *logzaudit.AuditEntryReqResp {
	if e.ReqResp.Req == nil {
		e.ReqResp.Req = &logzaudit.AuditRequest{}
	}
	if e.ReqResp.Resp == nil {
		e.ReqResp.Resp = &logzaudit.AuditResponse{}
	}
	if e.ReqResp.Req.Body == nil {
		e.ReqResp.Req.Body = &logzaudit.AuditBody{}
	}
	if e.ReqResp.Resp.Body == nil {
		e.ReqResp.Resp.Body = &logzaudit.AuditBody{}
	}
	if e.headTailReqBody != nil {
		setBody(e.ReqResp.Req.Body, e.ReqCompressed, e.rawReqBody, e.headTailReqBody)
	}
	if e.headTailRespBody != nil {
		setBody(e.ReqResp.Resp.Body, e.RespCompressed, e.rawRespBody, e.headTailRespBody)
	}

	return &e.ReqResp
}

func (e *Entry) CreateReqBody(r io.Reader) io.Reader {
	if r == nil {
		return nil
	}
	if e.rawReqBody == nil {
		e.rawReqBody = ioz.NewReaderCounter(r)
	}
	if e.headTailReqBody == nil {
		e.headTailReqBody = ioz.NewReaderHeadTail(e.rawReqBody, logzaudit.LargeBodyThreashold, logzaudit.TailBodyThreashold)
	}
	return e.headTailReqBody
}

func (e *Entry) CreateRespBody(w io.Writer) io.Writer {
	if w == nil {
		return nil
	}
	if e.rawRespBody == nil {
		e.rawRespBody = ioz.NewWriterCounter(w)
	}
	if e.headTailRespBody == nil {
		e.headTailRespBody = ioz.NewWriterHeadTail(e.rawRespBody, logzaudit.LargeBodyThreashold, logzaudit.TailBodyThreashold)
	}
	return e.headTailRespBody
}

func GetEntry(ctx context.Context) *Entry {
	checker.NotNil(ctx, "context cannot be nil")
	ret := ctx.Value(k)
	checker.NotNil(ret, "AuditEntryReqResp not found")
	return ret.(*Entry)
}
