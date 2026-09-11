package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Result struct {
	Display string `json:"display"`
	Result  string `json:"result"`
}

type FractionReq struct {
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	Precision   int     `json:"precision"`
	TraceId     string  `json:"x-i6-trace-id"`
	Reason      string  `json:"reason"`
}

type FractionResp struct {
	TraceMessage string  `json:"x-i6-trace-message"`
	Result       *Result `json:"result"`
}

type FractionApi struct {
	Req  *FractionReq
	Resp *FractionResp
}

func Api() *FractionApi {
	return &FractionApi{}
}

func (f *FractionApi) ApiSpec() apiz.ApiSpec {
	return apiz.ApiSpec{
		Id:     "samplefraction",
		Desc:   nil,
		Method: "POST",
		Path:   "/api/gox/routez/sample/fraction/{numerator}/{denominator}",
	}
}

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

func (f *FractionApi) schemaRespBody() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"display": {Raw: func() any { return &f.Resp.Result.Display }},
				"result":  {Raw: func() any { return &f.Resp.Result.Result }},
			}
		},
	}
}

func (f *FractionApi) schemaRespHeaders() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"x_i6_trace_message": {Str: schemaz.ParseStr(&f.Resp.TraceMessage)},
			}
		},
	}
}

func (f *FractionApi) schemaReqBody() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"reason": {Raw: func() any { return &f.Req.Reason }},
			}
		},
	}
}

func (f *FractionApi) schemaReqHeaders() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"x_i6_trace_id": {Str: schemaz.ParseStr(&f.Req.TraceId)},
			}
		},
	}
}

func (f *FractionApi) schemaQueryParams() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"precision": {Str: schemaz.ParseStrNumber(&f.Req.Precision)},
			}
		},
	}
}

func (f *FractionApi) schemaPathParams() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"numerator":   {Str: schemaz.ParseStrNumber(&f.Req.Numerator)},
				"denominator": {Str: schemaz.ParseStrNumber(&f.Req.Denominator)},
			}
		},
	}
}
