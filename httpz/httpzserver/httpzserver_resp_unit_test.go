package httpzserver_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

type failWriter struct{}

func (failWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("simulated write error")
}

func TestUnitRespRedirect(t *testing.T) {
	type testScenario struct {
		targetURL          string
		initialHeaders     http.Header
		expectedStatusCode int
		expectedLocation   string
		expectedHeaders    map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		var recordedStatus int
		var recordedHeaders http.Header
		var buf bytes.Buffer

		var resp httpzserver.Resp = func(status int, headers http.Header) io.Writer {
			recordedStatus = status
			recordedHeaders = headers
			return &buf
		}

		resp.Redirect(s.initialHeaders, s.targetURL)

		require.Equal(t, s.expectedStatusCode, recordedStatus)
		require.NotNil(t, recordedHeaders)
		require.Equal(t, s.expectedLocation, recordedHeaders.Get("Location"))

		for k, v := range s.expectedHeaders {
			require.Equal(t, v, recordedHeaders.Get(k))
		}
	}

	t.Run("Redirect with nil headers sets 307 and Location", func(t *testing.T) {
		check(t, testScenario{
			targetURL:          "/new-location",
			initialHeaders:     nil,
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/new-location",
		})
	})

	t.Run("Redirect with empty headers sets 307 and Location", func(t *testing.T) {
		check(t, testScenario{
			targetURL:          "/empty-headers",
			initialHeaders:     http.Header{},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/empty-headers",
		})
	})

	t.Run("Redirect with existing custom headers preserves them", func(t *testing.T) {
		check(t, testScenario{
			targetURL: "https://example.com/target",
			initialHeaders: http.Header{
				"X-Custom-Header": []string{"custom-value"},
				"X-Trace":         []string{"t-1"},
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "https://example.com/target",
			expectedHeaders: map[string]string{
				"X-Custom-Header": "custom-value",
				"X-Trace":         "t-1",
			},
		})
	})

	t.Run("Redirect overrides prior Location header", func(t *testing.T) {
		check(t, testScenario{
			targetURL: "https://example.com/final",
			initialHeaders: http.Header{
				"Location": []string{"https://example.com/old"},
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "https://example.com/final",
		})
	})

	t.Run("Redirect with URL containing query parameters and fragment", func(t *testing.T) {
		check(t, testScenario{
			targetURL:          "/search?q=gox&limit=10#results",
			initialHeaders:     nil,
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/search?q=gox&limit=10#results",
		})
	})
}

func TestUnitRespRedirectJS(t *testing.T) {
	type testScenario struct {
		targetURL           string
		initialHeaders      http.Header
		expectedStatusCode  int
		expectedContentType string
		expectedBody        string
		expectedHeaders     map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		var recordedStatus int
		var recordedHeaders http.Header
		var buf bytes.Buffer

		var resp httpzserver.Resp = func(status int, headers http.Header) io.Writer {
			recordedStatus = status
			recordedHeaders = headers
			return &buf
		}

		resp.RedirectJS(s.initialHeaders, s.targetURL)

		require.Equal(t, s.expectedStatusCode, recordedStatus)
		require.NotNil(t, recordedHeaders)
		require.Equal(t, s.expectedContentType, recordedHeaders.Get("Content-Type"))
		require.Equal(t, s.expectedBody, buf.String())

		for k, v := range s.expectedHeaders {
			require.Equal(t, v, recordedHeaders.Get(k))
		}
	}

	t.Run("RedirectJS with nil headers sets 200, Content-Type, and HTML body", func(t *testing.T) {
		check(t, testScenario{
			targetURL:           "/destination",
			initialHeaders:      nil,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/destination'"></body></html>`,
		})
	})

	t.Run("RedirectJS with empty headers sets 200, Content-Type, and HTML body", func(t *testing.T) {
		check(t, testScenario{
			targetURL:           "/empty-headers-dest",
			initialHeaders:      http.Header{},
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/empty-headers-dest'"></body></html>`,
		})
	})

	t.Run("RedirectJS with existing custom headers preserves them", func(t *testing.T) {
		check(t, testScenario{
			targetURL: "https://example.com/dashboard",
			initialHeaders: http.Header{
				"X-Trace-Id": []string{"trace-123"},
			},
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='https://example.com/dashboard'"></body></html>`,
			expectedHeaders: map[string]string{
				"X-Trace-Id": "trace-123",
			},
		})
	})

	t.Run("RedirectJS overrides prior Content-Type", func(t *testing.T) {
		check(t, testScenario{
			targetURL: "/login",
			initialHeaders: http.Header{
				"Content-Type": []string{"application/json"},
			},
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/login'"></body></html>`,
		})
	})

	t.Run("RedirectJS with relative URL sets expected body", func(t *testing.T) {
		check(t, testScenario{
			targetURL:           "../login?return_to=/app",
			initialHeaders:      nil,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='../login?return_to=/app'"></body></html>`,
		})
	})

	t.Run("RedirectJS with query params and fragment sets expected body", func(t *testing.T) {
		check(t, testScenario{
			targetURL:           "/oauth/callback?code=xyz&state=abc#token",
			initialHeaders:      nil,
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/oauth/callback?code=xyz&state=abc#token'"></body></html>`,
		})
	})
}

func TestUnitRespRedirectJSWriteError(t *testing.T) {
	type testScenario struct {
		targetURL string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		var resp httpzserver.Resp = func(status int, headers http.Header) io.Writer {
			return failWriter{}
		}

		require.Panics(t, func() {
			resp.RedirectJS(nil, s.targetURL)
		})
	}

	t.Run("RedirectJS panics via errorz.Check when writer fails", func(t *testing.T) {
		check(t, testScenario{
			targetURL: "/error-destination",
		})
	})
}

func TestUnitRespServerIntegration(t *testing.T) {
	type testScenario struct {
		path                 string
		handlerAction        func(resp httpzserver.Resp)
		followRedirects      bool
		expectedStatusCode   int
		expectedLocation     string
		expectedContentType  string
		expectedBody         string
		expectedExtraHeaders map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddHandler("GET", s.path, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			s.handlerAction(resp)
		})

		srv.AddHandler("GET", "/final-destination", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			w := resp(http.StatusOK, nil)
			_, err := w.Write([]byte("destination-reached"))
			require.NoError(t, err)
		})

		srv.Start()

		client := &http.Client{}
		if !s.followRedirects {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString(s.path).String(), nil)
		require.NoError(t, err)

		httpResp, err := client.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedStatusCode, httpResp.StatusCode)

		if s.expectedLocation != "" {
			require.Equal(t, s.expectedLocation, httpResp.Header.Get("Location"))
		}
		if s.expectedContentType != "" {
			require.Equal(t, s.expectedContentType, httpResp.Header.Get("Content-Type"))
		}
		for k, v := range s.expectedExtraHeaders {
			require.Equal(t, v, httpResp.Header.Get(k))
		}
		if s.expectedBody != "" {
			bodyBytes, err := io.ReadAll(httpResp.Body)
			require.NoError(t, err)
			require.Equal(t, s.expectedBody, string(bodyBytes))
		}
	}

	t.Run("Server responds to Redirect with 307 and Location header (nil headers)", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-nil",
			handlerAction: func(resp httpzserver.Resp) {
				resp.Redirect(nil, "/target-page")
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/target-page",
		})
	})

	t.Run("Server responds to Redirect with custom headers preserved", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-custom",
			handlerAction: func(resp httpzserver.Resp) {
				resp.Redirect(http.Header{"X-Redirect-Reason": []string{"moved-temporarily"}}, "/target-page")
			},
			expectedStatusCode: http.StatusTemporaryRedirect,
			expectedLocation:   "/target-page",
			expectedExtraHeaders: map[string]string{
				"X-Redirect-Reason": "moved-temporarily",
			},
		})
	})

	t.Run("Server responds to RedirectJS with 200, Content-Type, and HTML body (nil headers)", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-js-nil",
			handlerAction: func(resp httpzserver.Resp) {
				resp.RedirectJS(nil, "/target-page")
			},
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/target-page'"></body></html>`,
		})
	})

	t.Run("Server responds to RedirectJS with custom headers preserved", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-js-custom",
			handlerAction: func(resp httpzserver.Resp) {
				resp.RedirectJS(http.Header{"X-Server-Version": []string{"2.0.0"}}, "/target-page")
			},
			expectedStatusCode:  http.StatusOK,
			expectedContentType: "text/html; charset=UTF-8",
			expectedBody:        `<html><body onload="window.location.href='/target-page'"></body></html>`,
			expectedExtraHeaders: map[string]string{
				"X-Server-Version": "2.0.0",
			},
		})
	})

	t.Run("Client following redirects successfully reaches target from Redirect", func(t *testing.T) {
		check(t, testScenario{
			path: "/redirect-follow",
			handlerAction: func(resp httpzserver.Resp) {
				resp.Redirect(nil, "/final-destination")
			},
			followRedirects:    true,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "destination-reached",
		})
	})
}
