package schemazsamplefraction

import (
	"context"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type ReqBody struct {
	Reason string `json:"reason"`
}

type RespBody struct {
	Display string  `json:"display"`
	Result  float64 `json:"result"`
}

func Api() schemaz.Api[ReqBody, RespBody] {
	return schemaz.Api[ReqBody, RespBody]{
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
			{Name: "precsion", Desc: schemaz.Desc{Summary: "precision"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
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
		ReqType: ReqBody{},

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
		RespType: RespBody{},
	}
}

func AddHandler[I, O any](s *httpzserver.Server, api schemaz.Api[I, O]) {
	s.AddHandler(api.Method, api.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		// reqBody, err := jsonz.NewReader[I](req.Body).ReadItem()

	})
}
