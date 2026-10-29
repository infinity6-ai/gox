package httpzclient_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientMustSuccess(t *testing.T) {
	ctx := t.Context()

	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("GET", "/200", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("success-response"))
	})

	s.AddHandler("GET", "/404", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusNotFound, nil)
		_, _ = w.Write([]byte("resource-not-found"))
	})

	s.AddHandler("POST", "/500", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusInternalServerError, nil)
		_, _ = w.Write([]byte("server-failure-message"))
	})

	s.AddHandler("GET", "/long-error", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusBadRequest, nil)
		_, _ = w.Write([]byte(strings.Repeat("a", 2000)))
	})

	s.Listen()
	s.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})

	type testScenario struct {
		method         string
		path           string
		reqBody        io.Reader
		expectPanic    bool
		expectedErrSub string
		expectedResp   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		req := httpzrequest.New(s.method, s.path)
		if s.reqBody != nil {
			req.SetBody(s.reqBody)
		}

		if s.expectPanic {
			require.PanicsWithError(t, s.expectedErrSub, func() {
				client.MustSuccess(ctx, req)
			})
			return
		}

		var resp *httpzclient.Resp
		require.NotPanics(t, func() {
			resp = client.MustSuccess(ctx, req)
		})
		require.NotNil(t, resp)
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, s.expectedResp, string(bodyBytes))
	}

	t.Run("200 OK returns response successfully", func(t *testing.T) {
		check(t, testScenario{
			method:       "GET",
			path:         "/200",
			expectPanic:  false,
			expectedResp: "success-response",
		})
	})

	t.Run("404 with nil request body panics with status and response head body", func(t *testing.T) {
		check(t, testScenario{
			method:         "GET",
			path:           "/404",
			expectPanic:    true,
			expectedErrSub: "http response error: 404, head body: 'resource-not-found'",
		})
	})

	t.Run("500 with request body panics with status and response head body", func(t *testing.T) {
		check(t, testScenario{
			method:         "POST",
			path:           "/500",
			reqBody:        strings.NewReader("request-payload"),
			expectPanic:    true,
			expectedErrSub: "http response error: 500, head body: 'server-failure-message'",
		})
	})

	t.Run("400 with long body truncates head body to 1024 bytes", func(t *testing.T) {
		check(t, testScenario{
			method:         "GET",
			path:           "/long-error",
			expectPanic:    true,
			expectedErrSub: fmt.Sprintf("http response error: 400, head body: '%s'", strings.Repeat("a", 1024)),
		})
	})
}

func TestUnitClientMustDo(t *testing.T) {
	ctx := t.Context()

	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("GET", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, nil)
		_, _ = w.Write([]byte("ok"))
	})
	s.Listen()
	s.Start()

	t.Run("Valid request succeeds without panic", func(t *testing.T) {
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})
		req := httpzrequest.New("GET", "/test")
		var resp *httpzclient.Resp
		require.NotPanics(t, func() {
			resp = client.MustDo(ctx, req)
		})
		require.NotNil(t, resp)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Request to non-existent server panics", func(t *testing.T) {
		sClosed := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		sClosed.Listen()
		base := sClosed.Base()
		sClosed.Close()

		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: base})
		req := httpzrequest.New("GET", "/test")
		require.Panics(t, func() {
			client.MustDo(ctx, req)
		})
	})
}
