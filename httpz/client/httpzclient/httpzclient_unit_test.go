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

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
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

func TestUnitClientPathResolution(t *testing.T) {
	// 1. Setup server
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("GET", "/api/v1/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("path-prefix-test"))
		require.NoError(t, err)
	})
	s.Listen()
	s.Start()

	t.Run("with path prefix and trailing slash", func(t *testing.T) {
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base().MustJoinPathString("/api/v1/")})
		req := httpzclient.NewReq("GET", "test")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "path-prefix-test", string(respBody))
	})

	t.Run("with path prefix and no trailing slash", func(t *testing.T) {
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base().MustJoinPathString("/api/v1")})
		req := httpzclient.NewReq("GET", "/test")
		_, err := client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "path escaped error")
	})

	t.Run("without path prefix", func(t *testing.T) {
		s.AddHandler("GET", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
			w := resp(http.StatusOK, http.Header{})
			_, err := w.Write([]byte("no-prefix-test"))
			require.NoError(t, err)
		})

		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})
		req := httpzclient.NewReq("GET", "/test")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "no-prefix-test", string(respBody))
	})
}

func TestUnitClientUrlResolution(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("GET", "/absolute-url-test", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("absolute-url-response"))
		require.NoError(t, err)
	})

	s.AddHandler("GET", "/relative-url-test", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("relative-url-response"))
		require.NoError(t, err)
	})

	s.Listen()
	s.Start()

	type testScenario struct {
		name          string
		clientBaseUrl *urlz.Url
		requestUrl    *urlz.Url
		requestPath   string
		expectedBody  string
		expectedError string
		skipServerReq bool
	}

	check := func(t *testing.T, scenario testScenario) {
		t.Helper()
		clientOpts := httpzclient.Options{}
		if scenario.clientBaseUrl != nil {
			clientOpts.BaseUrl = scenario.clientBaseUrl
		}

		client := httpzclient.New(ctx, clientOpts)
		req := &httpzclient.Req{
			Method: "GET",
		}
		if scenario.requestUrl != nil {
			req.Url = scenario.requestUrl
		}
		if scenario.requestPath != "" {
			req.Path = pathz.MustParse(scenario.requestPath)
		}

		if scenario.skipServerReq {
			// Test cases that are expected to panic before making a server request
			require.Panics(t, func() {
				_, _ = client.Do(req)
			})
			return
		}

		resp, err := client.Do(req)

		if scenario.expectedError != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), scenario.expectedError)
			return
		}

		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, scenario.expectedBody, string(respBody))
	}

	t.Run("absolute URL without client BaseUrl", func(t *testing.T) {
		check(t, testScenario{
			name:         "absolute URL without client BaseUrl",
			requestUrl:   urlz.MustParse(s.Base().String() + "/absolute-url-test"),
			expectedBody: "absolute-url-response",
		})
	})

	t.Run("absolute URL with matching client BaseUrl", func(t *testing.T) {
		check(t, testScenario{
			name:          "absolute URL with matching client BaseUrl",
			clientBaseUrl: s.Base(),
			requestUrl:    urlz.MustParse(s.Base().String() + "/absolute-url-test"),
			expectedBody:  "absolute-url-response",
		})
	})

	t.Run("absolute URL with non-matching client BaseUrl", func(t *testing.T) {
		check(t, testScenario{
			name:          "absolute URL with non-matching client BaseUrl",
			clientBaseUrl: urlz.MustParse("http://localhost:9999"), // Mismatching base URL
			requestUrl:    urlz.MustParse(s.Base().String() + "/absolute-url-test"),
			expectedError: "base url mismatch",
		})
	})

	t.Run("relative URL with client BaseUrl", func(t *testing.T) {
		check(t, testScenario{
			name:          "relative URL with client BaseUrl",
			clientBaseUrl: s.Base(),
			requestUrl:    urlz.MustParse("/relative-url-test"),
			expectedBody:  "relative-url-response",
		})
	})

	t.Run("relative URL without client BaseUrl should fail", func(t *testing.T) {
		check(t, testScenario{
			name:          "relative URL without client BaseUrl should fail",
			requestUrl:    urlz.MustParse("/relative-url-test"),
			expectedError: "invalid base URL: must be an absolute URL",
		})
	})

	t.Run("using both Path and Url should panic", func(t *testing.T) {
		check(t, testScenario{
			name:          "using both Path and Url should panic",
			requestUrl:    urlz.MustParse("/some-url"),
			requestPath:   "/some-path",
			skipServerReq: true,
		})
	})

	t.Run("using neither Path nor Url should panic", func(t *testing.T) {
		check(t, testScenario{
			name:          "using neither Path nor Url should panic",
			skipServerReq: true,
		})
	})
}

func TestUnitClientErrorHandling(t *testing.T) {
	ctx := t.Context()

	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
	s.Start()
	s.Close()
	s.Close()

	t.Run("Request execution fails", func(t *testing.T) {
		// Using a non-routable IP address to simulate a network error
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})
		req := httpzclient.NewReq("GET", "/")
		_, err := client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to execute request")
	})
}
