package httpz_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

type trackingReader struct {
	io.Reader
	closed bool
}

func (t *trackingReader) Close() error {
	t.closed = true
	return nil
}

func TestUnitValidateRespSuccess(t *testing.T) {
	type testScenario struct {
		name       string
		statusCode int
		expectErr  bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		resp := &httpzclient.Resp{StatusCode: s.statusCode}
		err := httpz.ValidateRespSuccess(resp)
		if s.expectErr {
			require.Error(t, err)
			require.ErrorIs(t, err, httpz.ErrStatus)
			return
		}
		require.NoError(t, err)
	}

	t.Run("Status 200 OK succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 200 OK succeeds",
			statusCode: http.StatusOK,
			expectErr:  false,
		})
	})

	t.Run("Status 201 Created succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 201 Created succeeds",
			statusCode: http.StatusCreated,
			expectErr:  false,
		})
	})

	t.Run("Status 204 No Content succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 204 No Content succeeds",
			statusCode: http.StatusNoContent,
			expectErr:  false,
		})
	})

	t.Run("Status 299 succeeds", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 299 succeeds",
			statusCode: 299,
			expectErr:  false,
		})
	})

	t.Run("Status 100 Continue returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 100 Continue returns error",
			statusCode: http.StatusContinue,
			expectErr:  true,
		})
	})

	t.Run("Status 199 returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 199 returns error",
			statusCode: 199,
			expectErr:  true,
		})
	})

	t.Run("Status 301 Moved Permanently returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 301 Moved Permanently returns error",
			statusCode: http.StatusMovedPermanently,
			expectErr:  true,
		})
	})

	t.Run("Status 400 Bad Request returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 400 Bad Request returns error",
			statusCode: http.StatusBadRequest,
			expectErr:  true,
		})
	})

	t.Run("Status 404 Not Found returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 404 Not Found returns error",
			statusCode: http.StatusNotFound,
			expectErr:  true,
		})
	})

	t.Run("Status 500 Internal Server Error returns error", func(t *testing.T) {
		check(t, testScenario{
			name:       "Status 500 Internal Server Error returns error",
			statusCode: http.StatusInternalServerError,
			expectErr:  true,
		})
	})
}

func TestUnitDoValidation(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	server.AddHandler("GET", "/ok", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("ok response"))
		require.NoError(t, err)
	})

	server.AddHandler("GET", "/bad-request", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusBadRequest, http.Header{})
		_, err := w.Write([]byte("invalid parameter"))
		require.NoError(t, err)
	})

	server.Listen()
	server.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: server.Base()})

	type testScenario struct {
		name          string
		path          string
		customValidate func(resp *httpzclient.Resp) error
		expectedErrMsg string
		expectErrIs   error
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		opts := httpz.Options{
			Client:   client,
			Req:      httpzrequest.New("GET", s.path),
			Validate: s.customValidate,
		}
		err := httpz.Do(ctx, opts)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			if s.expectErrIs != nil {
				require.ErrorIs(t, err, s.expectErrIs)
			}
			return
		}
		require.NoError(t, err)
	}

	t.Run("Default validator succeeds on 200 OK", func(t *testing.T) {
		check(t, testScenario{
			name: "Default validator succeeds on 200 OK",
			path: "/ok",
		})
	})

	t.Run("Default validator fails on 400 Bad Request with body", func(t *testing.T) {
		check(t, testScenario{
			name:           "Default validator fails on 400 Bad Request with body",
			path:           "/bad-request",
			expectedErrMsg: "http validating error: 400, head body: invalid parameter, err: Http Status Error",
			expectErrIs:    httpz.ErrStatus,
		})
	})

	t.Run("Custom validator overrides default and accepts 400", func(t *testing.T) {
		check(t, testScenario{
			name: "Custom validator overrides default and accepts 400",
			path: "/bad-request",
			customValidate: func(resp *httpzclient.Resp) error {
				if resp.StatusCode == http.StatusBadRequest {
					return nil
				}
				return httpz.ErrStatus
			},
		})
	})

	t.Run("Custom validator rejects 200 with custom error", func(t *testing.T) {
		errCustom := errors.New("custom validation failure")
		check(t, testScenario{
			name: "Custom validator rejects 200 with custom error",
			path: "/ok",
			customValidate: func(resp *httpzclient.Resp) error {
				return errCustom
			},
			expectedErrMsg: "custom validation failure",
			expectErrIs:    errCustom,
		})
	})
}

func TestUnitDoFormatting(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	var receivedBody string
	server.AddHandler("POST", "/echo", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		b, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		receivedBody = string(b)
		w := resp(http.StatusOK, http.Header{})
		_, err = w.Write([]byte("ok"))
		require.NoError(t, err)
	})

	server.Listen()
	server.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: server.Base()})

	type testScenario struct {
		name           string
		preSetBody     bool
		formatFn       func(ctx context.Context) (io.ReadCloser, error)
		expectedBody   string
		expectedErrMsg string
		expectPanic    bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req := httpzrequest.New("POST", "/echo")
		if s.preSetBody {
			req.SetBody(strings.NewReader("already set"))
		}

		opts := httpz.Options{
			Client: client,
			Req:    req,
			Format: s.formatFn,
		}

		if s.expectPanic {
			require.Panics(t, func() {
				_ = httpz.Do(ctx, opts)
			})
			return
		}

		err := httpz.Do(ctx, opts)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			return
		}

		require.NoError(t, err)
		if s.expectedBody != "" {
			require.Equal(t, s.expectedBody, receivedBody)
		}
	}

	t.Run("Panics when request body is not nil initially", func(t *testing.T) {
		check(t, testScenario{
			name:        "Panics when request body is not nil initially",
			preSetBody:  true,
			expectPanic: true,
		})
	})

	t.Run("Format function populates body and closes reader", func(t *testing.T) {
		tracker := &trackingReader{Reader: strings.NewReader("formatted body content")}
		check(t, testScenario{
			name: "Format function populates body and closes reader",
			formatFn: func(ctx context.Context) (io.ReadCloser, error) {
				return tracker, nil
			},
			expectedBody: "formatted body content",
		})
		require.True(t, tracker.closed, "formatted reader must be closed after Do")
	})

	t.Run("Format function returning error causes Do to return wrapped error", func(t *testing.T) {
		check(t, testScenario{
			name: "Format function returning error causes Do to return wrapped error",
			formatFn: func(ctx context.Context) (io.ReadCloser, error) {
				return nil, errors.New("cannot format payload")
			},
			expectedErrMsg: "error creating request body: cannot format payload",
		})
	})
}

func TestUnitDoParsing(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	server.AddHandler("GET", "/parse", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{"X-Custom-Header": []string{"custom-val"}})
		_, err := w.Write([]byte("parsed payload"))
		require.NoError(t, err)
	})

	server.Listen()
	server.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: server.Base()})

	type testScenario struct {
		name           string
		parseFn        func(status int, headers http.Header, r io.Reader) error
		expectedErrMsg string
		expectedStatus int
		expectedHeader string
		expectedBody   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		opts := httpz.Options{
			Client: client,
			Req:    httpzrequest.New("GET", "/parse"),
			Parse:  s.parseFn,
		}

		err := httpz.Do(ctx, opts)
		if s.expectedErrMsg != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.expectedErrMsg)
			return
		}
		require.NoError(t, err)
	}

	t.Run("Without parse function completes successfully", func(t *testing.T) {
		check(t, testScenario{
			name:    "Without parse function completes successfully",
			parseFn: nil,
		})
	})

	t.Run("Parse function receives status headers and body correctly", func(t *testing.T) {
		var gotStatus int
		var gotHeader string
		var gotBody string

		check(t, testScenario{
			name: "Parse function receives status headers and body correctly",
			parseFn: func(status int, headers http.Header, r io.Reader) error {
				gotStatus = status
				gotHeader = headers.Get("X-Custom-Header")
				b, err := io.ReadAll(r)
				if err != nil {
					return fmt.Errorf("read failed: %w", err)
				}
				gotBody = string(b)
				return nil
			},
		})

		require.Equal(t, http.StatusOK, gotStatus)
		require.Equal(t, "custom-val", gotHeader)
		require.Equal(t, "parsed payload", gotBody)
	})

	t.Run("Parse function returning error causes Do to return wrapped error", func(t *testing.T) {
		check(t, testScenario{
			name: "Parse function returning error causes Do to return wrapped error",
			parseFn: func(status int, headers http.Header, r io.Reader) error {
				return errors.New("parse failed")
			},
			expectedErrMsg: "error parsing response: parse failed",
		})
	})
}

func TestUnitDoClientError(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	server.Listen()
	server.Start()
	addr := server.Base()
	server.Close() // Close immediately to trigger connection failure

	type testScenario struct {
		name           string
		expectedErrMsg string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: addr})
		opts := httpz.Options{
			Client: client,
			Req:    httpzrequest.New("GET", "/"),
		}
		err := httpz.Do(ctx, opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), s.expectedErrMsg)
	}

	t.Run("Client connection failure returns wrapped error", func(t *testing.T) {
		check(t, testScenario{
			name:           "Client connection failure returns wrapped error",
			expectedErrMsg: "failed to execute request",
		})
	})
}

func TestUnitMustDo(t *testing.T) {
	ctx := t.Context()
	server := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer server.Close()

	server.AddHandler("GET", "/ok", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, http.Header{})
		_, err := w.Write([]byte("success"))
		require.NoError(t, err)
	})

	server.AddHandler("GET", "/fail", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusInternalServerError, http.Header{})
		_, err := w.Write([]byte("error"))
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
		opts := httpz.Options{
			Client: client,
			Req:    httpzrequest.New("GET", s.path),
		}

		if s.expectPanic {
			require.Panics(t, func() {
				httpz.MustDo(ctx, opts)
			})
			return
		}

		require.NotPanics(t, func() {
			httpz.MustDo(ctx, opts)
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
