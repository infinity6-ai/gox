package httpzserver_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitServerReqId(t *testing.T) {
	type testScenario struct {
		name                string
		requestHeaders      map[string]string
		handlerHeaders      map[string]string
		clientSentReqId     string
		preventHandlerReqId string
		repeatRequests      int
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		var filterReceivedReqId string
		srv.AddFilter(func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, next httpzserver.Handler) {
			filterReceivedReqId = req.Headers.Get(httpz.HeaderReqId)
			next(ctx, resp, req)
		})

		var handlerReceivedReqId string
		srv.AddHandler("GET", "/test-req-id", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			handlerReceivedReqId = req.Headers.Get(httpz.HeaderReqId)
			var respHeaders http.Header
			if len(s.handlerHeaders) > 0 {
				respHeaders = make(http.Header)
				for k, v := range s.handlerHeaders {
					respHeaders.Set(k, v)
				}
			}
			w := resp(http.StatusOK, respHeaders)
			_, err := w.Write([]byte("ok"))
			require.NoError(t, err)
		})

		srv.Start()

		count := s.repeatRequests
		if count < 1 {
			count = 1
		}
		seenReqIds := make(map[string]bool)

		for i := 0; i < count; i++ {
			req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/test-req-id").String(), nil)
			require.NoError(t, err)
			for k, v := range s.requestHeaders {
				req.Header.Set(k, v)
			}

			httpResp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			_, err = io.ReadAll(httpResp.Body)
			require.NoError(t, err)
			err = httpResp.Body.Close()
			require.NoError(t, err)

			respReqId := httpResp.Header.Get(httpz.HeaderReqId)

			require.NotEmpty(t, handlerReceivedReqId)
			require.Equal(t, handlerReceivedReqId, respReqId)
			require.Equal(t, handlerReceivedReqId, filterReceivedReqId)

			// idgen.B64() generates a 16-byte UUID raw URL base64 string (22 characters)
			decoded, err := base64.RawURLEncoding.DecodeString(respReqId)
			require.NoError(t, err)
			require.Len(t, decoded, 16)

			if s.clientSentReqId != "" {
				require.NotEqual(t, s.clientSentReqId, handlerReceivedReqId)
				require.NotEqual(t, s.clientSentReqId, respReqId)
			}

			if s.preventHandlerReqId != "" {
				require.NotEqual(t, s.preventHandlerReqId, respReqId)
			}

			require.False(t, seenReqIds[respReqId], "request ID %q should be unique", respReqId)
			seenReqIds[respReqId] = true
		}

		require.Len(t, seenReqIds, count)
	}

	t.Run("Auto-generated request ID when no header provided", func(t *testing.T) {
		check(t, testScenario{
			name:           "Auto-generated request ID when no header provided",
			requestHeaders: nil,
		})
	})

	t.Run("Client X-I6-Req-Id header is overridden by server", func(t *testing.T) {
		check(t, testScenario{
			name: "Client X-I6-Req-Id header is overridden by server",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "client-custom-req-id-12345",
			},
			clientSentReqId: "client-custom-req-id-12345",
		})
	})

	t.Run("Empty client X-I6-Req-Id header is overridden with generated ID", func(t *testing.T) {
		check(t, testScenario{
			name: "Empty client X-I6-Req-Id header is overridden with generated ID",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "",
			},
		})
	})

	t.Run("Handler response headers cannot override server-generated request ID", func(t *testing.T) {
		check(t, testScenario{
			name: "Handler response headers cannot override server-generated request ID",
			handlerHeaders: map[string]string{
				httpz.HeaderReqId: "handler-attempted-id-99999",
			},
			preventHandlerReqId: "handler-attempted-id-99999",
		})
	})

	t.Run("Both client header and handler header are overridden", func(t *testing.T) {
		check(t, testScenario{
			name: "Both client header and handler header are overridden",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "client-custom-req-id-11111",
			},
			handlerHeaders: map[string]string{
				httpz.HeaderReqId: "handler-attempted-id-22222",
			},
			clientSentReqId:     "client-custom-req-id-11111",
			preventHandlerReqId: "handler-attempted-id-22222",
		})
	})

	t.Run("Multiple requests generate distinct unique request IDs", func(t *testing.T) {
		check(t, testScenario{
			name: "Multiple requests generate distinct unique request IDs",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "fixed-client-req-id",
			},
			clientSentReqId: "fixed-client-req-id",
			repeatRequests:  5,
		})
	})
}
