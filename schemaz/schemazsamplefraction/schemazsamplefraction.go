package schemazsamplefraction

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemahttpz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type ReqParams struct {
	Numerator   float64 `json:"numerator"`
	Denumerator float64 `json:"denumerator"`
}

type ReqQuery struct {
	Precision float64 `json:"precision"`
}

type ReqHeaders struct {
	TraceId string `json:"trace_id"`
}

type ReqBody struct {
	Reason string `json:"reason"`
}

type RespHeaders struct {
	ReqId string `json:"req_id"`
}

type RespBody struct {
	Display string  `json:"display"`
	Result  float64 `json:"result"`
}

func Api() *schemaz.Api {
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

func Handlers(s *httpzserver.Server) {
	schemahttpz.Add(s, &schemahttpz.Api[*ReqParams, *ReqHeaders, *ReqHeaders, *ReqBody, *RespHeaders, *RespBody]{
		Schema: Api(),
		Handler: func(ctx context.Context, req *schemahttpz.Req[*ReqParams, *ReqHeaders, *ReqHeaders, *ReqBody]) (*RespHeaders, *RespBody, error) {
			return &RespHeaders{
					ReqId: fmt.Sprintf("reason: %s, trace: %s", req.ReqBody.Reason, req.ReqHeaders.TraceId),
				},
				&RespBody{
					Display: fmt.Sprintf("%f/%f", req.PathParams.Numerator, req.PathParams.Denumerator),
					Result:  req.PathParams.Numerator / req.PathParams.Denumerator,
				},
				nil
		},
	})
}

// type Req[P any, Q any, IH any, IB any] struct {
// 	PathParams  P
// 	QueryParams Q
// 	ReqHeaders  IH
// 	ReqBody     IB
// }

// type Resp[OH any, OB any] struct {
// 	RespHeaders OH
// 	RespBody    OB
// }

// func AddHandler[P any, Q any, IH any, IB any, OH any, OB any](s *httpzserver.Server, api schemaz.Api[IB, OB], handler func(ctx context.Context, req *Req[I, O], resp *Resp[I, O])) {

// 	s.AddHandler(api.Method, api.Path, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
// 		if api.ReqBody != nil {
// 			// reqBody := jsonz.NewReader[I](req.Body).MustReadItem()
// 		}

// 	})
// }
