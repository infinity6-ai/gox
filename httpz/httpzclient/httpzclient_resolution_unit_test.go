package httpzclient_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

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
		req := httpzrequest.NewReq("GET", "test")
		resp, err := client.Do(ctx, req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "path-prefix-test", string(respBody))
	})

	t.Run("with path prefix and no trailing slash", func(t *testing.T) {
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base().MustJoinPathString("/api/v1")})
		req := httpzrequest.NewReq("GET", "/test")
		_, err := client.Do(ctx, req)
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
		req := httpzrequest.NewReq("GET", "/test")
		resp, err := client.Do(ctx, req)
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
		req := &httpzrequest.Req{
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
				_, _ = client.Do(ctx, req)
			})
			return
		}

		resp, err := client.Do(ctx, req)

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
			requestPath:   "/absolute-url-test",
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
			requestPath:   "/relative-url-test",
			expectedBody:  "relative-url-response",
		})
	})

	t.Run("relative URL without client BaseUrl should fail", func(t *testing.T) {
		check(t, testScenario{
			name:          "relative URL without client BaseUrl should fail",
			requestPath:   "/relative-url-test",
			expectedError: "base url not found",
		})
	})

	t.Run("using both Path and Url should panic", func(t *testing.T) {
		check(t, testScenario{
			name:          "using both Path and Url should panic",
			requestUrl:    urlz.MustParse(s.Base().String()),
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
