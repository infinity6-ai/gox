package httpzhelper_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/internal/httpzhelper"
	"github.com/stretchr/testify/require"
)

func TestUnitFromHttpRequest(t *testing.T) {
	type testScenario struct {
		name         string
		httpReq      *http.Request
		expectPanic  bool
		expectedPath string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		output := &httpzrequest.Req{}

		if s.expectPanic {
			require.Panics(t, func() {
				httpzhelper.FromHttpRequest(s.httpReq, output)
			})
			return
		}

		httpzhelper.FromHttpRequest(s.httpReq, output)
		require.Equal(t, s.httpReq.Method, output.Method)
		require.Equal(t, s.expectedPath, output.Path.String())
		require.Equal(t, s.httpReq.Header, output.Headers)
		require.Equal(t, s.httpReq.URL.Query(), output.Query)

		if s.httpReq.Body != nil {
			b, err := io.ReadAll(output.Body)
			require.NoError(t, err)
			require.Equal(t, "payload", string(b))
		}
	}

	t.Run("Valid request with absolute path", func(t *testing.T) {
		u, err := url.Parse("http://example.com/api/v1/resource?a=1&b=2")
		require.NoError(t, err)

		req := &http.Request{
			Method: http.MethodPost,
			URL:    u,
			Header: http.Header{"Authorization": []string{"Bearer token"}},
			Body:   io.NopCloser(strings.NewReader("payload")),
		}

		check(t, testScenario{
			name:         "Valid request with absolute path",
			httpReq:      req,
			expectPanic:  false,
			expectedPath: "/api/v1/resource",
		})
	})

	t.Run("Relative path panics", func(t *testing.T) {
		u, err := url.Parse("api/v1/resource")
		require.NoError(t, err)

		req := &http.Request{
			Method: http.MethodGet,
			URL:    u,
		}

		check(t, testScenario{
			name:        "Relative path panics",
			httpReq:     req,
			expectPanic: true,
		})
	})
}
