package routezsamplefractionv2

import (
	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
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
			Denominator *float64 `json:"denominator"`
		}{
			&f.Req.Numerator,
			&f.Req.Denominator,
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

func (f *FractionReqResp) GetDataRefsV2() *apiz.DataRefsV2 {
	if f.Req == nil {
		f.Req = &FractionReq{}
	}
	if f.Resp == nil {
		f.Resp = &FractionResp{}
	}
	if f.Resp.Result == nil {
		f.Resp.Result = &Result{}
	}
	return &apiz.DataRefsV2{
		PathParams: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"numerator":   {Str: func(unformatted string) { strconvz.MustParseNumberInto(unformatted, &f.Req.Numerator) }},
				"denominator": {Str: func(unformatted string) { strconvz.MustParseNumberInto(unformatted, &f.Req.Denominator) }},
			}
		}},
		QueryParams: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"precision": {Str: func(unformatted string) { strconvz.MustParseNumberInto(unformatted, &f.Req.Precision) }},
			}
		}},
		ReqHeaders: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"x_i6_trace_id": {Raw: func() any { return &f.Req.TraceId }},
			}
		}},
		ReqBody: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"reason": {Raw: func() any { return &f.Req.Reason }},
			}
		}},
		RespHeaders: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"x_i6_trace_message": {Raw: func() any { return &f.Resp.TraceMessage }},
			}
		}},
		RespBody: &schemazv2.Schema{Raw: func() any { return &f.Resp.Result }},
	}
}
