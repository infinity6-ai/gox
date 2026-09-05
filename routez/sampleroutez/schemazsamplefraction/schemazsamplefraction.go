package schemazsamplefraction

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

func Api() *apiz.Api[Fraction, Precision, Options, Reason, Meta, Result] {
	return &apiz.Api[Fraction, Precision, Options, Reason, Meta, Result]{
		Schema: Schema(),
		Handler: func(ctx context.Context, req *apiz.Req[Fraction, Precision, Options, Reason]) (*apiz.Resp[Meta, Result], error) {
			return &apiz.Resp[Meta, Result]{
				Status: 201,
				RespHeaders: Meta{
					ReqId: "reason: " + req.ReqBody.Reason + ", trace: " + req.ReqHeaders.TraceId,
				},
				RespBody: Result{
					Display: fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(req.QueryParams.Precision), int(req.QueryParams.Precision)), req.PathParams.Numerator, req.PathParams.Denumerator),
					Result:  strconv.FormatFloat(req.PathParams.Numerator/req.PathParams.Denumerator, 'f', req.QueryParams.Precision, 64),
				},
			}, nil
		},
	}
}
