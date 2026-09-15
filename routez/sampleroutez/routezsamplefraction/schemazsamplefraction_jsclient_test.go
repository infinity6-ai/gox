package routezsamplefraction_test

import (
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/jsclientz"
	"github.com/infinity6-ai/gox/routez/routez"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/stretchr/testify/require"
)

// TestUnitJsClient gera o SDK JS a partir do schema atual e roda ele de verdade contra um
// servidor real, via node — se o schema mudar sem o gerador ser ajustado, este teste quebra.
func TestUnitJsClient(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	routez.Register(s, routezsamplefraction.Services()...)

	svc := routezsamplefraction.Services()[0].New()
	fnName, code, err := jsclientz.BuildJsFetch(svc.Api().ApiSpec(), svc.Api().GetDataRefs())
	require.NoError(t, err)

	require.Equal(t, "samplefraction", fnName)
	require.Contains(t, code, "@param {number} params.numerator")
	require.Contains(t, code, "@param {string} params.reason")
	require.Contains(t, code, "@returns {Promise<{status: number, headers: {x_i6_trace_message: string}, body: {display: string, result: string}}>}")
	// nomes reais de wire, não os do mockup antigo (que usava "trace-id"/"req-id" hardcoded)
	require.Contains(t, code, `headers["x-i6-trace-id"]`)
	require.Contains(t, code, `resp.headers.get("x-i6-trace-message")`)

	sampleParams := map[string]any{
		"numerator":     10,
		"denominator":   3,
		"precision":     3,
		"x_i6_trace_id": "xx",
		"reason":        "myreason",
	}

	result := jsclientz.RunJsFetch(t, fnName, code, s.Base().String(), sampleParams)

	require.InDelta(t, 201, result["status"], 0)
	require.Equal(t, "reason: myreason, trace: xx", result["headers"].(map[string]any)["x_i6_trace_message"])
	body := result["body"].(map[string]any)
	require.Equal(t, "10.000/3.000", body["display"])
	require.Equal(t, "3.333", body["result"])
}
