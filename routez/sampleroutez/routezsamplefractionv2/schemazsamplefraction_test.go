package routezsamplefractionv2_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/apiclientzv2"
	"github.com/infinity6-ai/gox/routez/routezv2"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefractionv2"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	routezv2.Register(s, routezsamplefractionv2.Services()...)

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
	respBody := jsonz.MustParseReader(resp.Body, &routezsamplefractionv2.Result{})
	require.Equal(t, &routezsamplefractionv2.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, respBody)

	reqResp := &routezsamplefractionv2.FractionApi{
		Req: &routezsamplefractionv2.FractionReq{
			Numerator:   10,
			Denominator: 3,
			Precision:   3,
			TraceId:     "xx",
			Reason:      "myreason",
		},
	}
	status, err := apiclientzv2.Do(ctx, c, reqResp)
	errorz.Check(err)
	require.Equal(t, 201, status)
	require.Equal(t, "reason: myreason, trace: xx", reqResp.Resp.TraceMessage)
	require.Equal(t, &routezsamplefractionv2.Result{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, reqResp.Resp.Result)

}
