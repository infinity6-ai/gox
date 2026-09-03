package schemazsamplefraction_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/infinity6-ai/gox/schemaz/schemazsamplefraction"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	schemazsamplefraction.Handlers(s)

	c := httpzclient.New(ctx, httpzclient.Options{
		BaseUrl: s.Base(),
	})
	req := httpzclient.NewReq("POST", "/api/gox/dataschema/sample/fraction/10/3").
		SetQuery("precision", "3").
		SetHeader("Trace-Id", "xx").
		SetBody(strings.NewReader("{\"reason\":\"myreason\"}"))
	resp, err := c.Do(req)
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, "application/json", resp.Headers.Get("content-type"))
	require.Equal(t, "reason: myreason, trace: xx", resp.Headers.Get("Req-Id"))
	respBody := jsonz.MustParseReader(resp.Body, &schemazsamplefraction.RespBody{})
	require.Equal(t, &schemazsamplefraction.RespBody{
		Display: "10.000/3.000",
		Result:  "3.333",
	}, respBody)
}
