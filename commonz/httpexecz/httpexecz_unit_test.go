package httpexecz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnitHttpexeczBuildArgs(t *testing.T) {
	type testScenario struct {
		name       string
		method     string
		url        string
		hasBody    bool
		headers    http.Header
		host       string
		reqClose   bool
		opts       Options
		wantArgs   []string
		wantSubstr []string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req, err := http.NewRequest(s.method, s.url, nil)
		require.NoError(t, err)
		if s.headers != nil {
			req.Header = s.headers
		}
		if s.host != "" {
			req.Host = s.host
		}
		req.Close = s.reqClose

		tr := NewTransport(s.opts)
		args, err := tr.buildArgs(req, s.hasBody)
		require.NoError(t, err)

		if len(s.wantArgs) > 0 {
			require.Equal(t, s.wantArgs, args)
		}
		for _, sub := range s.wantSubstr {
			require.Contains(t, args, sub)
		}
	}

	t.Run("Default GET request args", func(t *testing.T) {
		check(t, testScenario{
			method:   "GET",
			url:      "http://example.com/api",
			hasBody:  false,
			wantArgs: []string{"-s", "--http1.1", "-i", "http://example.com/api"},
		})
	})

	t.Run("HEAD request args use -I without -i or -X", func(t *testing.T) {
		check(t, testScenario{
			method:   "HEAD",
			url:      "http://example.com/api",
			hasBody:  false,
			wantArgs: []string{"-s", "--http1.1", "-I", "http://example.com/api"},
		})
	})

	t.Run("POST request with body", func(t *testing.T) {
		check(t, testScenario{
			method:   "POST",
			url:      "http://example.com/api",
			hasBody:  true,
			wantArgs: []string{"-s", "--http1.1", "-i", "-X", "POST", "--data-binary", "@-", "http://example.com/api"},
		})
	})

	t.Run("Disable HTTP11 omits flag", func(t *testing.T) {
		check(t, testScenario{
			method:   "GET",
			url:      "http://example.com/api",
			hasBody:  false,
			opts:     Options{DisableHTTP11: true},
			wantArgs: []string{"-s", "-i", "http://example.com/api"},
		})
	})

	t.Run("FollowRedirects adds -L flag", func(t *testing.T) {
		check(t, testScenario{
			method:   "GET",
			url:      "http://example.com/api",
			hasBody:  false,
			opts:     Options{FollowRedirects: true},
			wantArgs: []string{"-s", "--http1.1", "-L", "-i", "http://example.com/api"},
		})
	})

	t.Run("Custom headers and Host", func(t *testing.T) {
		check(t, testScenario{
			method:  "GET",
			url:     "http://example.com/api",
			hasBody: false,
			host:    "custom.host.com",
			headers: http.Header{
				"X-Custom": []string{"value1"},
				"Accept":   []string{"application/json"},
			},
			wantSubstr: []string{
				"-H", "Host: custom.host.com",
				"-H", "Accept: application/json",
				"-H", "X-Custom: value1",
			},
		})
	})

	t.Run("Connection close header added when req.Close is true", func(t *testing.T) {
		check(t, testScenario{
			method:     "GET",
			url:        "http://example.com/api",
			hasBody:    false,
			reqClose:   true,
			wantSubstr: []string{"-H", "Connection: close"},
		})
	})

	t.Run("ExtraArgs appended correctly", func(t *testing.T) {
		check(t, testScenario{
			method:   "GET",
			url:      "http://example.com/api",
			hasBody:  false,
			opts:     Options{ExtraArgs: []string{"--retry", "3"}},
			wantArgs: []string{"-s", "--http1.1", "-i", "--retry", "3", "http://example.com/api"},
		})
	})
}

func TestUnitHttpexeczMethods(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Method", r.Method)
		w.Header().Set("X-Received-Header", r.Header.Get("X-Test-Header"))
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil && len(bodyBytes) > 0 {
				w.Header().Set("X-Received-Body-Len", fmt.Sprintf("%d", len(bodyBytes)))
				_, _ = w.Write(bodyBytes)
				return
			}
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	type testScenario struct {
		method       string
		body         string
		header       string
		wantStatus   int
		wantBody     string
		wantMethod   string
		wantHeader   string
		wantBodyLen  string
		isHeadMethod bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var bodyReader io.Reader
		if s.body != "" {
			bodyReader = strings.NewReader(s.body)
		}

		req, err := http.NewRequest(s.method, ts.URL, bodyReader)
		require.NoError(t, err)
		if s.header != "" {
			req.Header.Set("X-Test-Header", s.header)
		}

		tr := NewTransport(Options{})
		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.wantStatus, resp.StatusCode)
		require.Equal(t, s.wantMethod, resp.Header.Get("X-Received-Method"))
		if s.wantHeader != "" {
			require.Equal(t, s.wantHeader, resp.Header.Get("X-Received-Header"))
		}
		if s.wantBodyLen != "" {
			require.Equal(t, s.wantBodyLen, resp.Header.Get("X-Received-Body-Len"))
		}

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if s.isHeadMethod {
			require.Empty(t, respBody)
		} else {
			require.Equal(t, s.wantBody, string(respBody))
		}
	}

	t.Run("GET method", func(t *testing.T) {
		check(t, testScenario{
			method:     "GET",
			header:     "my-get-token",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
			wantMethod: "GET",
			wantHeader: "my-get-token",
		})
	})

	t.Run("HEAD method", func(t *testing.T) {
		check(t, testScenario{
			method:       "HEAD",
			wantStatus:   http.StatusOK,
			wantMethod:   "HEAD",
			isHeadMethod: true,
		})
	})

	t.Run("POST method with body", func(t *testing.T) {
		check(t, testScenario{
			method:      "POST",
			body:        `{"message":"hello"}`,
			wantStatus:  http.StatusOK,
			wantBody:    `{"message":"hello"}`,
			wantMethod:  "POST",
			wantBodyLen: fmt.Sprintf("%d", len(`{"message":"hello"}`)),
		})
	})

	t.Run("PUT method with body", func(t *testing.T) {
		check(t, testScenario{
			method:      "PUT",
			body:        "updated content",
			wantStatus:  http.StatusOK,
			wantBody:    "updated content",
			wantMethod:  "PUT",
			wantBodyLen: fmt.Sprintf("%d", len("updated content")),
		})
	})

	t.Run("PATCH method with body", func(t *testing.T) {
		check(t, testScenario{
			method:      "PATCH",
			body:        "partial content",
			wantStatus:  http.StatusOK,
			wantBody:    "partial content",
			wantMethod:  "PATCH",
			wantBodyLen: fmt.Sprintf("%d", len("partial content")),
		})
	})

	t.Run("DELETE method without body", func(t *testing.T) {
		check(t, testScenario{
			method:     "DELETE",
			wantStatus: http.StatusOK,
			wantBody:   "ok",
			wantMethod: "DELETE",
		})
	})
}

func TestUnitHttpexeczStatusCodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/created":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("created"))
		case "/bad-request":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request error"))
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found error"))
		case "/internal-error":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server error"))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	type testScenario struct {
		path       string
		wantStatus int
		wantBody   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req, err := http.NewRequest("GET", ts.URL+s.path, nil)
		require.NoError(t, err)

		tr := NewTransport(Options{})
		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.wantStatus, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, s.wantBody, string(body))
	}

	t.Run("201 Created", func(t *testing.T) {
		check(t, testScenario{
			path:       "/created",
			wantStatus: http.StatusCreated,
			wantBody:   "created",
		})
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		check(t, testScenario{
			path:       "/bad-request",
			wantStatus: http.StatusBadRequest,
			wantBody:   "bad request error",
		})
	})

	t.Run("404 Not Found", func(t *testing.T) {
		check(t, testScenario{
			path:       "/not-found",
			wantStatus: http.StatusNotFound,
			wantBody:   "not found error",
		})
	})

	t.Run("500 Internal Server Error", func(t *testing.T) {
		check(t, testScenario{
			path:       "/internal-error",
			wantStatus: http.StatusInternalServerError,
			wantBody:   "server error",
		})
	})
}

func TestUnitHttpexeczRedirects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		if r.URL.Path == "/target" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("redirected content"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	type testScenario struct {
		followRedirects bool
		wantStatus      int
		wantBody        string
		wantLocation    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		req, err := http.NewRequest("GET", ts.URL+"/redirect", nil)
		require.NoError(t, err)

		tr := NewTransport(Options{
			FollowRedirects: s.followRedirects,
		})
		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.wantStatus, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		if s.wantBody != "" {
			require.Equal(t, s.wantBody, string(body))
		}
		if s.wantLocation != "" {
			require.Equal(t, s.wantLocation, resp.Header.Get("Location"))
		}
	}

	t.Run("FollowRedirects false returns 302 with Location", func(t *testing.T) {
		check(t, testScenario{
			followRedirects: false,
			wantStatus:      http.StatusFound,
			wantLocation:    "/target",
		})
	})

	t.Run("FollowRedirects true follows location to 200", func(t *testing.T) {
		check(t, testScenario{
			followRedirects: true,
			wantStatus:      http.StatusOK,
			wantBody:        "redirected content",
		})
	})
}

func TestUnitHttpexeczTimeoutAndCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("delayed response"))
	}))
	defer ts.Close()

	type testScenario struct {
		cancelDirectly bool
		clientTimeout  time.Duration
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := context.Background()
		var cancel context.CancelFunc
		if s.cancelDirectly {
			ctx, cancel = context.WithCancel(ctx)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", ts.URL, nil)
		require.NoError(t, err)

		tr := NewTransport(Options{
			Timeout: s.clientTimeout,
		})

		if s.cancelDirectly {
			go func() {
				time.Sleep(20 * time.Millisecond)
				cancel()
			}()
		}

		start := time.Now()
		resp, err := tr.RoundTrip(req)
		elapsed := time.Since(start)

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		require.Error(t, err)
		require.True(t, elapsed < 400*time.Millisecond, "request should have aborted before server finished delay")
	}

	t.Run("Request context cancellation aborts process quickly", func(t *testing.T) {
		check(t, testScenario{
			cancelDirectly: true,
		})
	})

	t.Run("Transport timeout aborts process quickly", func(t *testing.T) {
		check(t, testScenario{
			clientTimeout: 50 * time.Millisecond,
		})
	})
}

func TestUnitHttpexeczErrors(t *testing.T) {
	type testScenario struct {
		name          string
		req           func() *http.Request
		opts          Options
		wantErrSubstr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		tr := NewTransport(s.opts)
		resp, err := tr.RoundTrip(s.req())
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		require.Error(t, err)
		require.Contains(t, err.Error(), s.wantErrSubstr)
	}

	t.Run("Nil request URL returns error", func(t *testing.T) {
		check(t, testScenario{
			req: func() *http.Request {
				return &http.Request{}
			},
			wantErrSubstr: "request URL is nil",
		})
	})

	t.Run("Non-existent binary command returns error", func(t *testing.T) {
		check(t, testScenario{
			req: func() *http.Request {
				r, _ := http.NewRequest("GET", "http://localhost:1", nil)
				return r
			},
			opts:          Options{Command: "non_existent_binary_xyz_123"},
			wantErrSubstr: "failed to start command",
		})
	})

	t.Run("Connection refused returns wrapped error with stderr", func(t *testing.T) {
		check(t, testScenario{
			req: func() *http.Request {
				// Use an unused high port on localhost
				r, _ := http.NewRequest("GET", "http://127.0.0.1:49151/invalid", nil)
				return r
			},
			opts:          Options{Timeout: 2 * time.Second},
			wantErrSubstr: "httpexecz",
		})
	})
}

func TestUnitHttpexeczNewClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("client ok"))
	}))
	defer ts.Close()

	type testScenario struct {
		url      string
		wantBody string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		client := NewClient(Options{})
		resp, err := client.Get(s.url)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, s.wantBody, string(body))
	}

	t.Run("NewClient works transparently with standard http.Client", func(t *testing.T) {
		check(t, testScenario{
			url:      ts.URL,
			wantBody: "client ok",
		})
	})
}
