package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func Schema() *schemaz.Api {
	return &schemaz.Api{
		Id: "samplefraction",

		Desc: schemaz.Desc{
			Name:     "Sample Fraction",
			Summary:  "Calculates the result of a fraction with a given precision.",
			Markdown: "# Sample Fraction API\n\nThis API demonstrates a simple fraction calculation. It accepts a numerator and a denominator as path parameters, a precision as a query parameter, and returns the result in a JSON object.",
		},

		Method: "POST",
		Path:   "/api/gox/routez/sample/fraction/{numerator}/{denominator}",

		ReqParams: []schemaz.Field{
			{Name: "numerator", Desc: schemaz.Desc{Summary: "numerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			{Name: "denominator", Desc: schemaz.Desc{Summary: "denominator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqQuery: []schemaz.Field{
			{Name: "precision", Desc: schemaz.Desc{Summary: "precision"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqHeaders: []schemaz.Field{
			{Name: "x_i6_trace_id", Desc: schemaz.Desc{Summary: "trace id"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
		},

		ReqBody: &schemaz.Spec{
			Type: schemaz.TypeObject,
			Fields: []schemaz.Field{
				{Name: "reason", Desc: schemaz.Desc{Summary: "reason"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
			},
		},

		RespHeaders: []schemaz.Field{
			{Name: "x_i6_trace_message", Desc: schemaz.Desc{Summary: "trace message"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
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
