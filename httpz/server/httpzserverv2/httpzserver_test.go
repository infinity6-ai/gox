package httpzserverv2_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserverv2"
	"github.com/stretchr/testify/require"
)

func TestUnitListen(t *testing.T) {
	ctx := t.Context()
	s := httpzserverv2.New(ctx, httpzserverv2.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
	s.Start()

	resp, err := http.Get(s.Base())
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "a", resp.Header.Get("x"))
	data, err := io.ReadAll(resp.Body)
	errorz.Check(err)
	require.Equal(t, "nok", string(data))
}
