package httpzclient_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientDefaultCheckRedirect(t *testing.T) {
	err := httpzclient.DefaultCheckRedirect(nil, nil)
	require.ErrorIs(t, err, http.ErrUseLastResponse)
}

func TestUnitClientRedirectHandling(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	var targetHits atomic.Int32

	s.AddHandler("GET", "/redirect-301", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusMovedPermanently, http.Header{
			"Location": []string{"/target"},
		})
		_, _ = w.Write([]byte("moved permanently"))
	})

	s.AddHandler("GET", "/redirect-302", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusFound, http.Header{
			"Location": []string{"/target"},
		})
		_, _ = w.Write([]byte("found"))
	})

	s.AddHandler("GET", "/redirect-303", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusSeeOther, http.Header{
			"Location": []string{"/target"},
		})
		_, _ = w.Write([]byte("see other"))
	})

	s.AddHandler("GET", "/redirect-307", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusTemporaryRedirect, http.Header{
			"Location": []string{"/target"},
		})
		_, _ = w.Write([]byte("temporary redirect"))
	})

	s.AddHandler("GET", "/redirect-308", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusPermanentRedirect, http.Header{
			"Location": []string{"/target"},
		})
		_, _ = w.Write([]byte("permanent redirect"))
	})

	s.AddHandler("GET", "/target", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		targetHits.Add(1)
		w := resp(http.StatusOK, http.Header{
			"X-Target": []string{"reached"},
		})
		_, _ = w.Write([]byte("target content"))
	})

	s.Listen()
	s.Start()

	type testScenario struct {
		path               string
		clientOpts         httpzclient.Options
		expectedStatusCode int
		expectedBody       string
		expectedLocation   string
		expectTargetHits   int32
		expectedErrMsg     string
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		initialTargetHits := targetHits.Load()

		opts := sc.clientOpts
		opts.BaseUrl = s.Base()
		client := httpzclient.New(ctx, opts)

		req := httpzrequest.New("GET", sc.path)
		resp, err := client.Do(ctx, req)

		if sc.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), sc.expectedErrMsg)
			return
		}

		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, sc.expectedStatusCode, resp.StatusCode)
		if sc.expectedLocation != "" {
			require.Equal(t, sc.expectedLocation, resp.Headers.Get("Location"))
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, sc.expectedBody, string(bodyBytes))

		require.Equal(t, initialTargetHits+sc.expectTargetHits, targetHits.Load())
	}

	t.Run("Default client does not follow 301 redirect", func(t *testing.T) {
		check(t, testScenario{
			path:               "/redirect-301",
			clientOpts:         httpzclient.Options{},
			expectedStatusCode: http.StatusMovedPermanently,
			expectedLocation:   "/target",
			expectedBody:       "moved permanently",
			expectTargetHits:   0,
		})
	})

	t.Run("Default client does not follow 302 redirect", func(t *testing.T) {
		check(t, testScenario{
			path:               "/redirect-302",
			clientOpts:         httpzclient.Options{},
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "/target",
			expectedBody:       "found",
			expectTargetHits:   0,
		})
	})

	t.Run("Default client does not follow 303 redirect", func(t *testing.T) {
		check(t, testScenario{
			path:               "/redirect-303",
			clientOpts:         httpzclient.Options{},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation:   "/target",
			expectedBody:       "see other",
			expectTargetHits:   0,
		})
	})

	t.Run("Default client does not follow 307 redirect", func(t *testing.T) {
		check(t, testScenario{
			path:               "/redirect-307",
			clientOpts:         httpzclient.Options{},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/target",
			expectedBody:       "temporary redirect",
			expectTargetHits:   0,
		})
	})

	t.Run("Default client does not follow 308 redirect", func(t *testing.T) {
		check(t, testScenario{
			path:               "/redirect-308",
			clientOpts:         httpzclient.Options{},
			expectedStatusCode: http.StatusPermanentRedirect,
			expectedLocation:   "/target",
			expectedBody:       "permanent redirect",
			expectTargetHits:   0,
		})
	})

	t.Run("Client with FollowRedirects follows redirect", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-302",
			clientOpts: httpzclient.Options{
				FollowRedirects: true,
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "target content",
			expectTargetHits:   1,
		})
	})

	t.Run("Client with custom CheckRedirect error halts redirect", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-302",
			clientOpts: httpzclient.Options{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return errors.New("custom redirect rejected")
				},
			},
			expectedErrMsg: "custom redirect rejected",
		})
	})

	t.Run("Client with custom GetClient does not follow redirect by default", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-302",
			clientOpts: httpzclient.Options{
				GetClient: func(ctx context.Context) *http.Client {
					return &http.Client{}
				},
			},
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "/target",
			expectedBody:       "found",
			expectTargetHits:   0,
		})
	})

	t.Run("Client with custom GetClient and FollowRedirects follows redirect", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-302",
			clientOpts: httpzclient.Options{
				FollowRedirects: true,
				GetClient: func(ctx context.Context) *http.Client {
					return &http.Client{}
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "target content",
			expectTargetHits:   1,
		})
	})
}
