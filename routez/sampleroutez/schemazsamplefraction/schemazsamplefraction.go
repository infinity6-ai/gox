package schemazsamplefraction

// import (
// 	"context"
// 	"fmt"
// 	"strconv"

// 	"github.com/infinity6-ai/gox/bla/schemaz"
// 	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
// 	"github.com/infinity6-ai/gox/routez/routez"
// )

// type ReqParams struct {
// 	Numerator   float64 `json:"numerator"`
// 	Denumerator float64 `json:"denumerator"`
// }

// type ReqQuery struct {
// 	Precision int `json:"precision"`
// }

// type ReqHeaders struct {
// 	TraceId string `json:"trace_id"`
// }

// type ReqBody struct {
// 	Reason string `json:"reason"`
// }

// type RespHeaders struct {
// 	ReqId string `json:"req_id"`
// }

// type RespBody struct {
// 	Display string `json:"display"`
// 	Result  string `json:"result"`
// }

// func Api() *schemaz.Api {
// 	return &schemaz.Api{
// 		Id: "samplefraction",

// 		Desc: schemaz.Desc{
// 			Name:     "Sample Fraction",
// 			Summary:  "That is a sample of how to use this",
// 			Markdown: "# Sample Fraction",
// 		},

// 		Method: "POST",
// 		Path:   "/api/gox/dataschema/sample/fraction/{numerator}/{denumerator}",

// 		ReqParams: []schemaz.Field{
// 			{Name: "numerator", Desc: schemaz.Desc{Summary: "numerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
// 			{Name: "denumerator", Desc: schemaz.Desc{Summary: "denumerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
// 		},

// 		ReqQuery: []schemaz.Field{
// 			{Name: "precision", Desc: schemaz.Desc{Summary: "precision"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
// 		},

// 		ReqHeaders: []schemaz.Field{
// 			{Name: "trace_id", Desc: schemaz.Desc{Summary: "trace id"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
// 		},

// 		ReqBody: &schemaz.Spec{
// 			Type: schemaz.TypeObject,
// 			Fields: []schemaz.Field{
// 				{Name: "reason", Desc: schemaz.Desc{Summary: "reason"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
// 			},
// 		},

// 		RespHeaders: []schemaz.Field{
// 			{Name: "req_id", Desc: schemaz.Desc{Summary: "request id"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
// 		},

// 		RespBody: &schemaz.Spec{
// 			Type: schemaz.TypeObject,
// 			Fields: []schemaz.Field{
// 				{Name: "display", Desc: schemaz.Desc{Summary: "fraction display"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
// 				{Name: "result", Desc: schemaz.Desc{Summary: "fraction result"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
// 			},
// 		},
// 	}
// }

// func Handlers(s *httpzserver.Server) {
// 	routez.Add(s, &routez.Api[ReqParams, ReqQuery, ReqHeaders, ReqBody, *RespHeaders, *RespBody]{
// 		Schema: Api(),
// 		Handler: func(ctx context.Context, req *routez.Req[ReqParams, ReqQuery, ReqHeaders, ReqBody]) (*routez.Resp[*RespHeaders, *RespBody], error) {
// 			return &routez.Resp[*RespHeaders, *RespBody]{
// 				Status: 201,
// 				RespHeaders: &RespHeaders{
// 					ReqId: "reason: " + req.ReqBody.Reason + ", trace: " + req.ReqHeaders.TraceId,
// 				},
// 				RespBody: &RespBody{
// 					Display: fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(req.QueryParams.Precision), int(req.QueryParams.Precision)), req.PathParams.Numerator, req.PathParams.Denumerator),
// 					Result:  strconv.FormatFloat(req.PathParams.Numerator/req.PathParams.Denumerator, 'f', req.QueryParams.Precision, 64),
// 				},
// 			}, nil
// 		},
// 	})
// }
