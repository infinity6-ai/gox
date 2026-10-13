package httpzserverfilteraudit_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/stretchr/testify/require"
)

func TestUnitAuditFilterBasic(t *testing.T) {
	type testScenario struct {
		name                string
		registerMethod      string
		registerPath        string
		requestMethod       string
		requestPath         string
		requestQuery        string
		requestHeaders      map[string]string
		requestBody         []byte
		handlerStatus       int
		handlerHeaders      map[string]string
		handlerBody         []byte
		handlerReadReqBody  bool
		service             *logzaudit.AuditService
		expectedStatus      int
		expectedRespBody    []byte
		expectedReqBodyNil  bool
		expectedRespBodyNil bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector() // clear prior entries

		if s.service != nil {
			srv.AddFilter(httpzserverfilteraudit.New(httpzserverfilteraudit.Options{
				Service: s.service,
			}))
		} else {
			srv.AddFilter(httpzserverfilteraudit.AuditFilter)
		}

		if s.registerMethod != "" {
			srv.AddHandler(s.registerMethod, s.registerPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
				if s.handlerReadReqBody && req.Body != nil {
					_, err := io.ReadAll(req.Body)
					require.NoError(t, err)
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
		}

		srv.Start()

		var bodyReader io.Reader
		if s.requestBody != nil {
			bodyReader = bytes.NewReader(s.requestBody)
		}
		urlStr := srv.Base().MustJoinPathString(s.requestPath).String() + s.requestQuery
		req, err := http.NewRequest(s.requestMethod, urlStr, bodyReader)
		require.NoError(t, err)
		for k, v := range s.requestHeaders {
			req.Header.Set(k, v)
		}

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		respBytes, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)
		require.Equal(t, s.expectedStatus, httpResp.StatusCode)
		require.Equal(t, s.expectedRespBody, respBytes)

		entries := collector()
		require.Len(t, entries, 1)
		require.Equal(t, "audit-req-resp", entries[0].Operation)

		audit := entries[0].Audit.ReqResp
		require.NotNil(t, audit)
		require.NotEmpty(t, audit.Id)

		// Verify custom request ID propagation if provided
		if customID, ok := s.requestHeaders[httpz.HeaderReqId]; ok {
			require.Equal(t, customID, audit.Id)
		} else if customID, ok := s.requestHeaders["X-Request-Id"]; ok {
			require.Equal(t, customID, audit.Id)
		}

		// Verify service
		if s.service != nil {
			require.NotNil(t, audit.Service)
			require.Equal(t, s.service.Name, audit.Service.Name)
			require.Equal(t, s.service.Version, audit.Service.Version)
		}

		// Verify Request audit
		require.NotNil(t, audit.Req)
		require.Greater(t, audit.Req.CreatedAt, int64(0))
		require.Equal(t, s.requestMethod, audit.Req.Method)
		require.Equal(t, s.requestPath, audit.Req.Path)

		if s.expectedReqBodyNil {
			require.Nil(t, audit.Req.Body)
		} else {
			require.NotNil(t, audit.Req.Body)
			require.Equal(t, len(s.requestBody), audit.Req.Body.Size)
			require.False(t, audit.Req.Body.Large)
			if len(s.requestBody) > 0 {
				require.Equal(t, s.requestBody, audit.Req.Body.Payload)
			} else {
				require.Nil(t, audit.Req.Body.Payload)
			}
			require.Nil(t, audit.Req.Body.Tail)
		}

		// Verify Response audit
		require.NotNil(t, audit.Resp)
		require.Greater(t, audit.Resp.CreatedAt, int64(0))
		require.GreaterOrEqual(t, audit.Resp.CreatedAt, audit.Req.CreatedAt)
		require.Equal(t, s.expectedStatus, audit.Resp.Status)

		if s.expectedRespBodyNil {
			require.Nil(t, audit.Resp.Body)
		} else {
			require.NotNil(t, audit.Resp.Body)
			require.Equal(t, len(s.expectedRespBody), audit.Resp.Body.Size)
			require.False(t, audit.Resp.Body.Large)
			if len(s.expectedRespBody) > 0 {
				require.Equal(t, s.expectedRespBody, audit.Resp.Body.Payload)
			} else {
				require.Nil(t, audit.Resp.Body.Payload)
			}
			require.Nil(t, audit.Resp.Body.Tail)
		}
	}

	t.Run("GET request without body returns 200 with response body", func(t *testing.T) {
		check(t, testScenario{
			registerMethod:      "GET",
			registerPath:        "/users/find",
			requestMethod:       "GET",
			requestPath:         "/users/find",
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("found"),
			expectedStatus:      http.StatusOK,
			expectedRespBody:    []byte("found"),
			expectedReqBodyNil:  true,
			expectedRespBodyNil: false,
		})
	})

	t.Run("POST request with small body read by handler", func(t *testing.T) {
		check(t, testScenario{
			registerMethod:      "POST",
			registerPath:        "/items",
			requestMethod:       "POST",
			requestPath:         "/items",
			requestBody:         []byte("item-data"),
			handlerStatus:       http.StatusCreated,
			handlerBody:         []byte("item-created"),
			handlerReadReqBody:  true,
			expectedStatus:      http.StatusCreated,
			expectedRespBody:    []byte("item-created"),
			expectedReqBodyNil:  false,
			expectedRespBodyNil: false,
		})
	})

	t.Run("POST request with body NOT read by handler is drained and captured", func(t *testing.T) {
		check(t, testScenario{
			registerMethod:      "POST",
			registerPath:        "/ignore-body",
			requestMethod:       "POST",
			requestPath:         "/ignore-body",
			requestBody:         []byte("unread-payload"),
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("ok"),
			handlerReadReqBody:  false,
			expectedStatus:      http.StatusOK,
			expectedRespBody:    []byte("ok"),
			expectedReqBodyNil:  false,
			expectedRespBodyNil: false,
		})
	})

	t.Run("Request with query parameters and custom headers", func(t *testing.T) {
		check(t, testScenario{
			registerMethod: "GET",
			registerPath:   "/search",
			requestMethod:  "GET",
			requestPath:    "/search",
			requestQuery:   "?q=test&limit=10",
			requestHeaders: map[string]string{
				"X-Custom-Header": "header-value",
				"X-Request-Id":    "req-uuid-12345",
			},
			handlerStatus:       http.StatusOK,
			handlerHeaders:      map[string]string{"X-Server-Resp": "resp-value"},
			handlerBody:         []byte("search-results"),
			expectedStatus:      http.StatusOK,
			expectedRespBody:    []byte("search-results"),
			expectedReqBodyNil:  true,
			expectedRespBodyNil: false,
		})
	})

	t.Run("Request with X-I6-Req-Id header propagates to audit log", func(t *testing.T) {
		check(t, testScenario{
			registerMethod: "GET",
			registerPath:   "/items",
			requestMethod:  "GET",
			requestPath:    "/items",
			requestHeaders: map[string]string{
				httpz.HeaderReqId: "custom-i6-req-id-789",
			},
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("items-list"),
			expectedStatus:      http.StatusOK,
			expectedRespBody:    []byte("items-list"),
			expectedReqBodyNil:  true,
			expectedRespBodyNil: false,
		})
	})

	t.Run("Route not found returns 404", func(t *testing.T) {
		check(t, testScenario{
			registerMethod:      "GET",
			registerPath:        "/existing",
			requestMethod:       "GET",
			requestPath:         "/not-existing",
			expectedStatus:      http.StatusNotFound,
			expectedRespBody:    []byte("Not Found"),
			expectedReqBodyNil:  true,
			expectedRespBodyNil: false,
		})
	})

	t.Run("204 No Content response has nil body", func(t *testing.T) {
		check(t, testScenario{
			registerMethod:      "DELETE",
			registerPath:        "/delete-item",
			requestMethod:       "DELETE",
			requestPath:         "/delete-item",
			handlerStatus:       http.StatusNoContent,
			expectedStatus:      http.StatusNoContent,
			expectedRespBody:    []byte{},
			expectedReqBodyNil:  true,
			expectedRespBodyNil: true,
		})
	})

	t.Run("Custom service options configured", func(t *testing.T) {
		check(t, testScenario{
			registerMethod: "GET",
			registerPath:   "/service-check",
			requestMethod:  "GET",
			requestPath:    "/service-check",
			service: &logzaudit.AuditService{
				Name:    "payment-gateway",
				Version: "v2.1.0",
			},
			handlerStatus:       http.StatusOK,
			handlerBody:         []byte("service-ok"),
			expectedStatus:      http.StatusOK,
			expectedRespBody:    []byte("service-ok"),
			expectedReqBodyNil:  true,
			expectedRespBodyNil: false,
		})
	})
}

func TestUnitAuditFilterLargeBody(t *testing.T) {
	type testScenario struct {
		name               string
		requestBody        []byte
		responseBody       []byte
		handlerReadReqBody bool
		streamResponse     bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector()

		srv.AddFilter(httpzserverfilteraudit.AuditFilter)

		srv.AddHandler("POST", "/data", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			if s.handlerReadReqBody && req.Body != nil {
				_, err := io.ReadAll(req.Body)
				require.NoError(t, err)
			}
			w := resp(http.StatusOK, nil)
			if s.streamResponse && len(s.responseBody) > 0 {
				chunkSize := 512
				for i := 0; i < len(s.responseBody); i += chunkSize {
					end := min(i+chunkSize, len(s.responseBody))
					_, err := w.Write(s.responseBody[i:end])
					require.NoError(t, err)
				}
			} else if s.responseBody != nil {
				_, err := w.Write(s.responseBody)
				require.NoError(t, err)
			}
		})

		srv.Start()

		var bodyReader io.Reader
		if s.requestBody != nil {
			bodyReader = bytes.NewReader(s.requestBody)
		}
		req, err := http.NewRequest(http.MethodPost, srv.Base().MustJoinPathString("/data").String(), bodyReader)
		require.NoError(t, err)

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		respBytes, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, httpResp.StatusCode)
		if s.responseBody != nil {
			require.Equal(t, s.responseBody, respBytes)
		}

		entries := collector()
		require.Len(t, entries, 1)
		audit := entries[0].Audit.ReqResp
		require.NotNil(t, audit)

		// Assert Request Body Audit
		if s.requestBody != nil {
			require.NotNil(t, audit.Req.Body)
			require.Equal(t, len(s.requestBody), audit.Req.Body.Size)
			isLarge := len(s.requestBody) > logzaudit.LargeBodyThreashold
			require.Equal(t, isLarge, audit.Req.Body.Large)

			expectedHeadLen := min(len(s.requestBody), logzaudit.LargeBodyThreashold)
			require.Equal(t, s.requestBody[:expectedHeadLen], audit.Req.Body.Payload)

			if isLarge {
				nonHead := s.requestBody[logzaudit.LargeBodyThreashold:]
				expectedTailLen := min(len(nonHead), logzaudit.TailBodyThreashold)
				expectedTail := nonHead[len(nonHead)-expectedTailLen:]
				require.Equal(t, expectedTail, audit.Req.Body.Tail)
			} else {
				require.Nil(t, audit.Req.Body.Tail)
			}
		}

		// Assert Response Body Audit
		if s.responseBody != nil {
			require.NotNil(t, audit.Resp.Body)
			require.Equal(t, len(s.responseBody), audit.Resp.Body.Size)
			isLarge := len(s.responseBody) > logzaudit.LargeBodyThreashold
			require.Equal(t, isLarge, audit.Resp.Body.Large)

			expectedHeadLen := min(len(s.responseBody), logzaudit.LargeBodyThreashold)
			require.Equal(t, s.responseBody[:expectedHeadLen], audit.Resp.Body.Payload)

			if isLarge {
				nonHead := s.responseBody[logzaudit.LargeBodyThreashold:]
				expectedTailLen := min(len(nonHead), logzaudit.TailBodyThreashold)
				expectedTail := nonHead[len(nonHead)-expectedTailLen:]
				require.Equal(t, expectedTail, audit.Resp.Body.Tail)
			} else {
				require.Nil(t, audit.Resp.Body.Tail)
			}
		}
	}

	t.Run("Request body exceeding LargeBodyThreashold (70KB) read by handler", func(t *testing.T) {
		reqData := make([]byte, 70*1024)
		for i := range reqData {
			reqData[i] = byte(i % 256)
		}
		check(t, testScenario{
			requestBody:        reqData,
			responseBody:       []byte("ok"),
			handlerReadReqBody: true,
		})
	})

	t.Run("Request body exceeding LargeBodyThreashold (70KB) NOT read by handler", func(t *testing.T) {
		reqData := make([]byte, 70*1024)
		for i := range reqData {
			reqData[i] = byte((i + 42) % 256)
		}
		check(t, testScenario{
			requestBody:        reqData,
			responseBody:       []byte("ok"),
			handlerReadReqBody: false,
		})
	})

	t.Run("Response body exceeding LargeBodyThreashold (80KB)", func(t *testing.T) {
		respData := make([]byte, 80*1024)
		for i := range respData {
			respData[i] = byte((i + 7) % 256)
		}
		check(t, testScenario{
			requestBody:        []byte("small-req"),
			responseBody:       respData,
			handlerReadReqBody: true,
		})
	})

	t.Run("Body exceeding LargeBodyThreashold with non-head portion smaller than TailBodyThreashold", func(t *testing.T) {
		// 64KB + 200 bytes
		data := make([]byte, logzaudit.LargeBodyThreashold+200)
		for i := range data {
			data[i] = byte(i % 256)
		}
		check(t, testScenario{
			requestBody:        data,
			responseBody:       data,
			handlerReadReqBody: true,
		})
	})

	t.Run("Body exactly equal to LargeBodyThreashold (64KB)", func(t *testing.T) {
		data := make([]byte, logzaudit.LargeBodyThreashold)
		for i := range data {
			data[i] = byte(i % 256)
		}
		check(t, testScenario{
			requestBody:        data,
			responseBody:       data,
			handlerReadReqBody: true,
		})
	})

	t.Run("Streaming response written in small chunks across threshold", func(t *testing.T) {
		respData := make([]byte, 75*1024)
		for i := range respData {
			respData[i] = byte(i % 256)
		}
		check(t, testScenario{
			requestBody:        []byte("ping"),
			responseBody:       respData,
			handlerReadReqBody: true,
			streamResponse:     true,
		})
	})
}

func TestUnitAuditFilterMemoryBound(t *testing.T) {
	// Verifies that a multi-megabyte body streams cleanly through the filter
	// and accurately records size, 64KB payload, and 1KB tail without buffering
	// the multi-megabyte payload into memory.
	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()
	srv.Listen()

	collector := httpzserverfilteraudit.Collector()
	collector()

	srv.AddFilter(httpzserverfilteraudit.AuditFilter)

	const totalSize = 2 * 1024 * 1024 // 2MB

	srv.AddHandler("POST", "/stream", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		n, err := io.Copy(io.Discard, req.Body)
		require.NoError(t, err)
		require.Equal(t, int64(totalSize), n)

		w := resp(http.StatusOK, nil)
		_, err = w.Write([]byte("stream-received"))
		require.NoError(t, err)
	})

	srv.Start()

	// Create a repeating pattern reader that generates totalSize bytes
	pattern := []byte("0123456789abcdef")
	fullData := bytes.Repeat(pattern, totalSize/len(pattern))
	require.Equal(t, totalSize, len(fullData))

	req, err := http.NewRequest(http.MethodPost, srv.Base().MustJoinPathString("/stream").String(), bytes.NewReader(fullData))
	require.NoError(t, err)

	httpResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("stream-received"), respBytes)

	entries := collector()
	require.Len(t, entries, 1)
	reqAudit := entries[0].Audit.ReqResp.Req
	require.NotNil(t, reqAudit)
	require.NotNil(t, reqAudit.Body)
	require.Equal(t, totalSize, reqAudit.Body.Size)
	require.True(t, reqAudit.Body.Large)
	require.Equal(t, logzaudit.LargeBodyThreashold, len(reqAudit.Body.Payload))
	require.Equal(t, logzaudit.TailBodyThreashold, len(reqAudit.Body.Tail))
	require.Equal(t, fullData[:logzaudit.LargeBodyThreashold], reqAudit.Body.Payload)
	require.Equal(t, fullData[totalSize-logzaudit.TailBodyThreashold:], reqAudit.Body.Tail)
}
