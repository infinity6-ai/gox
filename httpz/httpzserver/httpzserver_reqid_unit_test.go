package httpzserver_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitServerReqId(t *testing.T) {
	type testScenario struct {
		name              string
		requestHeaders    map[string]string
		expectedReqId     string
		validateGenerated bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		var handlerReceivedReqId string
		srv.AddHandler("GET", "/test-req-id", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			handlerReceivedReqId = req.Headers.Get(httpz.HeaderReqId)
			w := resp(http.StatusOK, nil)
			_, err := w.Write([]byte("ok"))
			require.NoError(t, err)
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/test-req-id").String(), nil)
		require.NoError(t, err)
		for k, v := range s.requestHeaders {
			req.Header.Set(k, v)
		}

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		respReqId := httpResp.Header.Get(httpz.HeaderReqId)

		if s.validateGenerated {
			require.NotEmpty(t, handlerReceivedReqId)
			require.Equal(t, handlerReceivedReqId, respReqId)
			// idgen.B64() generates a 16-byte UUID raw URL base64 string (22 characters)
			decoded, err := base64.RawURLEncoding.DecodeString(respReqId)
			require.NoError(t, err)
			require.Len(t, decoded, 16)
		} else {
			require.Equal(t, s.expectedReqId, handlerReceivedReqId)
			require.Equal(t, s.expectedReqId, respReqId)
		}
	}

	t.Run("Auto-generated request ID when no header provided", func(t *testing.T) {
		check(t, testScenario{
			name:              "Auto-generated request ID when no header provided",
			requestHeaders:    nil,
			validateGenerated: true,
		})
	})

	t.Run("Custom X-I6-Req-Id header preserved and propagated to response", func(t *testing.T) {
		check(t, testScenario{
			name: "Custom X-I6-Req-Id header preserved and propagated to response",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "custom-req-id-12345",
			},
			expectedReqId:     "custom-req-id-12345",
			validateGenerated: false,
		})
	})

	t.Run("Custom X-Request-Id header preserved and propagated as X-I6-Req-Id", func(t *testing.T) {
		check(t, testScenario{
			name: "Custom X-Request-Id header preserved and propagated as X-I6-Req-Id",
			requestHeaders: map[string]string{
				"X-Request-Id": "legacy-req-id-67890",
			},
			expectedReqId:     "legacy-req-id-67890",
			validateGenerated: false,
		})
	})
}
