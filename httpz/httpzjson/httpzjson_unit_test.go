package httpzjson_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzjson"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

type echoData struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func TestUnitJsonDo(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	server.AddHandler("POST", "/echo", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		w := resp(http.StatusOK, http.Header{"Content-Type": []string{"application/json"}})
		_, err = w.Write(body)
		require.NoError(t, err)
	})

	server.AddHandler("GET", "/invalid-json", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("{invalid-json-body}"))
		require.NoError(t, err)
	})

	server.AddHandler("GET", "/error", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusInternalServerError, http.Header{})
		_, err := w.Write([]byte("server error"))
		require.NoError(t, err)
	})

	server.Listen()
	server.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: server.Base()})

	type testScenario struct {
		name           string
		method         string
		path           string
		inputFn        func() any
		outputTarget   func() *echoData
		expectedErrMsg string
		expectedCount  int
		expectedMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var outPtr *echoData
		var outputFn func(status int, headers http.Header) any
		if s.outputTarget != nil {
			outPtr = s.outputTarget()
			if outPtr != nil {
				outputFn = func(status int, headers http.Header) any {
					return outPtr
				}
			}
		}

		opts := httpzjson.Options{
			Client: client,
			Req:    httpzrequest.New(s.method, s.path),
			Input:  s.inputFn,
			Output: outputFn,
		}

		err := httpzjson.Do(ctx, opts)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			return
		}

		require.NoError(t, err)
		if outPtr != nil {
			require.Equal(t, s.expectedMsg, outPtr.Message)
			require.Equal(t, s.expectedCount, outPtr.Count)
		}
	}

	t.Run("Input and Output serialization roundtrip succeeds", func(t *testing.T) {
		out := &echoData{}
		check(t, testScenario{
			name:   "Input and Output serialization roundtrip succeeds",
			method: "POST",
			path:   "/echo",
			inputFn: func() any {
				return echoData{Message: "hello", Count: 42}
			},
			outputTarget: func() *echoData {
				return out
			},
			expectedMsg:   "hello",
			expectedCount: 42,
		})
	})

	t.Run("Nil Input and nil Output succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:         "Nil Input and nil Output succeeds",
			method:       "POST",
			path:         "/echo",
			inputFn:      nil,
			outputTarget: nil,
		})
	})

	t.Run("Output returning nil skips parsing", func(t *testing.T) {
		check(t, testScenario{
			name:   "Output returning nil skips parsing",
			method: "POST",
			path:   "/echo",
			inputFn: func() any {
				return echoData{Message: "skip"}
			},
			outputTarget: func() *echoData {
				return nil
			},
		})
	})

	t.Run("Invalid JSON response returns parse error", func(t *testing.T) {
		out := &echoData{}
		check(t, testScenario{
			name:   "Invalid JSON response returns parse error",
			method: "GET",
			path:   "/invalid-json",
			outputTarget: func() *echoData {
				return out
			},
			expectedErrMsg: "failed to parse json response",
		})
	})

	t.Run("Server error status code returns validation error", func(t *testing.T) {
		check(t, testScenario{
			name:           "Server error status code returns validation error",
			method:         "GET",
			path:           "/error",
			expectedErrMsg: "http validating error: 500",
		})
	})
}

func TestUnitJsonMustDo(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	server.AddHandler("GET", "/ok", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte(`{"message":"ok"}`))
		require.NoError(t, err)
	})

	server.AddHandler("GET", "/fail", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusBadRequest, http.Header{})
		_, err := w.Write([]byte(`{"message":"bad request"}`))
		require.NoError(t, err)
	})

	server.Listen()
	server.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: server.Base()})

	type testScenario struct {
		name        string
		path        string
		expectPanic bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		opts := httpzjson.Options{
			Client: client,
			Req:    httpzrequest.New("GET", s.path),
		}

		if s.expectPanic {
			require.Panics(t, func() {
				httpzjson.MustDo(ctx, opts)
			})
			return
		}

		require.NotPanics(t, func() {
			httpzjson.MustDo(ctx, opts)
		})
	}

	t.Run("MustDo succeeds without panic on valid response", func(t *testing.T) {
		check(t, testScenario{
			name:        "MustDo succeeds without panic on valid response",
			path:        "/ok",
			expectPanic: false,
		})
	})

	t.Run("MustDo panics on validation error", func(t *testing.T) {
		check(t, testScenario{
			name:        "MustDo panics on validation error",
			path:        "/fail",
			expectPanic: true,
		})
	})
}
