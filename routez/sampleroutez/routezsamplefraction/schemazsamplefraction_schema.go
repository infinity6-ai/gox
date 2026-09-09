package routezsamplefraction

import (
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

func Schema() *schemaz.Api {
	return &schemaz.Api{
		Id: "samplefraction",

		Desc: schemaz.Desc{
			Name:     "Sample Fraction",
			Summary:  "That is a sample of how to use this",
			Markdown: "# Sample Fraction",
		},

		Method: "POST",
		Path:   "/api/gox/routez/sample/fraction/{numerator}/{denumerator}",

		ReqParams: []schemaz.Field{
			{Name: "numerator", Desc: schemaz.Desc{Summary: "numerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			{Name: "denumerator", Desc: schemaz.Desc{Summary: "denumerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
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
				{Name: "result", Desc: schemaz.Desc{Summary: "fraction result"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
			},
		},
	}
}
