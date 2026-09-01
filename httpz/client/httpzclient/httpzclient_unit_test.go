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
	defer client.Close()

	// Add a filter to the client
	client.AddFilter(func(req *httpzclient.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
		// Modify request
		req.AddHeader("X-Client-Filter", "client-filter-value")
		originalBody := req.Body
		req.Body = io.MultiReader(strings.NewReader("filter-prefix-"), originalBody)

		// Call next handler
		resp, err := next(req)
		require.NoError(t, err)

		// Modify response
		resp.Headers.Add("X-Client-Post-Filter", "post-filter-value")
		return resp, nil
	})

	// 3. Make request
	req := httpzclient.NewReq("POST", "/test").
		SetBody(strings.NewReader("original-body"))

	resp, err := client.Do(req)
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

func TestUnitClientErrorHandling(t *testing.T) {
	ctx := t.Context()

	t.Run("Invalid base URL", func(t *testing.T) {
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: "invalid-url"})
		req := httpzclient.NewReq("GET", "/")
		_, err := client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid base URL")
	})

	t.Run("Request execution fails", func(t *testing.T) {
		// Using a non-routable IP address to simulate a network error
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: "http://192.0.2.1:80"})
		req := httpzclient.NewReq("GET", "/")
		_, err := client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to execute request")
	})
}
