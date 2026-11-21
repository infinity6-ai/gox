package jsapiz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/routez/jsapiz"
	"github.com/infinity6-ai/gox/schemaz/jstypez"
	"github.com/stretchr/testify/require"
)

func TestUnitExtractTypedefs(t *testing.T) {
	api := &jsapiz.Api{
		Spec: jsapiz.ApiSpec{
			CamelId: "apiCustomAction",
			Method:  "POST",
			Path:    "/api/custom/action",
		},
		Req: jsapiz.Req{
			Payload: jstypez.Typedef{
				Head: &jstypez.Tuple{Typ: "apiCustomActionReqPayload", GoTyp: "models.CustomReq", Name: "payload"},
			},
		},
		Resp: jsapiz.Resp{
			Payload: jstypez.Typedef{
				Head: &jstypez.Tuple{Typ: "apiCustomActionRespPayload", GoTyp: "models.CustomResp", Name: "payload"},
			},
		},
	}

	tds := jsapiz.ExtractTypedefs([]*jsapiz.Api{api})
	allTds := tds.All()
	require.Len(t, allTds, 2)

	var reqTd *jstypez.Typedef
	var respTd *jstypez.Typedef
	for _, td := range allTds {
		if td.Head.Name == "apiCustomActionReq" {
			reqTd = td
		}
		if td.Head.Name == "apiCustomActionResp" {
			respTd = td
		}
	}
	require.NotNil(t, reqTd)
	require.NotNil(t, respTd)
	require.Equal(t, "models.CustomReq", reqTd.Properties.All()[0].GoTyp)
	require.Equal(t, "models.CustomResp", respTd.Properties.All()[0].GoTyp)
}
