package apispecz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/routez/apispecz"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
	"github.com/stretchr/testify/require"
)

func TestUnitFractionSample(t *testing.T) {
	ctx := t.Context()
	api := routezsamplefraction.NewService().Api()
	parser := apispecz.New(nil)
	actual := parser.ParseApi(ctx, api)
	jsonz.MustParse(jsonz.MustFormat(actual).String(), &actual)

	expected := &apispecz.Api{
		Spec: apispecz.ApiSpec{
			Id:      "sample-fraction",
			CamelId: "apiSampleFraction",
			Desc: &schemaz.Desc{
				Name:     "Sample Fraction",
				Summary:  "Sample API that performs a fraction operation",
				Markdown: "",
			},
			Method: "POST",
			Path:   "/api/gox/routez/sample/fraction/{numerator}/{denominator}",
		},
		Typedefs: typeastz.NewTypedefs([]*typeastz.Typedef{
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionParams", Summary: "apiSampleFraction params"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "number", GoTyp: "float64", Name: "denominator", Summary: "Fraction Denominator"},
					{Typ: "number", GoTyp: "float64", Name: "numerator", Summary: "Fraction Numerator"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionQuery", Summary: "apiSampleFraction query"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "number", GoTyp: "int", Name: "precision", Summary: "Fraction max precision"},
					{Typ: "apiSampleFractionQuerySmsg", Name: "smsg"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionQuerySmsg"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", GoTyp: "string", Name: "message", Summary: "message string"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionReqHeaders", Summary: "apiSampleFraction reqHeaders"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", GoTyp: "string", Name: "x_i6_trace_id", Summary: "Sample of Trace Id"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionReqPayload", Summary: "apiSampleFraction reqPayload"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "Array<apiSampleFractionReqPayloadMessagesElement>", Name: "messages"},
					{Typ: "string", GoTyp: "string", Name: "reason", Summary: "Sample reason of this fraction"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionReqPayloadMessagesElement", Summary: "message"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", GoTyp: "string", Name: "message", Summary: "message string"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionRespHeaders", Summary: "apiSampleFraction respHeaders"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "apiSampleFractionRespHeadersXI6Smsg", Name: "x_i6_smsg"},
					{Typ: "string", GoTyp: "string", Name: "x_i6_trace_message", Summary: "Sample of any trace message"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionRespHeadersXI6Smsg"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", GoTyp: "string", Name: "message", Summary: "message string"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "apiSampleFractionRespPayload", Summary: "apiSampleFraction respPayload"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", GoTyp: "string", Name: "display", Summary: "Human Fraction Representation"},
					{Typ: "Array<apiSampleFractionRespPayloadMessagesElement>", Name: "messages", Summary: "Output Messages"},
					{Typ: "string", GoTyp: "string", Name: "result", Summary: "Fraction Result"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "string", GoTyp: "string", Name: "apiSampleFractionRespPayloadMessagesElement"},
			},
			{
				Head: &typeastz.Tuple{Typ: "object", Name: "dszFileStat"},
				Properties: typeastz.NewTuples([]typeastz.Tuple{
					{Typ: "string", Name: "url", Summary: "missing summary"},
					{Typ: "string", Name: "content_type", Summary: "missing summary"},
					{Typ: "string", Name: "md5", Summary: "missing summary"},
					{Typ: "string", Name: "size", Summary: "missing summary"},
					{Typ: "string", Name: "etag", Summary: "missing summary"},
					{Typ: "dszTime", Name: "created_at", Summary: "missing summary"},
					{Typ: "dszTime", Name: "updated_at", Summary: "missing summary"},
				}),
			},
			{
				Head: &typeastz.Tuple{Typ: "string", Name: "dszTime", Summary: "Missing Summary"},
			},
		}),
		Req: apispecz.Req{
			Params: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionParams", Name: "params"},
				Properties: typeastz.NewTuples(nil),
			},
			Query: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionQuery", Name: "query"},
				Properties: typeastz.NewTuples(nil),
			},
			Headers: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionReqHeaders", Name: "reqHeaders"},
				Properties: typeastz.NewTuples(nil),
			},
			Payload: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionReqPayload", Name: "reqPayload"},
				Properties: typeastz.NewTuples(nil),
			},
		},
		Resp: apispecz.Resp{
			Headers: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionRespHeaders", Name: "respHeaders"},
				Properties: typeastz.NewTuples(nil),
			},
			Payload: typeastz.Typedef{
				Head:       &typeastz.Tuple{Typ: "apiSampleFractionRespPayload", Name: "respPayload"},
				Properties: typeastz.NewTuples(nil),
			},
		},
	}
	require.Equal(t, expected, actual)
}
