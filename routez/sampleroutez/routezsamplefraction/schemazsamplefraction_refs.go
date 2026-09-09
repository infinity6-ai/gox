package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/routez/apiz"
)

func (f *FractionReqResp) GetDataRefs() *apiz.DataRefs {
	if f.Req == nil {
		f.Req = &FractionReq{}
	}
	if f.Resp == nil {
		f.Resp = &FractionResp{}
	}
	if f.Resp.Result == nil {
		f.Resp.Result = &Result{}
	}
	return &apiz.DataRefs{
		PathParams: &struct {
			Numerator   *float64 `json:"numerator"`
			Denumerator *float64 `json:"denumerator"`
		}{
			&f.Req.Numerator,
			&f.Req.Denumerator,
		},
		QueryParams: &struct {
			Precision *int `json:"precision"`
		}{
			&f.Req.Precision,
		},
		ReqHeaders: &struct {
			TraceId *string `json:"x_i6_trace_id"`
		}{
			&f.Req.TraceId,
		},
		ReqBody: &struct {
			Reason *string `json:"reason"`
		}{
			&f.Req.Reason,
		},
		RespHeaders: &struct {
			TraceMessage *string `json:"x_i6_trace_message"`
		}{
			&f.Resp.TraceMessage,
		},
		RespBody: f.Resp.Result,
	}
}
