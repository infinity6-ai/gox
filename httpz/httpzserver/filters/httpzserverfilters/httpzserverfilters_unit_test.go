package httpzserverfilters_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzctx"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilters"
	"github.com/stretchr/testify/require"
)

func TestUnitBasicFilters(t *testing.T) {
	type testScenario struct {
		app             string
		version         string
		useConfig       bool
		expectedApp     string
		expectedVersion string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector()

		if s.useConfig {
			httpzserverfilters.Config(srv, s.app, s.version)
		} else {
			for _, f := range httpzserverfilters.BasicFilters(s.app, s.version) {
				srv.AddFilter(f)
			}
		}

		var capturedAppName, capturedAppVersion string
		srv.AddHandler(http.MethodGet, "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			m := logzctx.Get(ctx)
			capturedAppName = m[logz.CtxKeyAppName]
			capturedAppVersion = m[logz.CtxKeyAppVersion]
			w := resp(http.StatusOK, nil)
			_, _ = w.Write([]byte("ok"))
		})

		srv.Start()

		resp, err := http.Get(srv.Base().MustJoinPathString("/test").String())
		require.NoError(t, err)
		defer resp.Body.Close()
		_, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, s.expectedApp, capturedAppName)
		require.Equal(t, s.expectedVersion, capturedAppVersion)

		entries := collector()
		require.Len(t, entries, 1)
		reqResp := entries[0].Audit.ReqResp
		require.NotNil(t, reqResp.Service)
		require.Equal(t, s.expectedApp, reqResp.Service.Name)
		require.Equal(t, s.expectedVersion, reqResp.Service.Version)
	}

	t.Run("BasicFilters injects app and version into request context", func(t *testing.T) {
		check(t, testScenario{
			app:             "test-app",
			version:         "1.2.3",
			useConfig:       false,
			expectedApp:     "test-app",
			expectedVersion: "1.2.3",
		})
	})

	t.Run("Config configures server context and basic filters with app and version", func(t *testing.T) {
		check(t, testScenario{
			app:             "configured-app",
			version:         "2.0.0",
			useConfig:       true,
			expectedApp:     "configured-app",
			expectedVersion: "2.0.0",
		})
	})
}
