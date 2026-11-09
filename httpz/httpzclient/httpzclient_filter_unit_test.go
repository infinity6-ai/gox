package httpzclient_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitClientContextFilters(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	s.AddHandler("GET", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, nil)
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	})
	s.Listen()
	s.Start()

	type testScenario struct {
		name          string
		setupCtx      func(ctx context.Context, executionLog *[]string) context.Context
		setupClient   func(client *httpzclient.Client, executionLog *[]string)
		expectedOrder []string
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		var executionLog []string

		clientCtx := ctx
		if sc.setupCtx != nil {
			clientCtx = sc.setupCtx(clientCtx, &executionLog)
		}

		client := httpzclient.New(clientCtx, httpzclient.Options{BaseUrl: s.Base()})
		if sc.setupClient != nil {
			sc.setupClient(client, &executionLog)
		}

		req := httpzrequest.New("GET", "/test")
		resp, err := client.Do(clientCtx, req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Equal(t, sc.expectedOrder, executionLog)
	}

	t.Run("Only outer filters", func(t *testing.T) {
		check(t, testScenario{
			name: "Only outer filters",
			setupCtx: func(ctx context.Context, executionLog *[]string) context.Context {
				f1 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "outer1-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "outer1-out")
					return resp, err
				}
				f2 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "outer2-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "outer2-out")
					return resp, err
				}
				return httpzclient.WithFilters(ctx, []httpzclient.Filter{f1, f2}, nil)
			},
			expectedOrder: []string{"outer1-in", "outer2-in", "outer2-out", "outer1-out"},
		})
	})

	t.Run("Only inner filters", func(t *testing.T) {
		check(t, testScenario{
			name: "Only inner filters",
			setupCtx: func(ctx context.Context, executionLog *[]string) context.Context {
				f1 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "inner1-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "inner1-out")
					return resp, err
				}
				f2 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "inner2-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "inner2-out")
					return resp, err
				}
				return httpzclient.WithFilters(ctx, nil, []httpzclient.Filter{f1, f2})
			},
			expectedOrder: []string{"inner1-in", "inner2-in", "inner2-out", "inner1-out"},
		})
	})

	t.Run("Client and context filters execute in correct order", func(t *testing.T) {
		check(t, testScenario{
			name: "Client and context filters execute in correct order",
			setupCtx: func(ctx context.Context, executionLog *[]string) context.Context {
				fOut := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "outer-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "outer-out")
					return resp, err
				}
				fIn := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "inner-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "inner-out")
					return resp, err
				}
				return httpzclient.WithFilters(ctx, []httpzclient.Filter{fOut}, []httpzclient.Filter{fIn})
			},
			setupClient: func(client *httpzclient.Client, executionLog *[]string) {
				client.AddFilter(func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "client-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "client-out")
					return resp, err
				})
			},
			expectedOrder: []string{"outer-in", "client-in", "inner-in", "inner-out", "client-out", "outer-out"},
		})
	})

	t.Run("Multiple WithFilters calls stack correctly", func(t *testing.T) {
		check(t, testScenario{
			name: "Multiple WithFilters calls stack correctly",
			setupCtx: func(ctx context.Context, executionLog *[]string) context.Context {
				fOut1 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "outer1-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "outer1-out")
					return resp, err
				}
				fIn1 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "inner1-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "inner1-out")
					return resp, err
				}

				fOut2 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "outer2-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "outer2-out")
					return resp, err
				}
				fIn2 := func(ctx context.Context, req *httpzrequest.Req, next httpzclient.Handler) (*httpzclient.Resp, error) {
					*executionLog = append(*executionLog, "inner2-in")
					resp, err := next(ctx, req)
					*executionLog = append(*executionLog, "inner2-out")
					return resp, err
				}
				ctx = httpzclient.WithFilters(ctx, []httpzclient.Filter{fOut1}, []httpzclient.Filter{fIn1})
				ctx = httpzclient.WithFilters(ctx, []httpzclient.Filter{fOut2}, []httpzclient.Filter{fIn2})
				return ctx
			},
			expectedOrder: []string{"outer1-in", "outer2-in", "inner1-in", "inner2-in", "inner2-out", "inner1-out", "outer2-out", "outer1-out"},
		})
	})
}
