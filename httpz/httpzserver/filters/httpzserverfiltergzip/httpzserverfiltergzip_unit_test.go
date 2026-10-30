package httpzserverfiltergzip_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilteraudit"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltergzip"
	"github.com/stretchr/testify/require"
)

func compressString(s string, level int) []byte {
	var b bytes.Buffer
	w, err := gzip.NewWriterLevel(&b, level)
	if err != nil {
		panic(err)
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	w.Close()
	return b.Bytes()
}

func decompressBytes(b []byte) (string, error) {
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func TestUnitFilterGzip(t *testing.T) {
	type testScenario struct {
		name                   string
		reqMethod              string
		reqBody                []byte
		reqHeaders             map[string]string
		handlerResponseStatus  int
		handlerResponseBody    string
		handlerResponseHeaders map[string]string

		expectedHandlerReqBody    string
		expectedHandlerReqHeaders map[string]string

		expectedRespStatus  int
		expectedRespHeaders map[string]string
		expectedRespIsGzip  bool
		expectedRespBodyStr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfiltergzip.GzipFilter)

		var actualHandlerReqBody string
		var actualHandlerReqHeaders http.Header

		srv.AddHandler(s.reqMethod, "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			actualHandlerReqHeaders = req.Headers.Clone()
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				actualHandlerReqBody = string(b)
			}

			h := make(http.Header)
			for k, v := range s.handlerResponseHeaders {
				h.Set(k, v)
			}
			w := resp(s.handlerResponseStatus, h)
			if s.handlerResponseBody != "" {
				_, err := w.Write([]byte(s.handlerResponseBody))
				require.NoError(t, err)
			}
		})

		srv.Start()

		var reqBodyReader io.Reader
		if s.reqBody != nil {
			reqBodyReader = bytes.NewReader(s.reqBody)
		}

		req, err := http.NewRequest(s.reqMethod, srv.Base().MustJoinPathString("/test").String(), reqBodyReader)
		require.NoError(t, err)
		for k, v := range s.reqHeaders {
			req.Header.Set(k, v)
		}

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedRespStatus, httpResp.StatusCode)

		for k, expectedVal := range s.expectedRespHeaders {
			require.Equal(t, expectedVal, httpResp.Header.Get(k), "Mismatch for response header: "+k)
		}

		respBodyBytes, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)

		if s.expectedRespStatus == 400 && string(respBodyBytes) == "invalid gzip body" {
			// Early exit for invalid gzip body case
			return
		}

		if s.expectedHandlerReqBody != "" {
			require.Equal(t, s.expectedHandlerReqBody, actualHandlerReqBody)
		}

		for k, expectedVal := range s.expectedHandlerReqHeaders {
			require.Equal(t, expectedVal, actualHandlerReqHeaders.Get(k), "Mismatch for handler req header: "+k)
			if expectedVal == "" {
				require.Empty(t, actualHandlerReqHeaders.Get(k))
			}
		}

		if s.expectedRespIsGzip {
			require.Equal(t, "gzip", httpResp.Header.Get("Content-Encoding"))
			decompressed, err := decompressBytes(respBodyBytes)
			require.NoError(t, err)
			require.Equal(t, s.expectedRespBodyStr, decompressed)
		} else {
			require.NotEqual(t, "gzip", httpResp.Header.Get("Content-Encoding"))
			require.Equal(t, s.expectedRespBodyStr, string(respBodyBytes))
		}
	}

	t.Run("No compression requested, no compression applied", func(t *testing.T) {
		check(t, testScenario{
			name:                  "No compression requested, no compression applied",
			reqMethod:             "GET",
			reqHeaders:            nil,
			handlerResponseStatus: 200,
			handlerResponseBody:   "hello world",
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderAcceptsGzip: "true", // http.DefaultClient injects Accept-Encoding: gzip
			},
			expectedRespStatus:  200,
			expectedRespIsGzip:  false,
			expectedRespBodyStr: "hello world",
		})
	})

	t.Run("Client accepts gzip, server compresses response", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Client accepts gzip, server compresses response",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{"Accept-Encoding": "gzip, deflate, br"},
			handlerResponseStatus: 200,
			handlerResponseBody:   "hello compressible world",
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderAcceptsGzip: "true",
			},
			expectedRespStatus:  200,
			expectedRespIsGzip:  true,
			expectedRespBodyStr: "hello compressible world",
			expectedRespHeaders: map[string]string{
				"Vary": "Accept-Encoding",
			},
		})
	})

	t.Run("Client accepts gzip with wildcard, server compresses response", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Client accepts gzip with wildcard, server compresses response",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{"Accept-Encoding": "*"},
			handlerResponseStatus: 200,
			handlerResponseBody:   "hello compressible world",
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderAcceptsGzip: "true",
			},
			expectedRespStatus:  200,
			expectedRespIsGzip:  true,
			expectedRespBodyStr: "hello compressible world",
		})
	})

	t.Run("Client accepts gzip but with q=0, server does NOT compress", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Client accepts gzip but with q=0, server does NOT compress",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{"Accept-Encoding": "gzip;q=0"},
			handlerResponseStatus: 200,
			handlerResponseBody:   "hello world",
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderAcceptsGzip: "false",
			},
			expectedRespStatus:  200,
			expectedRespIsGzip:  false,
			expectedRespBodyStr: "hello world",
		})
	})

	t.Run("Status 204 No Content is not compressed", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Status 204 No Content is not compressed",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{"Accept-Encoding": "gzip"},
			handlerResponseStatus: 204,
			handlerResponseBody:   "",
			expectedRespStatus:    204,
			expectedRespIsGzip:    false,
			expectedRespBodyStr:   "",
		})
	})

	t.Run("Status 304 Not Modified is not compressed", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Status 304 Not Modified is not compressed",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{"Accept-Encoding": "gzip"},
			handlerResponseStatus: 304,
			handlerResponseBody:   "",
			expectedRespStatus:    304,
			expectedRespIsGzip:    false,
			expectedRespBodyStr:   "",
		})
	})

	t.Run("Response with Content-Encoding already set is not compressed again", func(t *testing.T) {
		check(t, testScenario{
			name:                   "Response with Content-Encoding already set is not compressed again",
			reqMethod:              "GET",
			reqHeaders:             map[string]string{"Accept-Encoding": "gzip"},
			handlerResponseStatus:  200,
			handlerResponseBody:    "already custom encoded",
			handlerResponseHeaders: map[string]string{"Content-Encoding": "custom"},
			expectedRespStatus:     200,
			expectedRespIsGzip:     false,
			expectedRespBodyStr:    "already custom encoded",
			expectedRespHeaders: map[string]string{
				"Content-Encoding": "custom",
			},
		})
	})

	t.Run("Request with gzip body is decompressed", func(t *testing.T) {
		reqBodyStr := "this is a gzipped request body"
		gzippedBody := compressString(reqBodyStr, gzip.DefaultCompression)

		check(t, testScenario{
			name:                   "Request with gzip body is decompressed",
			reqMethod:              "POST",
			reqBody:                gzippedBody,
			reqHeaders:             map[string]string{"Content-Encoding": "gzip"},
			handlerResponseStatus:  200,
			handlerResponseBody:    "ok",
			expectedHandlerReqBody: reqBodyStr,
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderContentEncodingWas: "gzip",
				"Content-Encoding":                      "",
				httpzserverfiltergzip.HeaderAcceptsGzip: "true", // injected by http.DefaultClient
			},
			expectedRespStatus:  200,
			expectedRespIsGzip:  false,
			expectedRespBodyStr: "ok",
		})
	})

	t.Run("Invalid gzip request body returns 400 Bad Request", func(t *testing.T) {
		check(t, testScenario{
			name:               "Invalid gzip request body returns 400 Bad Request",
			reqMethod:          "POST",
			reqBody:            []byte("not a valid gzip body"),
			reqHeaders:         map[string]string{"Content-Encoding": "gzip"},
			expectedRespStatus: 400,
		})
	})

	t.Run("Malicious incoming X-I6-Content-Encoding-Was header is removed", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Malicious incoming X-I6-Content-Encoding-Was header is removed",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{httpzserverfiltergzip.HeaderContentEncodingWas: "fake"},
			handlerResponseStatus: 200,
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderContentEncodingWas: "",
				httpzserverfiltergzip.HeaderAcceptsGzip:        "true", // injected by http.DefaultClient
			},
			expectedRespStatus: 200,
		})
	})

	t.Run("Malicious incoming X-I6-Accepts-Gzip header is removed", func(t *testing.T) {
		check(t, testScenario{
			name:                  "Malicious incoming X-I6-Accepts-Gzip header is removed",
			reqMethod:             "GET",
			reqHeaders:            map[string]string{httpzserverfiltergzip.HeaderAcceptsGzip: "true", "Accept-Encoding": "gzip;q=0"},
			handlerResponseStatus: 200,
			expectedHandlerReqHeaders: map[string]string{
				httpzserverfiltergzip.HeaderAcceptsGzip: "false",
			},
			expectedRespStatus: 200,
		})
	})
}

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

func TestUnitGzipFilterStandalone(t *testing.T) {
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

		srv.AddFilter(httpzserverfilteraudit.Filter)
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
