package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Message struct {
	Message string `json:"message"`
}

type Result struct {
	Display  string   `json:"display"`
	Result   string   `json:"result"`
	Messages []string `json:"messages"`
}

type FractionReq struct {
	Numerator   float64    `json:"numerator"`
	Denominator float64    `json:"denominator"`
	Precision   int        `json:"precision"`
	TraceId     string     `json:"x-i6-trace-id"`
	Reason      string     `json:"reason"`
	Messages    []*Message `json:"messages"`
}

type FractionResp struct {
	TraceMessage string  `json:"x-i6-trace-message"`
	Result       *Result `json:"result"`
}

type FractionApi struct {
	Req  *FractionReq
	Resp *FractionResp
}

func (f *FractionApi) ApiSpec() apiz.ApiSpec {
	return apiz.ApiSpec{
		Id: "sample-fraction",
		Desc: func() *schemaz.Desc {
			return &schemaz.Desc{
				Name:    "Sample Fraction",
				Summary: "Sample API that performs a fraction operation",
			}
		},
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
				"display": {Raw: func() any { return &f.Resp.Result.Display }, Desc: &schemaz.Desc{Summary: "Human Fraction Representation"}},
				"result":  {Raw: func() any { return &f.Resp.Result.Result }, Desc: &schemaz.Desc{Summary: "Fraction Result"}},
				"messages": {Raw: func() any {
					return &f.Resp.Result.Messages
				}, Desc: &schemaz.Desc{Summary: "Output Messages"}},
			}
		},
	}
}

func (f *FractionApi) schemaRespHeaders() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"x_i6_trace_message": {Str: schemaz.ParseStr(&f.Resp.TraceMessage), Desc: &schemaz.Desc{Summary: "Sample of any trace message"}},
			}
		},
	}
}

func (f *FractionApi) schemaReqBody() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"reason": {Raw: func() any { return &f.Req.Reason }, Desc: &schemaz.Desc{Summary: "Sample reason of this fraction"}},
				"messages": {
					Array: schemaz.NewArray(&f.Req.Messages, func(idx int, element **Message) *schemaz.Schema {
						return &schemaz.Schema{
							Desc: &schemaz.Desc{Summary: "message"},
							Object: func(read bool) map[string]*schemaz.Schema {
								return map[string]*schemaz.Schema{
									"message": {
										Desc: &schemaz.Desc{Summary: "message string"},
										Raw: func() any {
											if *element == nil {
												*element = &Message{}
											}
											return &(*element).Message
										},
									},
								}
							},
						}
					}),
					// Array: func() *schemaz.Array {
					// 	return &schemaz.Array{
					// 		Len: func() int { return len(f.Req.Messages) },
					// 		Get: func(idx int, read bool) *schemaz.Schema {
					// 			if !read {
					// 				f.Req.Messages = slicez.GrowLenTo(f.Req.Messages, idx+1)
					// 				for idx := range f.Req.Messages {
					// 					if f.Req.Messages[idx] == nil {
					// 						f.Req.Messages[idx] = &Message{}
					// 					}
					// 				}
					// 			}
					// 			return &schemaz.Schema{
					// 				Desc: &schemaz.Desc{Summary: "message"},
					// 				Object: func(read bool) map[string]*schemaz.Schema {
					// 					return map[string]*schemaz.Schema{
					// 						"message": {
					// 							Desc: &schemaz.Desc{Summary: "message string"},
					// 							Raw:  func() any { return &f.Req.Messages[idx].Message },
					// 						},
					// 					}
					// 				},
					// 			}
					// 		},
					// 	}
					// },
				},
			}
		},
	}
}

func (f *FractionApi) schemaReqHeaders() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"x_i6_trace_id": {Str: schemaz.ParseStr(&f.Req.TraceId), Desc: &schemaz.Desc{Summary: "Sample of Trace Id"}},
			}
		},
	}
}

func (f *FractionApi) schemaQueryParams() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"precision": {Str: schemaz.ParseStrNumber(&f.Req.Precision), Desc: &schemaz.Desc{Summary: "Fraction max precision"}},
			}
		},
	}
}

func (f *FractionApi) schemaPathParams() *schemaz.Schema {
	return &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"numerator":   {Str: schemaz.ParseStrNumber(&f.Req.Numerator), Desc: &schemaz.Desc{Summary: "Fraction Numerator"}},
				"denominator": {Str: schemaz.ParseStrNumber(&f.Req.Denominator), Desc: &schemaz.Desc{Summary: "Fraction Denominator"}},
			}
		},
	}
}
