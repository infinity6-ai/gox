package httpzserver_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

// TestUnitFilterAndHandlerInteraction is a comprehensive integration-style unit test
// that verifies the complex interactions between filters, response wrapping, and the final handler.
func TestUnitFilterAndHandlerInteraction(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()

	// First filter: wraps the request body and modifies response headers.
	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		preBody := strings.NewReader("reqpre-")
		originalBody := req.Body
		sufBody := strings.NewReader("-reqsuf")
		req.Body = io.MultiReader(preBody, originalBody, sufBody)

		nResp := func(status int, nHeaders http.Header) io.Writer {
			nHeaders.Set("a", "x1") // Set by filter 1
			nHeaders.Set("b", "x1") // Set by filter 1, should be overridden by filter 2
			return resp(status, nHeaders)
		}
		next(ctx, nResp, req)
	})

	// Second filter: uses WrapResponse to modify headers and inject content around the handler's output.
	s.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, next httpzserver.Handler) {
		w := next.WrapResponse(ctx, resp, req, func(outHeaders http.Header) int {
			outHeaders.Set("b", "x2") // Overrides header from filter 1
			outHeaders.Set("c", "x2") // Set by filter 2
			return 0
		}, func(outWriter io.Writer) io.Writer {
			_, err := outWriter.Write([]byte("UUUU "))
			errorz.Check(err)
			return outWriter
		})
		_, err := w.Write([]byte(" ZZZZ"))
		errorz.Check(err)
	})

	// Handler for the specific route being tested.
	s.AddHandler("POST", "/bla/{p1}/b/{p2}/c/*", func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
		reqBody, err := io.ReadAll(req.Body)
		errorz.Check(err)

		// Create a sorted string representation of params for stable output.
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sortedParams := make([]string, len(keys))
		for i, k := range keys {
			sortedParams[i] = fmt.Sprintf("%s:%s", k, params[k])
		}
		paramsStr := fmt.Sprintf("map[%s]", strings.Join(sortedParams, " "))

		body := fmt.Appendf(nil, "r: %s - %s, req: %s", req.Path, paramsStr, string(reqBody))

		w := resp(http.StatusBadRequest, http.Header{
			"a": []string{"y1"}, // Set by handler, should be overridden by filter 1
			"d": []string{"y"},  // Set by handler
		})
		_, err = w.Write(body)
		errorz.Check(err)
	})

	s.Start()

	// Perform the request
	resp, err := http.Post(s.Base()+"/bla/O1/b/O2/c/xyz", "text/plain", strings.NewReader("mybody"))
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Run("should return correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("should return correctly merged headers", func(t *testing.T) {
		// Handler sets "a":"y1", filter 1's wrapper overrides with "a":"x1".
		require.Equal(t, "x1", resp.Header.Get("a"))
		// Filter 2's wrapper sets "b":"x2", but filter 1's wrapper executes last and wins, setting "b":"x1".
		require.Equal(t, "x1", resp.Header.Get("b"))
		// Filter 2's wrapper sets "c":"x2".
		require.Equal(t, "x2", resp.Header.Get("c"))
		// Handler sets "d":"y", which is untouched by filters.
		require.Equal(t, "y", resp.Header.Get("d"))
	})

	t.Run("should return body with filter modifications", func(t *testing.T) {
		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		// Body should be wrapped by both filters and contain handler output.
		// Note: The original test had a non-deterministic map order. This is now fixed.
		expectedBody := "UUUU r: /bla/O1/b/O2/c/xyz - map[p1:O1 p2:O2], req: reqpre-mybody-reqsuf ZZZZ"
		require.Equal(t, expectedBody, string(data))
	})
}

func TestUnitRouting(t *testing.T) {
	type testScenario struct {
		name                string
		registeredMethod    string
		registeredPath      string
		requestMethod       string
		requestPath         string
		expectedStatus      int
		expectedBody        string
		expectedParams      map[string]string
		handlerShouldBeCalled bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		handlerCalled := false
		srv.AddHandler(s.registeredMethod, s.registeredPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzserver.Req, params map[string]string) {
			handlerCalled = true
			require.Equal(t, s.expectedParams, params)
			w := resp(s.expectedStatus, nil)
			_, err := w.Write([]byte(s.expectedBody))
			require.NoError(t, err)
		})

		srv.Start()

		req, err := http.NewRequest(s.requestMethod, srv.Base()+s.requestPath, nil)
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
			registeredMethod:    "GET",
			registeredPath:      "/users/find",
			requestMethod:       "GET",
			requestPath:         "/users/find",
			expectedStatus:      http.StatusOK,
			expectedBody:        "found",
			expectedParams:      map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Path with parameters", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:    "GET",
			registeredPath:      "/users/{id}/posts/{post_id}",
			requestMethod:       "GET",
			requestPath:         "/users/123/posts/abc",
			expectedStatus:      http.StatusOK,
			expectedBody:        "user 123, post abc",
			expectedParams:      map[string]string{"id": "123", "post_id": "abc"},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Prefix wildcard match", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:    "GET",
			registeredPath:      "/static/*",
			requestMethod:       "GET",
			requestPath:         "/static/css/style.css",
			expectedStatus:      http.StatusOK,
			expectedBody:        "static file",
			expectedParams:      map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Wildcard method match", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:    "*",
			registeredPath:      "/any_method",
			requestMethod:       "PUT",
			requestPath:         "/any_method",
			expectedStatus:      http.StatusOK,
			expectedBody:        "any method allowed",
			expectedParams:      map[string]string{},
			handlerShouldBeCalled: true,
		})
	})

	t.Run("Not Found", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:    "GET",
			registeredPath:      "/here",
			requestMethod:       "GET",
			requestPath:         "/not_here",
			expectedStatus:      http.StatusNotFound,
			expectedBody:        "Not Found",
			handlerShouldBeCalled: false,
		})
	})

	t.Run("Method mismatch", func(t *testing.T) {
		check(t, testScenario{
			registeredMethod:    "POST",
			registeredPath:      "/login",
			requestMethod:       "GET",
			requestPath:         "/login",
			expectedStatus:      http.StatusNotFound, // Routes are method-specific
			expectedBody:        "Not Found",
			handlerShouldBeCalled: false,
		})
	})
}

func TestUnitServerLifecycle(t *testing.T) {
	ctx := t.Context()

	t.Run("Listen and get address", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()

		require.Nil(t, s.Addr()) // Address should be nil before Listen
		s.Listen()
		require.NotNil(t, s.Addr())
		require.True(t, strings.HasPrefix(s.Addr().String(), "127.0.0.1:") || strings.HasPrefix(s.Addr().String(), "[::1]:"), "address should be local")
		
		addr, ok := s.Addr().(*net.TCPAddr)
		require.True(t, ok)
		require.NotZero(t, addr.Port, "port should not be zero after listen")

		base := s.Base()
		require.Equal(t, fmt.Sprintf("http://%s", s.Addr()), base)
	})

	t.Run("Double listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		s.Listen()
		require.Panics(t, func() {
			s.Listen()
		}, "calling Listen twice should panic")
	})

	t.Run("Start without listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		require.Panics(t, func() {
			s.Start()
		}, "calling Start before Listen should panic")
	})

	t.Run("Serve without listen panics", func(t *testing.T) {
		s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer s.Close()
		require.Panics(t, func() {
			s.Serve()
		}, "calling Serve before Listen should panic")
	})
}
