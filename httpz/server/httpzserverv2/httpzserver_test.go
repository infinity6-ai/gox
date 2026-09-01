package httpzserverv2_test

import (
	"context"
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

	s.AddFilter(func(ctx context.Context, resp *httpzserverv2.Resp, req *httpzserverv2.Req, next httpzserverv2.Handler) {
		resp.Headers.Set("a", "x1")
		resp.Headers.Set("b", "x1")
		resp.Headers.Set("c", "x1")
		next(ctx, resp, req)
	})

	s.AddFilter(func(ctx context.Context, resp *httpzserverv2.Resp, req *httpzserverv2.Req, next httpzserverv2.Handler) {
		resp.Headers.Set("b", "x2")
		next(ctx, resp, req)
	})

	s.AddFilter(func(ctx context.Context, resp *httpzserverv2.Resp, req *httpzserverv2.Req, next httpzserverv2.Handler) {
		resp.Status = http.StatusBadRequest
		resp.Headers.Set("a", "y")
		resp.Write([]byte("nok"))
	})

	s.Start()

	resp, err := http.Get(s.Base())
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "y", resp.Header.Get("a"))
	require.Equal(t, "x2", resp.Header.Get("b"))
	require.Equal(t, "x1", resp.Header.Get("c"))
	data, err := io.ReadAll(resp.Body)
	errorz.Check(err)
	require.Equal(t, "nok", string(data))
}
