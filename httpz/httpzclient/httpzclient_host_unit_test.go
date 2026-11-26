package httpzclient_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientHostHeader(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	var receivedHost atomic.Pointer[string]

	s.AddHandler("GET", "/test-host", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		host := req.Headers.Get("Host")
		receivedHost.Store(&host)

		w := resp(http.StatusOK, http.Header{})
		_, _ = w.Write([]byte("ok"))
	})
	s.Listen()
	s.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})

	type testScenario struct {
		headers      map[string]string
		rawHeaders   http.Header
		expectedHost string
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		empty := ""
		receivedHost.Store(&empty)

		req := httpzrequest.New("GET", "/test-host")
		for k, v := range sc.headers {
			req.AddHeader(k, v)
		}
		for k, v := range sc.rawHeaders {
			req.Headers[k] = v
		}

		resp, err := client.Do(ctx, req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, sc.expectedHost, *receivedHost.Load())
	}

	t.Run("Explicit Host header overrides request host", func(t *testing.T) {
		check(t, testScenario{
			headers:      map[string]string{"Host": "api.example.com"},
			expectedHost: "api.example.com",
		})
	})

	t.Run("Lowercase host header is recognized case-insensitively", func(t *testing.T) {
		check(t, testScenario{
			rawHeaders:   http.Header{"host": []string{"custom-domain.org:8080"}},
			expectedHost: "custom-domain.org:8080",
		})
	})

	t.Run("Default without Host header sends server address as host", func(t *testing.T) {
		check(t, testScenario{
			expectedHost: s.Base().Host,
		})
	})
}
