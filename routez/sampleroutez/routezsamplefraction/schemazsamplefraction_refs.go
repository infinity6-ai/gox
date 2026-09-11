package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
)

func (f *FractionApi) GetDataRefs() *apiz.DataRefs {
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
		PathParams:  f.schemaPathParams(),
		QueryParams: f.schemaQueryParams(),
		ReqHeaders:  f.schemaReqHeaders(),
		ReqBody:     f.schemaReqBody(),
		RespHeaders: f.schemaRespHeaders(),
		RespBody:    f.schemaRespBody(),
	}
}

func (f *FractionApi) schemaRespBody() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"display": {Raw: func() any { return &f.Resp.Result.Display }},
				"result":  {Raw: func() any { return &f.Resp.Result.Result }},
			}
		},
	}
}

func (f *FractionApi) schemaRespHeaders() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"x_i6_trace_message": {Str: schemazv2.ParseStr(&f.Resp.TraceMessage)},
			}
		},
	}
}

func (f *FractionApi) schemaReqBody() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"reason": {Raw: func() any { return &f.Req.Reason }},
			}
		},
	}
}

func (f *FractionApi) schemaReqHeaders() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"x_i6_trace_id": {Str: schemazv2.ParseStr(&f.Req.TraceId)},
			}
		},
	}
}

func (f *FractionApi) schemaQueryParams() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"precision": {Str: schemazv2.ParseStrNumber(&f.Req.Precision)},
			}
		},
	}
}

func (f *FractionApi) schemaPathParams() *schemazv2.Schema {
	return &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"numerator":   {Str: schemazv2.ParseStrNumber(&f.Req.Numerator)},
				"denominator": {Str: schemazv2.ParseStrNumber(&f.Req.Denominator)},
			}
		},
	}
}
