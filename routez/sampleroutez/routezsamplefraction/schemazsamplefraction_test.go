package routezsamplefraction_test

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
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/stretchr/testify/require"
)

func TestManualBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	routez.Register(s, routezsamplefraction.Api())

	c := httpzclient.New(ctx, httpzclient.Options{
		BaseUrl: s.Base(),
	})

	req := httpzrequest.New("POST", "/api/gox/routez/sample/fraction/10/3").
		SetQuery("precision", "3").
		SetHeader("x-i6-trace-id", "xx").
		SetBody(strings.NewReader("{\"reason\":\"myreason\"}"))
	resp, err := c.Do(ctx, req)
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, 201, resp.StatusCode)
	require.Equal(t, "application/json", resp.Headers.Get("content-type"))
	require.Equal(t, "reason: myreason, trace: xx", resp.Headers.Get("x-i6-trace-message"))
	respBody := jsonz.MustParseReader(resp.Body, &routezsamplefraction.Result{})
	require.Equal(t, &routezsamplefraction.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, respBody)

	ac := apiclientz.GetV2(c, routezsamplefraction.Api())

	reqResp := &routezsamplefraction.FractionReqResp{
		Req: &routezsamplefraction.FractionReq{
			Numerator:   10,
			Denominator: 3,
			Precision:   3,
			TraceId:     "xx",
			Reason:      "myreason",
		},
	}
	status, err := ac(ctx, reqResp)
	errorz.Check(err)
	require.Equal(t, 201, status)
	require.Equal(t, "reason: myreason, trace: xx", reqResp.Resp.TraceMessage)
	require.Equal(t, &routezsamplefraction.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, reqResp.Resp.Result)

}
