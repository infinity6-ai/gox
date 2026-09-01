package httpzclient_test

import (
	"io"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/stretchr/testify/require"
)

func TestUnitClientRequestBuilding(t *testing.T) {
	req := httpzclient.NewReq("GET", "/path")
	require.NotNil(t, req)
	require.Equal(t, "GET", req.Method)
	require.Equal(t, "/path", req.Path.String())

	req.AddQuery("q", "search")
	req.AddHeader("Accept", "application/json")
	req.SetBody(strings.NewReader("body"))

	require.Equal(t, "search", req.Query.Get("q"))
	require.Equal(t, "application/json", req.Headers.Get("Accept"))

	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, "body", string(bodyBytes))
}
