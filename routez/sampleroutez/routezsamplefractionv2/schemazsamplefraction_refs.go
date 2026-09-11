package routezsamplefractionv2

import (
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
		PathParams:  f.schemaPathParams(),
		QueryParams: f.schemaQueryParams(),
		ReqHeaders:  f.schemaReqHeaders(),
		// ReqBody: f.schemaReqBody(),
		// RespHeaders: f.schemaRespHeaders(),
		// RespBody: f.schemaRespBody(),
		// QueryParams: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
		// 	return map[string]*schemazv2.Schema{
		// 		"precision": {Str: func(v string) { strconvz.MustParseNumberInto(v, &f.Req.Precision) }},
		// 	}
		// }},
		// ReqHeaders: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
		// 	return map[string]*schemazv2.Schema{
		// 		"x_i6_trace_id": {Str: func(v string) { f.Req.TraceId = v }},
		// 	}
		// }},
		// ReqBody: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
		// 	return map[string]*schemazv2.Schema{
		// 		"reason": {Raw: func() any { return &f.Req.Reason }},
		// 	}
		// }},
		// RespHeaders: &schemazv2.Schema{Object: func() map[string]*schemazv2.Schema {
		// 	return map[string]*schemazv2.Schema{
		// 		"x_i6_trace_message": {Raw: func() any { return &f.Resp.TraceMessage }},
		// 	}
		// }},
		RespBody: &schemazv2.Schema{Raw: func() any { return &f.Resp.Result }},
	}
}

func (f *FractionReqResp) schemaReqHeaders() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"x_i6_trace_id": {Str: schemazv2.ParseStr(&f.Req.TraceId)},
			}
		},
	}
}

func (f *FractionReqResp) schemaQueryParams() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"precision": {Str: schemazv2.ParseStrNumber(&f.Req.Precision)},
			}
		},
	}
}

func (f *FractionReqResp) schemaPathParams() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"numerator":   {Str: schemazv2.ParseStrNumber(&f.Req.Numerator)},
				"denominator": {Str: schemazv2.ParseStrNumber(&f.Req.Denominator)},
			}
		},
	}
}
