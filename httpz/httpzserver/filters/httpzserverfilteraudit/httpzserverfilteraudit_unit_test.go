package httpzserverfilteraudit_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/stretchr/testify/require"
)

func TestUnitAuditFilterScenarios(t *testing.T) {
	type testScenario struct {
		name                 string
		auditService         *logzaudit.AuditService
		requestMethod        string
		requestPath          string
		requestHeaders       map[string]string
		requestBody          []byte
		handlerStatus        int
		handlerHeaders       map[string]string
		handlerBody          []byte
		skipRespCall         bool
		modifyEntryInHandler func(entry *httpzserverfilteraudit.Entry)

		expectedStatus      int
		expectedReqId       string
		expectedServiceName string
		expectedReqMethod   string
		expectedReqPath     string
		expectedReqBodyLen  int
		expectedRespBodyLen int
		expectedRespStatus  int
		expectedLarge       bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		if s.auditService != nil {
			ctx = logz.SetAttrs(ctx, map[string]string{
				logz.CtxKeyAppName:    s.auditService.Name,
				logz.CtxKeyAppVersion: s.auditService.Version,
			})
		}
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector() // clear prior log entries

		srv.AddFilter(httpzserverfilteraudit.Filter)

		srv.AddHandler(s.requestMethod, s.requestPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			if s.modifyEntryInHandler != nil {
				entry := httpzserverfilteraudit.GetEntry(ctx)
				s.modifyEntryInHandler(entry)
			}

			if req.Body != nil {
				_, _ = io.ReadAll(req.Body)
			}

			if s.skipRespCall {
				return
			}

			var respH http.Header
			if len(s.handlerHeaders) > 0 {
				respH = make(http.Header)
				for k, v := range s.handlerHeaders {
					respH.Set(k, v)
				}
			}
			w := resp(s.handlerStatus, respH)
			if s.handlerBody != nil {
				_, err := w.Write(s.handlerBody)
				require.NoError(t, err)
			}
		})

		srv.Start()

		var bodyReader io.Reader
		if s.requestBody != nil {
			bodyReader = bytes.NewReader(s.requestBody)
		}

		clientReq, err := http.NewRequest(s.requestMethod, srv.Base().MustJoinPathString(s.requestPath).String(), bodyReader)
		require.NoError(t, err)
		for k, v := range s.requestHeaders {
			clientReq.Header.Set(k, v)
		}

		httpResp, err := http.DefaultClient.Do(clientReq)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		_, err = io.ReadAll(httpResp.Body)
		require.NoError(t, err)

		require.Equal(t, s.expectedStatus, httpResp.StatusCode)

		// Verify audit entry from collector
		entries := collector()
		require.Len(t, entries, 1)
		audit := entries[0].Audit
		require.NotNil(t, audit)
		reqResp := audit.ReqResp
		require.NotNil(t, reqResp)

		// Assert Service
		if s.auditService != nil {
			require.NotNil(t, reqResp.Service)
			require.Equal(t, s.expectedServiceName, reqResp.Service.Name)
		}

		// Assert Request ID
		if s.expectedReqId != "" {
			require.Equal(t, s.expectedReqId, reqResp.Id)
		} else {
			require.NotEmpty(t, reqResp.Id)
		}

		// Assert Request metadata
		require.NotNil(t, reqResp.Req)
		require.Equal(t, s.expectedReqMethod, reqResp.Req.Method)
		require.Equal(t, s.expectedReqPath, reqResp.Req.Path)
		require.Greater(t, reqResp.Req.CreatedAt, int64(0))

		if reqResp.Req.Body != nil {
			require.Equal(t, s.expectedReqBodyLen, reqResp.Req.Body.Size)
		}

		// Assert Response metadata
		require.NotNil(t, reqResp.Resp)
		require.Equal(t, s.expectedRespStatus, reqResp.Resp.Status)
		require.GreaterOrEqual(t, reqResp.Resp.CreatedAt, reqResp.Req.CreatedAt)

		if reqResp.Resp.Body != nil {
			require.Equal(t, s.expectedRespBodyLen, reqResp.Resp.Body.Size)
			require.Equal(t, s.expectedLarge, reqResp.Resp.Body.Large)
			if s.expectedLarge {
				require.NotEmpty(t, reqResp.Resp.Body.Tail)
			}
		}
	}

	t.Run("GET request without body populates metadata and service correctly", func(t *testing.T) {
		check(t, testScenario{
			auditService: &logzaudit.AuditService{
				Name:    "order-service",
				Version: "1.0.0",
			},
			requestMethod:       http.MethodGet,
			requestPath:         "/orders",
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("orders list"),
			expectedStatus:      http.StatusOK,
			expectedServiceName: "order-service",
			expectedReqMethod:   http.MethodGet,
			expectedReqPath:     "/orders",
			expectedReqBodyLen:  0,
			expectedRespBodyLen: len("orders list"),
			expectedRespStatus:  http.StatusOK,
		})
	})

	t.Run("POST request with body records both request and response payloads", func(t *testing.T) {
		reqPayload := []byte(`{"item":"book","qty":2}`)
		respPayload := []byte(`{"id":"order-123","status":"created"}`)

		check(t, testScenario{
			auditService: &logzaudit.AuditService{
				Name: "checkout-api",
			},
			requestMethod: http.MethodPost,
			requestPath:   "/orders",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "custom-req-id-777",
			},
			requestBody:         reqPayload,
			handlerStatus:       http.StatusCreated,
			handlerBody:         respPayload,
			expectedStatus:      http.StatusCreated,
			expectedReqId:       "custom-req-id-777",
			expectedServiceName: "checkout-api",
			expectedReqMethod:   http.MethodPost,
			expectedReqPath:     "/orders",
			expectedReqBodyLen:  len(reqPayload),
			expectedRespBodyLen: len(respPayload),
			expectedRespStatus:  http.StatusCreated,
		})
	})

	t.Run("Handler skips resp call does not crash and audits request", func(t *testing.T) {
		check(t, testScenario{
			auditService: &logzaudit.AuditService{
				Name: "auth-service",
			},
			requestMethod:       http.MethodGet,
			requestPath:         "/no-resp",
			skipRespCall:        true,
			expectedStatus:      http.StatusOK, // Default status in http.ResponseWriter if untouched
			expectedServiceName: "auth-service",
			expectedReqMethod:   http.MethodGet,
			expectedReqPath:     "/no-resp",
			expectedReqBodyLen:  0,
			expectedRespBodyLen: 0,
			expectedRespStatus:  0,
		})
	})

	t.Run("GetEntry retrieves Entry allowing custom manipulation", func(t *testing.T) {
		check(t, testScenario{
			auditService: &logzaudit.AuditService{
				Name: "billing-service",
			},
			requestMethod: http.MethodGet,
			requestPath:   "/billing",
			modifyEntryInHandler: func(entry *httpzserverfilteraudit.Entry) {
				entry.ReqResp.Id = "handler-injected-id"
			},
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("billing ok"),
			expectedStatus:      http.StatusOK,
			expectedReqId:       "handler-injected-id",
			expectedServiceName: "billing-service",
			expectedReqMethod:   http.MethodGet,
			expectedReqPath:     "/billing",
			expectedReqBodyLen:  0,
			expectedRespBodyLen: len("billing ok"),
			expectedRespStatus:  http.StatusOK,
		})
	})

	t.Run("Large response body marks Large true and captures tail", func(t *testing.T) {
		largeBody := []byte(strings.Repeat("A", logzaudit.LargeBodyThreashold+2048))

		check(t, testScenario{
			auditService: &logzaudit.AuditService{
				Name: "blob-service",
			},
			requestMethod:       http.MethodGet,
			requestPath:         "/blob",
			handlerStatus:       http.StatusOK,
			handlerBody:         largeBody,
			expectedStatus:      http.StatusOK,
			expectedServiceName: "blob-service",
			expectedReqMethod:   http.MethodGet,
			expectedReqPath:     "/blob",
			expectedReqBodyLen:  0,
			expectedRespBodyLen: len(largeBody),
			expectedRespStatus:  http.StatusOK,
			expectedLarge:       true,
		})
	})
}

func TestUnitAuditFilterEntryDirect(t *testing.T) {
	entry := &httpzserverfilteraudit.Entry{}

	// Verify CreateReqBody and CreateRespBody with nil produce nil safely
	require.Nil(t, entry.CreateReqBody(nil))
	require.Nil(t, entry.CreateRespBody(nil))

	// Verify BuildReqResp initializes empty bodies without panic
	reqResp := entry.BuildReqResp()
	require.NotNil(t, reqResp)
	require.NotNil(t, reqResp.Req)
	require.NotNil(t, reqResp.Resp)
	require.NotNil(t, reqResp.Req.Body)
	require.NotNil(t, reqResp.Resp.Body)
	require.Equal(t, 0, reqResp.Req.Body.Size)
	require.Equal(t, 0, reqResp.Resp.Body.Size)
}

func TestUnitAuditFilterContextPanics(t *testing.T) {
	t.Run("Filter panics when already registered in context", func(t *testing.T) {
		req := httpzrequest.MustFormat("GET", "/test", nil)

		defer func() {
			r := recover()
			require.NotNil(t, r)
			require.Contains(t, r.(error).Error(), "audit filter already registered in context")
		}()

		// Execute filter inside filter to simulate duplicate registration
		httpzserverfilteraudit.Filter(context.Background(), nil, req, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req) {
			httpzserverfilteraudit.Filter(ctx, resp, req, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req) {})
		})
	})
}
