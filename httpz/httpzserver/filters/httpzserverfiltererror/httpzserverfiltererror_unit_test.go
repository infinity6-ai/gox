package httpzserverfiltererror_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/httpz/httpz"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfiltererror"
	"github.com/stretchr/testify/require"
)

func TestUnitFilterError(t *testing.T) {
	type testScenario struct {
		name                 string
		panicValue           any
		responseStarted      bool
		expectedStatusCode   int
		expectedBodyContains string
		expectedLogMessage   string
		expectedPanic        bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		srv.AddFilter(httpzserverfiltererror.Filter)

		srv.AddHandler("GET", "/panic", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			if s.responseStarted {
				w := resp(http.StatusOK, nil)
				_, err := w.Write([]byte("ok"))
				require.NoError(t, err)
			}
			panic(s.panicValue)
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/panic").String(), nil)
		require.NoError(t, err)

		collector := httpzserverfiltererror.Collector()

		if s.expectedPanic {
			require.Panics(t, func() {
				_, _ = http.DefaultClient.Do(req)
			})
			
			logs := collector()
			require.NotEmpty(t, logs)
			lastLog := logs[len(logs)-1]
			require.Equal(t, s.expectedLogMessage, lastLog.Operation)
			return
		}

		httpResp, err := http.DefaultClient.Do(req)

		if s.responseStarted {
			require.Error(t, err)
			logs := collector()
			require.NotEmpty(t, logs)
			lastLog := logs[len(logs)-1]
			require.Equal(t, s.expectedLogMessage, lastLog.Operation)
			return
		}

		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedStatusCode, httpResp.StatusCode)
		
		body, err := io.ReadAll(httpResp.Body)
		require.NoError(t, err)
		if s.expectedBodyContains != "" {
			require.Contains(t, string(body), s.expectedBodyContains)
		}

		require.Equal(t, "true", httpResp.Header.Get(httpz.HeaderError))
		require.Equal(t, "application/json", httpResp.Header.Get("Content-Type"))

		logs := collector()
		require.NotEmpty(t, logs)
		lastLog := logs[len(logs)-1]
		require.Equal(t, s.expectedLogMessage, lastLog.Operation)
	}

	t.Run("Standard error panic", func(t *testing.T) {
		check(t, testScenario{
			name:                 "Standard error panic",
			panicValue:           fmt.Errorf("something went wrong"),
			responseStarted:      false,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: `"name":"InternalError"`,
			expectedLogMessage:   "http handler panic",
			expectedPanic:        false,
		})
	})

	t.Run("String panic", func(t *testing.T) {
		check(t, testScenario{
			name:                 "String panic",
			panicValue:           "string panic value",
			responseStarted:      false,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: `"name":"InternalError"`,
			expectedLogMessage:   "http handler panic",
			expectedPanic:        false,
		})
	})

	t.Run("Business error panic", func(t *testing.T) {
		bizErr := errorz.Detail(http.StatusBadRequest, "invalid input", "username", true, fmt.Errorf("invalid input"))
		check(t, testScenario{
			name:                 "Business error panic",
			panicValue:           bizErr,
			responseStarted:      false,
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: `"name":"invalid input"`,
			expectedLogMessage:   "http handler panic",
			expectedPanic:        false,
		})
	})

	t.Run("Response already started panics again", func(t *testing.T) {
		// When the server panics in a handler, if it panics again inside the filter, the http client might just see an EOF or connection reset. 
		// Actually, standard http.Server recovers panics and closes the connection.
		// So http.DefaultClient.Do(req) will return an error (EOF).
		check(t, testScenario{
			name:                 "Response already started panics again",
			panicValue:           fmt.Errorf("late panic"),
			responseStarted:      true,
			expectedStatusCode:   0,
			expectedBodyContains: "",
			expectedLogMessage:   "http handler panic",
			expectedPanic:        false, 
		})
	})
}
