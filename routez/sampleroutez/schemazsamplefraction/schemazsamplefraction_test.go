package schemazsamplefraction_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiclientz"
	"github.com/infinity6-ai/gox/routez/routez"
	"github.com/infinity6-ai/gox/routez/sampleroutez/schemazsamplefraction"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	routez.RegisterV2(s, schemazsamplefraction.ApiV2())

	c := httpzclient.New(ctx, httpzclient.Options{
		BaseUrl: s.Base(),
	})

	req := httpzrequest.New("POST", "/api/gox/dataschema/sample/fraction/10/3").
		SetQuery("precision", "3").
		SetHeader("Trace-Id", "xx").
		SetBody(strings.NewReader("{\"reason\":\"myreason\"}"))
	resp, err := c.Do(ctx, req)
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)
	require.Equal(t, "application/json", resp.Headers.Get("content-type"))
	require.Equal(t, "reason: myreason, trace: xx", resp.Headers.Get("Req-Id"))
	respBody := jsonz.MustParseReader(resp.Body, &schemazsamplefraction.Result{})
	require.Equal(t, &schemazsamplefraction.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, respBody)

	acV2 := apiclientz.GetV2(c, schemazsamplefraction.ApiV2())

	reqResp := &schemazsamplefraction.FractionReqResp{
		Fraction: &schemazsamplefraction.Fraction{
			Numerator:   10,
			Denumerator: 3,
		},
		Precision: &schemazsamplefraction.Precision{
			Precision: 3,
		},
		Options: &schemazsamplefraction.Options{
			TraceId: "xx",
		},
		Reason: &schemazsamplefraction.Reason{
			Reason: "myreason",
		},
	}
	acV2Resp, err := acV2(ctx, reqResp)
	errorz.Check(err)
	require.Equal(t, 201, acV2Resp)
	require.Equal(t, "reason: myreason, trace: xx", reqResp.Meta.ReqId)
	require.Equal(t, &schemazsamplefraction.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, reqResp.Result)

}
