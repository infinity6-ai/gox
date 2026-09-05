package httpzserver_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitRouting(t *testing.T) {
	type testScenario struct {
		name                  string
		registeredMethod      string
		registeredPath        string
		requestMethod         string
		requestPath           string
		expectedStatus        int
		expectedBody          string
		expectedParams        map[string]string
		handlerShouldBeCalled bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		handlerCalled := false
		srv.AddHandler(s.registeredMethod, s.registeredPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			handlerCalled = true
			require.Equal(t, s.expectedParams, params)
			w := resp(s.expectedStatus, nil)
			_, err := w.Write([]byte(s.expectedBody))
			require.NoError(t, err)
		})

		srv.Start()

		req, err := http.NewRequest(s.requestMethod, srv.Base().MustJoinPathString(s.requestPath).String(), nil)
		require.NoError(t, err)

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.handlerShouldBeCalled, handlerCalled, "handler execution status mismatch")
		require.Equal(t, s.expectedStatus, httpResp.StatusCode)

		body, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)
		require.Equal(t, s.expectedBody, string(body))
	}

	t.Run("Exact match", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "GET",
			registeredPath:        "/users/find",
			requestMethod:         "GET",
			requestPath:           "/users/find",
			expectedStatus:        http.StatusOK,
			expectedBody:          "found",
			expectedParams:        map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Path with parameters", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "GET",
			registeredPath:        "/users/{id}/posts/{post_id}",
			requestMethod:         "GET",
			requestPath:           "/users/123/posts/abc",
			expectedStatus:        http.StatusOK,
			expectedBody:          "user 123, post abc",
			expectedParams:        map[string]string{"id": "123", "post_id": "abc"},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Prefix wildcard match", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "GET",
			registeredPath:        "/static/*",
			requestMethod:         "GET",
			requestPath:           "/static/css/style.css",
			expectedStatus:        http.StatusOK,
			expectedBody:          "static file",
			expectedParams:        map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Wildcard method match", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "*",
			registeredPath:        "/any_method",
			requestMethod:         "PUT",
			requestPath:           "/any_method",
			expectedStatus:        http.StatusOK,
			expectedBody:          "any method allowed",
			expectedParams:        map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Not Found", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "GET",
			registeredPath:        "/here",
			requestMethod:         "GET",
			requestPath:           "/not_here",
			expectedStatus:        http.StatusNotFound,
			expectedBody:          "Not Found",
			handlerShouldBeCalled: false,
		})
	})

	t.Run("Method mismatch", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:      "POST",
			registeredPath:        "/login",
			requestMethod:         "GET",
			requestPath:           "/login",
			expectedStatus:        http.StatusNotFound, // Routes are method-specific
			expectedBody:          "Not Found",
			handlerShouldBeCalled: false,
		})
	})
}
