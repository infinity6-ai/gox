package httpzserverfiltererror_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestUnitErrorFilterBasic(t *testing.T) {
	type testScenario struct {
		name              string
		panicVal          any
		normalBody        string
		normalStatus      int
		expectedStatus    int
		expectedErrorzHdr bool
		expectedCode      int
		expectedName      string
		expectedPayload   string
		expectedBusiness  bool
		expectedCause     string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := t.Context()
		srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
		defer srv.Close()
		srv.Listen()

		collector := httpzserverfiltererror.Collector()
		collector() // clear prior entries

		srv.AddFilter(httpzserverfiltererror.Filter)

		srv.AddHandler("GET", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
			if s.panicVal != nil {
				panic(s.panicVal)
			}
			w := resp(s.normalStatus, nil)
			_, err := w.Write([]byte(s.normalBody))
			require.NoError(t, err)
		})

		srv.Start()

		req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/test").String(), nil)
		require.NoError(t, err)

		httpResp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer httpResp.Body.Close()

		require.Equal(t, s.expectedStatus, httpResp.StatusCode)

		if s.expectedErrorzHdr {
			require.Equal(t, "true", httpResp.Header.Get(httpz.HeaderError))
			require.Equal(t, "application/json", httpResp.Header.Get("Content-Type"))

			var data errorz.Data
			err := json.NewDecoder(httpResp.Body).Decode(&data)
			require.NoError(t, err)

			require.Equal(t, s.expectedCode, data.Code)
			require.Equal(t, s.expectedName, data.Name)
			require.Equal(t, s.expectedPayload, data.Payload)
			require.Equal(t, s.expectedBusiness, data.Business)
			require.Equal(t, s.expectedCause, data.Cause)

			// Verify error log was emitted
			entries := collector()
			require.Len(t, entries, 1)
			require.Equal(t, "http handler panic", entries[0].Operation)
		} else {
			require.Empty(t, httpResp.Header.Get(httpz.HeaderError))
			bodyBytes, err := io.ReadAll(httpResp.Body)
			require.NoError(t, err)
			require.Equal(t, s.normalBody, string(bodyBytes))
		}
	}

	t.Run("Normal execution without panic", func(t *testing.T) {
		check(t, testScenario{
			normalStatus:      http.StatusOK,
			normalBody:        "success",
			expectedStatus:    http.StatusOK,
			expectedErrorzHdr: false,
		})
	})

	t.Run("Panic with business StructuredError returns code, errorz JSON, and X-Errorz header", func(t *testing.T) {
		check(t, testScenario{
			panicVal:          errorz.Detail(404, "UserNotFound", "user-42", true, errors.New("user not found")),
			expectedStatus:    http.StatusNotFound,
			expectedErrorzHdr: true,
			expectedCode:      404,
			expectedName:      "UserNotFound",
			expectedPayload:   "user-42",
			expectedBusiness:  true,
			expectedCause:     "user not found",
		})
	})

	t.Run("Panic with business StructuredError with code 0 defaults to 400", func(t *testing.T) {
		check(t, testScenario{
			panicVal:          errorz.Detail(0, "InvalidInput", "missing field", true, errors.New("field required")),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorzHdr: true,
			expectedCode:      http.StatusBadRequest,
			expectedName:      "InvalidInput",
			expectedPayload:   "missing field",
			expectedBusiness:  true,
			expectedCause:     "field required",
		})
	})

	t.Run("Panic with non-business StructuredError sanitizes to 500 InternalError", func(t *testing.T) {
		check(t, testScenario{
			panicVal:          errorz.Detail(500, "SqlError", "secret connection string", false, errors.New("db down")),
			expectedStatus:    http.StatusInternalServerError,
			expectedErrorzHdr: true,
			expectedCode:      500,
			expectedName:      "InternalError",
			expectedPayload:   "",
			expectedBusiness:  false,
			expectedCause:     "InternalError",
		})
	})

	t.Run("Panic with standard Go error sanitizes to 500 InternalError", func(t *testing.T) {
		check(t, testScenario{
			panicVal:          errors.New("unhandled internal pointer dereference"),
			expectedStatus:    http.StatusInternalServerError,
			expectedErrorzHdr: true,
			expectedCode:      500,
			expectedName:      "InternalError",
			expectedPayload:   "",
			expectedBusiness:  false,
			expectedCause:     "InternalError",
		})
	})

	t.Run("Panic with string value sanitizes to 500 InternalError", func(t *testing.T) {
		check(t, testScenario{
			panicVal:          "raw string panic message",
			expectedStatus:    http.StatusInternalServerError,
			expectedErrorzHdr: true,
			expectedCode:      500,
			expectedName:      "InternalError",
			expectedPayload:   "",
			expectedBusiness:  false,
			expectedCause:     "InternalError",
		})
	})
}

func TestUnitErrorFilterResponseAlreadyStarted(t *testing.T) {
	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()
	srv.Listen()

	collector := httpzserverfiltererror.Collector()
	collector()

	srv.AddFilter(httpzserverfiltererror.Filter)

	srv.AddHandler("GET", "/panic-after-write", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		w := resp(http.StatusOK, nil)
		_, err := w.Write([]byte("first-part-"))
		require.NoError(t, err)
		panic(errors.New("crash after write started"))
	})

	srv.Start()

	req, err := http.NewRequest(http.MethodGet, srv.Base().MustJoinPathString("/panic-after-write").String(), nil)
	require.NoError(t, err)

	httpResp, err := http.DefaultClient.Do(req)
	if err == nil {
		defer httpResp.Body.Close()
	}

	// Verify the logger recorded the panic
	entries := collector()
	require.Len(t, entries, 1)
	require.Equal(t, "http handler panic", entries[0].Operation)
}

// func TestUnitErrorFilterWithAuditAndGzip(t *testing.T) {
// 	ctx := t.Context()
// 	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
// 	defer srv.Close()
// 	srv.Listen()

// 	auditCollector := httpzserverfilteraudit.Collector()
// 	auditCollector()
// 	errorCollector := httpzserverfiltererror.Collector()
// 	errorCollector()

// 	// Pipeline order: Gzip -> Audit -> Error -> Handler
// 	srv.AddFilter(httpzserverfiltergzip.GzipFilter)
// 	// srv.AddFilter(httpzserverfilteraudit.AuditFilter)
// 	srv.AddFilter(httpzserverfiltererror.Filter)

// 	srv.AddHandler("POST", "/api/item", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
// 		panic(errorz.Detail(422, "ValidationFailed", "name too short", true, errors.New("validation failed")))
// 	})

// 	srv.Start()

// 	req, err := http.NewRequest(http.MethodPost, srv.Base().MustJoinPathString("/api/item").String(), nil)
// 	require.NoError(t, err)
// 	req.Header.Set("Accept-Encoding", "gzip")

// 	httpResp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)
// 	defer httpResp.Body.Close()

// 	require.Equal(t, 422, httpResp.StatusCode)
// 	require.Equal(t, "true", httpResp.Header.Get(httpz.HeaderError))
// 	require.Equal(t, "gzip", httpResp.Header.Get("Content-Encoding"))

// 	// Decompress and verify JSON payload
// 	gr, err := gzip.NewReader(httpResp.Body)
// 	require.NoError(t, err)
// 	defer gr.Close()

// 	var data errorz.Data
// 	err = json.NewDecoder(gr).Decode(&data)
// 	require.NoError(t, err)
// 	require.Equal(t, 422, data.Code)
// 	require.Equal(t, "ValidationFailed", data.Name)
// 	require.Equal(t, "name too short", data.Payload)
// 	require.True(t, data.Business)

// 	// Verify AuditFilter audited the response with uncompressed body!
// 	auditEntries := auditCollector()
// 	require.Len(t, auditEntries, 1)
// 	audit := auditEntries[0].Audit.ReqResp
// 	require.NotNil(t, audit)
// 	require.NotNil(t, audit.Resp)
// 	require.Equal(t, 422, audit.Resp.Status)
// 	require.Equal(t, "true", audit.Resp.Headers.Get(httpz.HeaderError))
// 	require.NotNil(t, audit.Resp.Body)
// 	require.Contains(t, string(audit.Resp.Body.Payload), "ValidationFailed")
// 	require.Contains(t, string(audit.Resp.Body.Payload), "name too short")

// 	// Verify ErrorFilter logged the panic
// 	errorEntries := errorCollector()
// 	require.Len(t, errorEntries, 1)
// 	require.Equal(t, "http handler panic", errorEntries[0].Operation)
// }
