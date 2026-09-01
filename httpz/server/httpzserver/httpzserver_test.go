package httpzserver_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitListen(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()

	s.AddFilter(func(ctx context.Context, resp *httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		resp.Headers.Set("a", "x1")
		resp.Headers.Set("b", "x1")
		resp.Headers.Set("c", "x1")
		next(ctx, resp, req)
	})

	s.AddFilter(func(ctx context.Context, resp *httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		resp.Headers.Set("b", "x2")
		next(ctx, resp, req)
	})

	s.AddHandler("POST", "/bla/{p1}/b/{p2}/c/*", func(ctx context.Context, resp *httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		resp.Status = http.StatusBadRequest
		resp.Headers.Set("a", "y")
		reqBody, err := io.ReadAll(req.Body)
		errorz.Check(err)
		resp.Write(fmt.Appendf(nil, "r: %s - %s, req: %s", req.Path, params, string(reqBody)))
	})

	s.Start()

	resp, err := http.Post(s.Base()+"/bla/O1/b/O2/c/xyz", "text/plain", strings.NewReader("mybody"))
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "y", resp.Header.Get("a"))
	require.Equal(t, "x2", resp.Header.Get("b"))
	require.Equal(t, "x1", resp.Header.Get("c"))
	data, err := io.ReadAll(resp.Body)
	errorz.Check(err)
	require.Equal(t, "r: /bla/O1/b/O2/c/xyz - map[p1:O1 p2:O2], req: mybody", string(data))
}
