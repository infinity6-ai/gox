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

	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		preBody := strings.NewReader("reqpre-")
		originalBody := req.Body
		sufBody := strings.NewReader("-reqsuf")
		req.Body = io.MultiReader(preBody, originalBody, sufBody)
		nResp := func(status int, nHeaders http.Header) io.Writer {
			nHeaders.Set("a", "x1")
			nHeaders.Set("b", "x1")
			return resp(status, nHeaders)
		}
		next(ctx, nResp, req)

	})

	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		w := next.WrapResponse(ctx, resp, req, func(outHeaders http.Header) int {
			outHeaders.Set("b", "x2")
			outHeaders.Set("c", "x2")
			return 0
		}, func(outWriter io.Writer) io.Writer {
			_, err := outWriter.Write([]byte("UUUU "))
			errorz.Check(err)
			return outWriter
		})
		_, err := w.Write([]byte(" ZZZZ"))
		errorz.Check(err)
	})

	s.AddHandler("POST", "/bla/{p1}/b/{p2}/c/*", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		reqBody, err := io.ReadAll(req.Body)
		errorz.Check(err)
		body := fmt.Appendf(nil, "r: %s - %s, req: %s", req.Path, params, string(reqBody))
		w := resp(http.StatusBadRequest, http.Header{
			"a": []string{"x1"},
			"d": []string{"y"},
		})
		_, err = w.Write(body)
		errorz.Check(err)

	})

	s.Start()

	resp, err := http.Post(s.Base()+"/bla/O1/b/O2/c/xyz", "text/plain", strings.NewReader("mybody"))
	errorz.Check(err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "x1", resp.Header.Get("a"))
	require.Equal(t, "x1", resp.Header.Get("b"))
	require.Equal(t, "x2", resp.Header.Get("c"))
	require.Equal(t, "y", resp.Header.Get("d"))
	data, err := io.ReadAll(resp.Body)
	errorz.Check(err)
	require.Equal(t, "UUUU r: /bla/O1/b/O2/c/xyz - map[p1:O1 p2:O2], req: reqpre-mybody-reqsuf ZZZZ", string(data))
}
