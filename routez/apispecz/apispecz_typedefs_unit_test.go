package apispecz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/routez/apispecz"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/stretchr/testify/require"
)

func TestUnitExtractTypedefs(t *testing.T) {
	api := &apispecz.Api{
		Spec: apispecz.ApiSpec{
			CamelId: "apiCustomAction",
			Method:  "POST",
			Path:    "/api/custom/action",
		},
		Req: apispecz.Req{
			Payload: typeastz.Typedef{
				Head: &typeastz.Tuple{Typ: "apiCustomActionReqPayload", GoTyp: "models.CustomReq", Name: "payload"},
			},
		},
		Resp: apispecz.Resp{
			Payload: typeastz.Typedef{
				Head: &typeastz.Tuple{Typ: "apiCustomActionRespPayload", GoTyp: "models.CustomResp", Name: "payload"},
			},
		},
	}

	tds := apispecz.ExtractTypedefs([]*apispecz.Api{api})
	allTds := tds.All()
	require.Len(t, allTds, 2)

	var reqTd *typeastz.Typedef
	var respTd *typeastz.Typedef
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
