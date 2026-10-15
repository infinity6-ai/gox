package httpzserverfiltergzip_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltergzip"
	"github.com/stretchr/testify/require"
)

func TestUnitFilterGzipWithAudit(t *testing.T) {
	type testScenario struct {
		name                   string
		auditService           *logzaudit.AuditService
		requestMethod          string
		requestPath            string
		requestHeaders         map[string]string
		requestBody            []byte
		handlerStatus          int
		handlerHeaders         map[string]string
		handlerBody            []byte
		verifyInHandler        func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry)
		expectedStatus         int
		expectedRespEncoding   string
		expectedRespBody       []byte
		verifyDecompressedResp bool

		// Audit verification expectations
		expectedReqId          string
		expectedServiceName    string
		expectedReqCompressed  bool
		expectedRespCompressed bool
		expectedReqBodySize    int
		expectedReqCompSize    int
		expectedReqPayload     []byte
		expectedReqLarge       bool
		expectedReqTail        []byte
		expectedRespBodySize   int
		expectedRespCompSize   int
		expectedRespPayload    []byte
		expectedRespLarge      bool
		expectedRespTail       []byte
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector() // clear prior log entries

		srv.AddFilter(httpzserverfilteraudit.Filter)
		srv.AddFilter(httpzserverfiltergzip.GzipFilter)

		srv.AddHandler(s.requestMethod, s.requestPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			entry := httpzserverfilteraudit.GetEntry(ctx)
			require.NotNil(t, entry)

			if req.Body != nil {
				_, err := io.ReadAll(req.Body)
				require.NoError(t, err)
			}

			if s.verifyInHandler != nil {
				s.verifyInHandler(t, req, entry)
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

		client := &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		}

		httpResp, err := client.Do(clientReq)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedStatus, httpResp.StatusCode)
		require.Equal(t, s.expectedRespEncoding, httpResp.Header.Get("Content-Encoding"))

		rawRespBody, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)

		if s.verifyDecompressedResp {
			decompressed := gzipDecompress(t, rawRespBody)
			require.Equal(t, s.expectedRespBody, decompressed)
		} else if s.expectedRespBody != nil {
			require.Equal(t, s.expectedRespBody, rawRespBody)
		}

		// Verify Audit Logs
		entries := collector()
		require.Len(t, entries, 1)

		audit := entries[0].Audit
		require.NotNil(t, audit)
		reqResp := audit.ReqResp
		require.NotNil(t, reqResp)

		if s.auditService != nil {
			require.NotNil(t, reqResp.Service)
			require.Equal(t, s.expectedServiceName, reqResp.Service.Name)
		}

		if s.expectedReqId != "" {
			require.Equal(t, s.expectedReqId, reqResp.Id)
		} else {
			require.NotEmpty(t, reqResp.Id)
		}

		// Verify Request Audit Body
		require.NotNil(t, reqResp.Req)
		require.Equal(t, s.requestMethod, reqResp.Req.Method)
		require.Equal(t, s.requestPath, reqResp.Req.Path)
		if reqResp.Req.Body != nil {
			require.Equal(t, s.expectedReqBodySize, reqResp.Req.Body.Size)
			require.Equal(t, s.expectedReqCompSize, reqResp.Req.Body.CompressedSize)
			require.Equal(t, s.expectedReqLarge, reqResp.Req.Body.Large)
			if s.expectedReqPayload != nil {
				require.Equal(t, s.expectedReqPayload, reqResp.Req.Body.Payload)
			}
			if s.expectedReqTail != nil {
				require.Equal(t, s.expectedReqTail, reqResp.Req.Body.Tail)
			}
		}

		// Verify Response Audit Body
		require.NotNil(t, reqResp.Resp)
		require.Equal(t, s.expectedStatus, reqResp.Resp.Status)
		if reqResp.Resp.Body != nil {
			require.Equal(t, s.expectedRespBodySize, reqResp.Resp.Body.Size)
			if s.expectedRespCompressed {
				require.Equal(t, len(rawRespBody), reqResp.Resp.Body.CompressedSize)
				require.Greater(t, reqResp.Resp.Body.CompressedSize, 0)
			} else {
				require.Equal(t, s.expectedRespCompSize, reqResp.Resp.Body.CompressedSize)
			}
			require.Equal(t, s.expectedRespLarge, reqResp.Resp.Body.Large)
			if s.expectedRespPayload != nil {
				require.Equal(t, s.expectedRespPayload, reqResp.Resp.Body.Payload)
			}
			if s.expectedRespTail != nil {
				require.Equal(t, s.expectedRespTail, reqResp.Resp.Body.Tail)
			}
		}
	}

	testService := &logzaudit.AuditService{
		Name:    "audit-gzip-service",
		Version: "1.0.0",
	}

	t.Run("Gzip request decompression records uncompressed payload and compressed size in audit", func(t *testing.T) {
		plainPayload := []byte("decompressed request payload content for audit test")
		compressedPayload := gzipCompress(t, plainPayload)

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			expectedReqId:       "req-audit-gzip-1",
			requestMethod:       http.MethodPost,
			requestPath:         "/upload",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
				httpz.HeaderReqId:  "req-audit-gzip-1",
			},
			requestBody: compressedPayload,
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.True(t, entry.ReqCompressed)
				require.Equal(t, "gzip", req.Headers.Get(httpzserverfiltergzip.HeaderContentEncodingWas))
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            []byte("ok"),
			expectedStatus:         http.StatusOK,
			expectedRespBody:       []byte("ok"),
			expectedReqCompressed:  true,
			expectedReqBodySize:    len(plainPayload),
			expectedReqCompSize:    len(compressedPayload),
			expectedReqPayload:     plainPayload,
			expectedRespCompressed: false,
			expectedRespBodySize:   len("ok"),
			expectedRespCompSize:   0,
			expectedRespPayload:    []byte("ok"),
		})
	})

	t.Run("Gzip response compression records uncompressed payload and compressed size in audit", func(t *testing.T) {
		plainResp := []byte("uncompressed response body that should be compressed by filter for client")

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodGet,
			requestPath:         "/download",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			handlerStatus: http.StatusOK,
			handlerBody:   plainResp,
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.False(t, entry.ReqCompressed)
			},
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "gzip",
			expectedRespBody:       plainResp,
			verifyDecompressedResp: true,
			expectedReqCompressed:  false,
			expectedRespCompressed: true,
			expectedRespBodySize:   len(plainResp),
			expectedRespPayload:    plainResp,
		})
	})

	t.Run("Both request and response compressed record audit metrics simultaneously", func(t *testing.T) {
		plainReq := []byte("incoming gzipped request data to process")
		compressedReq := gzipCompress(t, plainReq)
		plainResp := []byte("outgoing gzipped response data processed successfully")

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodPost,
			requestPath:         "/transform",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
				"Accept-Encoding":  "gzip",
			},
			requestBody: compressedReq,
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.True(t, entry.ReqCompressed)
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            plainResp,
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "gzip",
			expectedRespBody:       plainResp,
			verifyDecompressedResp: true,
			expectedReqCompressed:  true,
			expectedReqBodySize:    len(plainReq),
			expectedReqCompSize:    len(compressedReq),
			expectedReqPayload:     plainReq,
			expectedRespCompressed: true,
			expectedRespBodySize:   len(plainResp),
			expectedRespPayload:    plainResp,
		})
	})

	t.Run("Uncompressed request and response have zero CompressedSize in audit", func(t *testing.T) {
		plainReq := []byte("plain text request body")
		plainResp := []byte("plain text response body")

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodPost,
			requestPath:         "/plain",
			requestBody:         plainReq,
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.False(t, entry.ReqCompressed)
				require.False(t, entry.RespCompressed)
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            plainResp,
			expectedStatus:         http.StatusOK,
			expectedRespBody:       plainResp,
			expectedReqCompressed:  false,
			expectedReqBodySize:    len(plainReq),
			expectedReqCompSize:    0,
			expectedReqPayload:     plainReq,
			expectedRespCompressed: false,
			expectedRespBodySize:   len(plainResp),
			expectedRespCompSize:   0,
			expectedRespPayload:    plainResp,
		})
	})

	t.Run("Status 204 No Content response is not compressed and audit reflects zero compressed size", func(t *testing.T) {
		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodDelete,
			requestPath:         "/item/123",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.False(t, entry.RespCompressed)
			},
			handlerStatus:          http.StatusNoContent,
			expectedStatus:         http.StatusNoContent,
			expectedRespEncoding:   "",
			expectedRespBody:       []byte{},
			expectedRespCompressed: false,
			expectedRespBodySize:   0,
			expectedRespCompSize:   0,
		})
	})

	t.Run("Status 304 Not Modified response is not compressed and audit reflects zero compressed size", func(t *testing.T) {
		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodGet,
			requestPath:         "/cached",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			verifyInHandler: func(t *testing.T, req *httpzrequest.Req, entry *httpzserverfilteraudit.Entry) {
				require.False(t, entry.RespCompressed)
			},
			handlerStatus:          http.StatusNotModified,
			expectedStatus:         http.StatusNotModified,
			expectedRespEncoding:   "",
			expectedRespBody:       []byte{},
			expectedRespCompressed: false,
			expectedRespBodySize:   0,
			expectedRespCompSize:   0,
		})
	})

	t.Run("Response with existing Content-Encoding is not re-compressed and audit reflects not compressed", func(t *testing.T) {
		customBody := []byte("already-custom-encoded-data")

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodGet,
			requestPath:         "/custom-encoded",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			handlerHeaders: map[string]string{
				"Content-Encoding": "custom",
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            customBody,
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "custom",
			expectedRespBody:       customBody,
			expectedRespCompressed: false,
			expectedRespBodySize:   len(customBody),
			expectedRespCompSize:   0,
			expectedRespPayload:    customBody,
		})
	})

	t.Run("Client with Accept-Encoding gzip with q=0 is not compressed and audit reflects uncompressed", func(t *testing.T) {
		plainResp := []byte("plain response for q=0 client")

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodGet,
			requestPath:         "/q0",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip;q=0, deflate",
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            plainResp,
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "",
			expectedRespBody:       plainResp,
			expectedRespCompressed: false,
			expectedRespBodySize:   len(plainResp),
			expectedRespCompSize:   0,
			expectedRespPayload:    plainResp,
		})
	})

	t.Run("Large compressed request exceeding threshold flags Large and captures uncompressed head and tail", func(t *testing.T) {
		chunk := "Large-Request-Body-Chunk-0123456789-" // 36 bytes
		repeatCount := 2000                             // 72,000 bytes > 64KB (65,536)
		plainPayload := []byte(strings.Repeat(chunk, repeatCount))
		compressedPayload := gzipCompress(t, plainPayload)

		expectedHead := plainPayload[:logzaudit.LargeBodyThreashold]
		expectedTail := plainPayload[len(plainPayload)-logzaudit.TailBodyThreashold:]

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodPost,
			requestPath:         "/large-upload",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:            compressedPayload,
			handlerStatus:          http.StatusOK,
			handlerBody:            []byte("received-large"),
			expectedStatus:         http.StatusOK,
			expectedRespBody:       []byte("received-large"),
			expectedReqCompressed:  true,
			expectedReqBodySize:    len(plainPayload),
			expectedReqCompSize:    len(compressedPayload),
			expectedReqPayload:     expectedHead,
			expectedReqLarge:       true,
			expectedReqTail:        expectedTail,
			expectedRespCompressed: false,
			expectedRespBodySize:   len("received-large"),
			expectedRespCompSize:   0,
			expectedRespPayload:    []byte("received-large"),
		})
	})

	t.Run("Large compressed response exceeding threshold flags Large and captures uncompressed head and tail", func(t *testing.T) {
		chunk := "Large-Response-Body-Chunk-9876543210-" // 37 bytes
		repeatCount := 2000                              // 74,000 bytes > 64KB (65,536)
		plainResp := []byte(strings.Repeat(chunk, repeatCount))

		expectedHead := plainResp[:logzaudit.LargeBodyThreashold]
		expectedTail := plainResp[len(plainResp)-logzaudit.TailBodyThreashold:]

		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodGet,
			requestPath:         "/large-download",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            plainResp,
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "gzip",
			expectedRespBody:       plainResp,
			verifyDecompressedResp: true,
			expectedReqCompressed:  false,
			expectedRespCompressed: true,
			expectedRespBodySize:   len(plainResp),
			expectedRespLarge:      true,
			expectedRespPayload:    expectedHead,
			expectedRespTail:       expectedTail,
		})
	})

	t.Run("Invalid gzip request body returns 400 Bad Request and audit captures error response", func(t *testing.T) {
		check(t, testScenario{
			auditService:        testService,
			expectedServiceName: "audit-gzip-service",
			requestMethod:       http.MethodPost,
			requestPath:         "/invalid-upload",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:            []byte("not a valid gzip header data"),
			handlerStatus:          http.StatusOK,
			handlerBody:            []byte("should not execute"),
			expectedStatus:         http.StatusBadRequest,
			expectedRespBody:       []byte("invalid gzip body"),
			expectedReqCompressed:  false,
			expectedReqBodySize:    len("not a valid gzip header data"),
			expectedReqCompSize:    0,
			expectedReqPayload:     []byte("not a valid gzip header data"),
			expectedRespCompressed: false,
			expectedRespBodySize:   len("invalid gzip body"),
			expectedRespCompSize:   0,
			expectedRespPayload:    []byte("invalid gzip body"),
		})
	})
}
