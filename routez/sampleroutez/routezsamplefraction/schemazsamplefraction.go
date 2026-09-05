package routezsamplefraction

import (
	"context"
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type Fraction struct {
	Numerator   float64 `json:"numerator"`
	Denumerator float64 `json:"denumerator"`
}

type Precision struct {
	Precision int `json:"precision"`
}

type Options struct {
	TraceId string `json:"trace_id"`
}

type Reason struct {
	Reason string `json:"reason"`
}

type Meta struct {
	ReqId string `json:"req_id"`
}

type Result struct {
	Display string `json:"display"`
	Result  string `json:"result"`
}

func Schema() *schemaz.Api {
	return &schemaz.Api{
		Id: "samplefraction",

		Desc: schemaz.Desc{
			Name:     "Sample Fraction",
			Summary:  "That is a sample of how to use this",
			Markdown: "# Sample Fraction",
		},

		Method: "POST",
		Path:   "/api/gox/dataschema/sample/fraction/{numerator}/{denumerator}",

		ReqParams: []schemaz.Field{
			{Name: "numerator", Desc: schemaz.Desc{Summary: "numerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			{Name: "denumerator", Desc: schemaz.Desc{Summary: "denumerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqQuery: []schemaz.Field{
			{Name: "precision", Desc: schemaz.Desc{Summary: "precision"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqHeaders: []schemaz.Field{
			{Name: "trace_id", Desc: schemaz.Desc{Summary: "trace id"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqBody: &schemaz.Spec{
			Type: schemaz.TypeObject,
			Fields: []schemaz.Field{
				{Name: "reason", Desc: schemaz.Desc{Summary: "reason"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
			},
		},

		RespHeaders: []schemaz.Field{
			{Name: "req_id", Desc: schemaz.Desc{Summary: "request id"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
		},

		RespBody: &schemaz.Spec{
			Type: schemaz.TypeObject,
			Fields: []schemaz.Field{
				{Name: "display", Desc: schemaz.Desc{Summary: "fraction display"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
				{Name: "result", Desc: schemaz.Desc{Summary: "fraction result"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			},
		},
	}
}

type FractionReqResp struct {
	Fraction  *Fraction
	Precision *Precision
	Options   *Options
	Reason    *Reason
	Meta      *Meta
	Result    *Result
}

func (f *FractionReqResp) GetDataRefs() *apiz.DataRefs {
	if f.Fraction == nil {
		f.Fraction = &Fraction{}
	}
	if f.Precision == nil {
		f.Precision = &Precision{}
	}
	if f.Options == nil {
		f.Options = &Options{}
	}
	if f.Reason == nil {
		f.Reason = &Reason{}
	}
	if f.Meta == nil {
		f.Meta = &Meta{}
	}
	if f.Result == nil {
		f.Result = &Result{}
	}
	return &apiz.DataRefs{
		PathParams:  f.Fraction,
		QueryParams: f.Precision,
		ReqHeaders:  f.Options,
		ReqBody:     f.Reason,
		RespHeaders: f.Meta,
		RespBody:    f.Result,
	}
}

func Api() *apiz.Api[*FractionReqResp] {
	return &apiz.Api[*FractionReqResp]{
		Schema: Schema(),
		Handler: func(ctx context.Context, reqResp *FractionReqResp) (int, error) {
			reqResp.Meta.ReqId = "reason: " + reqResp.Reason.Reason + ", trace: " + reqResp.Options.TraceId
			reqResp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(reqResp.Precision.Precision), int(reqResp.Precision.Precision)), reqResp.Fraction.Numerator, reqResp.Fraction.Denumerator)
			reqResp.Result.Result = strconv.FormatFloat(reqResp.Fraction.Numerator/reqResp.Fraction.Denumerator, 'f', reqResp.Precision.Precision, 64)
			return 201, nil
		},
	}
}
