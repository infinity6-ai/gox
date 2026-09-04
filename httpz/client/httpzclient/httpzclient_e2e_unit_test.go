package httpzclient_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientWithServer(t *testing.T) {
	// 1. Setup server
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("POST", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		// Check request headers from client filter
		require.Equal(t, "client-filter-value", req.Headers.Get("X-Client-Filter"))

		// Check request body from client filter
		require.Equal(t, "filter-prefix-original-body", string(body))

		w := resp(http.StatusOK, http.Header{
			"X-Server-Header": []string{"server-value"},
		})
		_, err = w.Write([]byte("server-response"))
		require.NoError(t, err)
	})
	s.Listen()
	s.Start()

	// 2. Setup client
	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})

	// Add a filter to the client
	client.AddFilter(func(ctx context.Context, req *httpzclient.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
		// Modify request
		req.AddHeader("X-Client-Filter", "client-filter-value")
		originalBody := req.Body
		req.Body = io.MultiReader(strings.NewReader("filter-prefix-"), originalBody)

		// Call next handler
		resp, err := next(ctx, req)
		require.NoError(t, err)

		// Modify response
		resp.Headers.Add("X-Client-Post-Filter", "post-filter-value")
		return resp, nil
	})

	// 3. Make request
	req := httpzclient.NewReq("POST", "/test").
		SetBody(strings.NewReader("original-body"))

	resp, err := client.Do(ctx, req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 4. Assertions
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Check headers
	require.Equal(t, "server-value", resp.Headers.Get("X-Server-Header"))
	require.Equal(t, "post-filter-value", resp.Headers.Get("X-Client-Post-Filter"))

	// Check body
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "server-response", string(respBody))
}
