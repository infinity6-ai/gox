package httpzclient_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/infinity6-ai/gox/httpz/httpzclient"
	"github.com/infinity6-ai/gox/httpz/httpzrequest"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

type dummyReader struct {
	r io.Reader
}

func (d dummyReader) Read(p []byte) (int, error) {
	return d.r.Read(p)
}

type dummyReadCloser struct {
	r      io.Reader
	closed bool
}

func (d *dummyReadCloser) Read(p []byte) (int, error) {
	return d.r.Read(p)
}

func (d *dummyReadCloser) Close() error {
	d.closed = true
	return nil
}

func TestUnitClientContentLength(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()

	var receivedHeader atomic.Pointer[string]
	var receivedBody atomic.Pointer[string]

	s.AddHandler("POST", "/test", func(ctx context.Context, resp httpzserver.Resp, req *httpzrequest.Req, params map[string]string) {
		cl := req.Headers.Get("Content-Length")
		receivedHeader.Store(&cl)

		var bodyStr string
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err == nil {
				bodyStr = string(b)
			}
		}
		receivedBody.Store(&bodyStr)

		w := resp(http.StatusOK, http.Header{})
		_, _ = w.Write([]byte("ok"))
	})
	s.Listen()
	s.Start()

	client := httpzclient.New(ctx, httpzclient.Options{BaseUrl: s.Base()})

	type testScenario struct {
		body                  io.Reader
		headers               map[string]string
		rawHeaders            http.Header
		expectedContentLength string
		expectedBody          string
		expectedErr           string
		checkCloser           *dummyReadCloser
	}

	check := func(t *testing.T, sc testScenario) {
		t.Helper()
		empty := ""
		receivedHeader.Store(&empty)
		receivedBody.Store(&empty)

		req := httpzrequest.New("POST", "/test")
		if sc.body != nil {
			req.SetBody(sc.body)
		}
		for k, v := range sc.headers {
			req.AddHeader(k, v)
		}
		for k, v := range sc.rawHeaders {
			req.Headers[k] = v
		}

		resp, err := client.Do(ctx, req)
		if sc.expectedErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), sc.expectedErr)
			return
		}

		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, sc.expectedContentLength, *receivedHeader.Load())
		require.Equal(t, sc.expectedBody, *receivedBody.Load())

		if sc.checkCloser != nil {
			require.True(t, sc.checkCloser.closed)
		}
	}

	t.Run("Generic reader with Content-Length header sends explicit ContentLength", func(t *testing.T) {
		check(t, testScenario{
			body:                  dummyReader{r: strings.NewReader("hello-world")},
			headers:               map[string]string{"Content-Length": "11"},
			expectedContentLength: "11",
			expectedBody:          "hello-world",
		})
	})

	t.Run("Lowercase content-length header is recognized case-insensitively", func(t *testing.T) {
		check(t, testScenario{
			body:                  dummyReader{r: strings.NewReader("case-insensitive")},
			rawHeaders:            http.Header{"content-length": []string{"16"}},
			expectedContentLength: "16",
			expectedBody:          "case-insensitive",
		})
	})

	t.Run("Generic reader without Content-Length defaults to chunked transfer encoding", func(t *testing.T) {
		check(t, testScenario{
			body:                  dummyReader{r: strings.NewReader("chunked-body")},
			expectedContentLength: "",
			expectedBody:          "chunked-body",
		})
	})

	t.Run("Content-Length 0 with empty body closes stream and sends 0 length without chunked", func(t *testing.T) {
		rc := &dummyReadCloser{r: strings.NewReader("")}
		check(t, testScenario{
			body:                  rc,
			headers:               map[string]string{"Content-Length": "0"},
			expectedContentLength: "0",
			expectedBody:          "",
			checkCloser:           rc,
		})
	})

	t.Run("Content-Length 0 with nil body sends 0 length", func(t *testing.T) {
		check(t, testScenario{
			body:                  nil,
			headers:               map[string]string{"Content-Length": "0"},
			expectedContentLength: "0",
			expectedBody:          "",
		})
	})

	t.Run("Invalid non-numeric Content-Length returns descriptive error", func(t *testing.T) {
		check(t, testScenario{
			body:        dummyReader{r: strings.NewReader("data")},
			headers:     map[string]string{"Content-Length": "not-a-number"},
			expectedErr: "failed to parse Content-Length header",
		})
	})

	t.Run("Negative Content-Length returns error", func(t *testing.T) {
		check(t, testScenario{
			body:        dummyReader{r: strings.NewReader("data")},
			headers:     map[string]string{"Content-Length": "-5"},
			expectedErr: "invalid negative Content-Length header",
		})
	})
}
