package httpzserverfiltergzip_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltergzip"
	"github.com/stretchr/testify/require"
)

func gzipCompress(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	require.NoError(t, err)
	err = gw.Close()
	require.NoError(t, err)
	return buf.Bytes()
}

func gzipDecompress(t *testing.T, data []byte) []byte {
	t.Helper()
	gr, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer gr.Close()
	decompressed, err := io.ReadAll(gr)
	require.NoError(t, err)
	return decompressed
}

func TestManualGzipFilterStandalone(t *testing.T) {
	type testScenario struct {
		name                   string
		requestMethod          string
		requestPath            string
		requestHeaders         map[string]string
		requestBody            []byte
		handlerStatus          int
		handlerHeaders         map[string]string
		handlerBody            []byte
		expectedStatus         int
		expectedRespEncoding   string
		expectedRespVary       string
		expectedDecompressed   []byte
		handlerReceivedBody    []byte
		verifyDecompressedBody bool
		expectedWasHeader      string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfiltergzip.GzipFilter)

		var receivedReqBody []byte
		var receivedWasHeader string
		srv.AddHandler(s.requestMethod, s.requestPath, func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			receivedWasHeader = req.Headers.Get(httpzserverfiltergzip.HeaderContentEncodingWas)
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				receivedReqBody = b
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
		req, err := http.NewRequest(s.requestMethod, srv.Base().MustJoinPathString(s.requestPath).String(), bodyReader)
		require.NoError(t, err)
		for k, v := range s.requestHeaders {
			req.Header.Set(k, v)
		}

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedStatus, httpResp.StatusCode)
		require.Equal(t, s.expectedRespEncoding, httpResp.Header.Get("Content-Encoding"))
		if s.expectedRespVary != "" {
			require.Contains(t, httpResp.Header.Get("Vary"), s.expectedRespVary)
		}

		rawRespBody, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)

		if s.verifyDecompressedBody {
			decompressed := gzipDecompress(t, rawRespBody)
			require.Equal(t, s.expectedDecompressed, decompressed)
		} else if s.expectedDecompressed != nil {
			require.Equal(t, s.expectedDecompressed, rawRespBody)
		}

		if s.handlerReceivedBody != nil {
			require.Equal(t, s.handlerReceivedBody, receivedReqBody)
		}

		if s.expectedWasHeader != "" {
			require.Equal(t, s.expectedWasHeader, receivedWasHeader)
		} else {
			require.Empty(t, receivedWasHeader)
		}
	}

	t.Run("Gzip request body is decompressed for handler and sets X-I6-Content-Encoding-Was", func(t *testing.T) {
		plainPayload := []byte("decompressed request payload content")
		compressedPayload := gzipCompress(t, plainPayload)

		check(t, testScenario{
			requestMethod: "POST",
			requestPath:   "/upload",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:          compressedPayload,
			handlerStatus:        http.StatusOK,
			handlerBody:          []byte("processed"),
			expectedStatus:       http.StatusOK,
			expectedDecompressed: []byte("processed"),
			handlerReceivedBody:  plainPayload,
			expectedWasHeader:    "gzip",
		})
	})

	t.Run("Spoofed X-I6-Content-Encoding-Was header is removed when request is not gzipped", func(t *testing.T) {
		check(t, testScenario{
			requestMethod: "POST",
			requestPath:   "/spoofed",
			requestHeaders: map[string]string{
				httpzserverfiltergzip.HeaderContentEncodingWas: "gzip",
			},
			requestBody:          []byte("plain body"),
			handlerStatus:        http.StatusOK,
			handlerBody:          []byte("ok"),
			expectedStatus:       http.StatusOK,
			expectedDecompressed: []byte("ok"),
			handlerReceivedBody:  []byte("plain body"),
			expectedWasHeader:    "",
		})
	})

	t.Run("Spoofed X-I6-Content-Encoding-Was header is overwritten with actual encoding on gzip request", func(t *testing.T) {
		plainPayload := []byte("decompressed content")
		check(t, testScenario{
			requestMethod: "POST",
			requestPath:   "/spoofed-replace",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
				httpzserverfiltergzip.HeaderContentEncodingWas: "fake-value",
			},
			requestBody:          gzipCompress(t, plainPayload),
			handlerStatus:        http.StatusOK,
			handlerBody:          []byte("ok"),
			expectedStatus:       http.StatusOK,
			expectedDecompressed: []byte("ok"),
			handlerReceivedBody:  plainPayload,
			expectedWasHeader:    "gzip",
		})
	})

	t.Run("Invalid gzip request body returns 400 Bad Request", func(t *testing.T) {
		check(t, testScenario{
			requestMethod: "POST",
			requestPath:   "/invalid-gzip",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:          []byte("not a valid gzip header data"),
			handlerStatus:        http.StatusOK,
			handlerBody:          []byte("should not be called"),
			expectedStatus:       http.StatusBadRequest,
			expectedDecompressed: []byte("invalid gzip body"),
		})
	})

	t.Run("Response is compressed when Accept-Encoding contains gzip", func(t *testing.T) {
		plainResp := []byte("uncompressed response body to be compressed by filter")

		check(t, testScenario{
			requestMethod: "GET",
			requestPath:   "/compress-me",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip, deflate, br",
			},
			handlerStatus:          http.StatusOK,
			handlerBody:            plainResp,
			expectedStatus:         http.StatusOK,
			expectedRespEncoding:   "gzip",
			expectedRespVary:       "Accept-Encoding",
			expectedDecompressed:   plainResp,
			verifyDecompressedBody: true,
		})
	})

	t.Run("Response is not compressed when Accept-Encoding does not contain gzip", func(t *testing.T) {
		plainResp := []byte("plain text response")

		check(t, testScenario{
			requestMethod: "GET",
			requestPath:   "/no-gzip",
			requestHeaders: map[string]string{
				"Accept-Encoding": "deflate, br",
			},
			handlerStatus:        http.StatusOK,
			handlerBody:          plainResp,
			expectedStatus:       http.StatusOK,
			expectedRespEncoding: "",
			expectedDecompressed: plainResp,
		})
	})

	t.Run("Response is not compressed when Accept-Encoding has gzip with q=0", func(t *testing.T) {
		plainResp := []byte("plain text response without gzip")

		check(t, testScenario{
			requestMethod: "GET",
			requestPath:   "/gzip-q0",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip;q=0, deflate",
			},
			handlerStatus:        http.StatusOK,
			handlerBody:          plainResp,
			expectedStatus:       http.StatusOK,
			expectedRespEncoding: "",
			expectedDecompressed: plainResp,
		})
	})

	t.Run("204 No Content response is not compressed", func(t *testing.T) {
		check(t, testScenario{
			requestMethod: "DELETE",
			requestPath:   "/item",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			handlerStatus:        http.StatusNoContent,
			expectedStatus:       http.StatusNoContent,
			expectedRespEncoding: "",
			expectedDecompressed: []byte{},
		})
	})

	t.Run("Response with existing Content-Encoding is not re-compressed", func(t *testing.T) {
		customBody := []byte("custom-encoded-bytes")

		check(t, testScenario{
			requestMethod: "GET",
			requestPath:   "/custom-enc",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			handlerStatus: http.StatusOK,
			handlerHeaders: map[string]string{
				"Content-Encoding": "custom",
			},
			handlerBody:          customBody,
			expectedStatus:       http.StatusOK,
			expectedRespEncoding: "custom",
			expectedDecompressed: customBody,
		})
	})
}

func TestUnitGzipFilterWithAuditFilter(t *testing.T) {
	// Tests pipeline: GzipFilter runs before AuditFilter
	type testScenario struct {
		name                string
		plainReqBody        []byte
		plainRespBody       []byte
		requestCompressed   bool
		acceptsGzip         bool
		expectedAuditLargeR bool
		expectedAuditLargeW bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfilteraudit.Collector()
		collector() // clear prior entries

		// Register GzipFilter BEFORE AuditFilter
		srv.AddFilter(httpzserverfiltergzip.GzipFilter)
		srv.AddFilter(httpzserverfilteraudit.AuditFilter)

		var receivedHandlerBody []byte
		srv.AddHandler("POST", "/audit-gzip", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				receivedHandlerBody = b
			}
			w := resp(http.StatusOK, nil)
			if s.plainRespBody != nil {
				_, err := w.Write(s.plainRespBody)
				require.NoError(t, err)
			}
		})

		srv.Start()

		var bodyReader io.Reader
		reqHeaders := make(http.Header)
		if s.requestCompressed && s.plainReqBody != nil {
			compressed := gzipCompress(t, s.plainReqBody)
			bodyReader = bytes.NewReader(compressed)
			reqHeaders.Set("Content-Encoding", "gzip")
		} else if s.plainReqBody != nil {
			bodyReader = bytes.NewReader(s.plainReqBody)
		}

		if s.acceptsGzip {
			reqHeaders.Set("Accept-Encoding", "gzip")
		}

		req, err := http.NewRequest(http.MethodPost, srv.Base().MustJoinPathString("/audit-gzip").String(), bodyReader)
		require.NoError(t, err)
		req.Header = reqHeaders

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, http.StatusOK, httpResp.StatusCode)

		// Verify client received and decompressed response
		rawRespBody, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)

		if s.acceptsGzip && len(s.plainRespBody) > 0 {
			require.Equal(t, "gzip", httpResp.Header.Get("Content-Encoding"))
			decompressed := gzipDecompress(t, rawRespBody)
			require.Equal(t, s.plainRespBody, decompressed)
		} else {
			require.Equal(t, s.plainRespBody, rawRespBody)
		}

		// Verify handler received uncompressed body
		if s.plainReqBody != nil {
			require.Equal(t, s.plainReqBody, receivedHandlerBody)
		}

		// Verify AuditFilter captured uncompressed request and response payloads!
		entries := collector()
		require.Len(t, entries, 1)
		audit := entries[0].Audit.ReqResp
		require.NotNil(t, audit)

		// Request audit asserts
		if s.plainReqBody != nil {
			require.NotNil(t, audit.Req.Body)
			require.Equal(t, len(s.plainReqBody), audit.Req.Body.Size)
			require.Equal(t, s.expectedAuditLargeR, audit.Req.Body.Large)

			expectedHeadLen := min(len(s.plainReqBody), logzaudit.LargeBodyThreashold)
			require.Equal(t, s.plainReqBody[:expectedHeadLen], audit.Req.Body.Payload)

			if s.expectedAuditLargeR {
				nonHead := s.plainReqBody[logzaudit.LargeBodyThreashold:]
				expectedTailLen := min(len(nonHead), logzaudit.TailBodyThreashold)
				expectedTail := nonHead[len(nonHead)-expectedTailLen:]
				require.Equal(t, expectedTail, audit.Req.Body.Tail)
			}
		}

		// Verify HeaderContentEncodingWas in audit log
		if s.requestCompressed {
			require.Equal(t, "gzip", audit.Req.Headers.Get(httpzserverfiltergzip.HeaderContentEncodingWas))
		} else {
			require.Empty(t, audit.Req.Headers.Get(httpzserverfiltergzip.HeaderContentEncodingWas))
		}

		// Response audit asserts
		if s.plainRespBody != nil {
			require.NotNil(t, audit.Resp.Body)
			require.Equal(t, len(s.plainRespBody), audit.Resp.Body.Size)
			require.Equal(t, s.expectedAuditLargeW, audit.Resp.Body.Large)

			expectedHeadLen := min(len(s.plainRespBody), logzaudit.LargeBodyThreashold)
			require.Equal(t, s.plainRespBody[:expectedHeadLen], audit.Resp.Body.Payload)

			if s.expectedAuditLargeW {
				nonHead := s.plainRespBody[logzaudit.LargeBodyThreashold:]
				expectedTailLen := min(len(nonHead), logzaudit.TailBodyThreashold)
				expectedTail := nonHead[len(nonHead)-expectedTailLen:]
				require.Equal(t, expectedTail, audit.Resp.Body.Tail)
			}
		}
	}

	t.Run("Small gzipped request and response with AuditFilter", func(t *testing.T) {
		check(t, testScenario{
			plainReqBody:        []byte("hello uncompressed request for audit"),
			plainRespBody:       []byte("hello uncompressed response from handler"),
			requestCompressed:   true,
			acceptsGzip:         true,
			expectedAuditLargeR: false,
			expectedAuditLargeW: false,
		})
	})

	t.Run("Large gzipped request (>64KB) and large response (>64KB) with AuditFilter", func(t *testing.T) {
		largeReq := make([]byte, 70*1024)
		for i := range largeReq {
			largeReq[i] = byte((i * 3) % 256)
		}
		largeResp := make([]byte, 80*1024)
		for i := range largeResp {
			largeResp[i] = byte((i * 7) % 256)
		}

		check(t, testScenario{
			plainReqBody:        largeReq,
			plainRespBody:       largeResp,
			requestCompressed:   true,
			acceptsGzip:         true,
			expectedAuditLargeR: true,
			expectedAuditLargeW: true,
		})
	})
}
