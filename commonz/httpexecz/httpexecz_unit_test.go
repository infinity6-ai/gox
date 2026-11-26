package httpexecz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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
		args, _, err := tr.buildArgs(req, s.hasBody, "test-dump")
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

	t.Run("Trace adds --trace flag", func(t *testing.T) {
		check(t, testScenario{
			method:     "GET",
			url:        "http://example.com/api",
			hasBody:    false,
			opts:       Options{Trace: "/tmp/traces"},
			wantSubstr: []string{"--trace"},
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
		useClient    bool
		stopRedirect bool
		wantStatus   int
		wantBody     string
		wantLocation string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.useClient {
			client := NewClient(Options{})
			if s.stopRedirect {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}
			resp, err := client.Get(ts.URL + "/redirect")
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
			return
		}

		req, err := http.NewRequest("GET", ts.URL+"/redirect", nil)
		require.NoError(t, err)

		tr := NewTransport(Options{})
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

	t.Run("RoundTrip returns 302 Found directly with Location header", func(t *testing.T) {
		check(t, testScenario{
			useClient:    false,
			wantStatus:   http.StatusFound,
			wantLocation: "/target",
		})
	})

	t.Run("Standard http.Client follows redirect to 200 OK via Go redirect handler", func(t *testing.T) {
		check(t, testScenario{
			useClient:  true,
			wantStatus: http.StatusOK,
			wantBody:   "redirected content",
		})
	})

	t.Run("Client with custom CheckRedirect halts redirect and returns 302", func(t *testing.T) {
		check(t, testScenario{
			useClient:    true,
			stopRedirect: true,
			wantStatus:   http.StatusFound,
			wantLocation: "/target",
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

func TestUnitHttpexeczBeforeExec(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Injected", r.Header.Get("X-Injected"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hook ok"))
	}))
	defer ts.Close()

	type testScenario struct {
		beforeExec   func(cmd string, args []string) (string, []string, error)
		wantErr      string
		wantInjected string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		tr := NewTransport(Options{
			BeforeExec: s.beforeExec,
		})
		req, err := http.NewRequest("GET", ts.URL, nil)
		require.NoError(t, err)

		resp, err := tr.RoundTrip(req)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}

		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, s.wantInjected, resp.Header.Get("X-Injected"))
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "hook ok", string(body))
	}

	t.Run("BeforeExec injects argument successfully", func(t *testing.T) {
		check(t, testScenario{
			beforeExec: func(cmd string, args []string) (string, []string, error) {
				// Inject header before the last argument (URL)
				urlArg := args[len(args)-1]
				newArgs := append(args[:len(args)-1], "-H", "X-Injected: from-hook", urlArg)
				return cmd, newArgs, nil
			},
			wantInjected: "from-hook",
		})
	})

	t.Run("BeforeExec error aborts round-trip", func(t *testing.T) {
		check(t, testScenario{
			beforeExec: func(cmd string, args []string) (string, []string, error) {
				return "", nil, fmt.Errorf("blocked by security policy")
			},
			wantErr: "BeforeExec failed: blocked by security policy",
		})
	})

	t.Run("BeforeExec can mutate command binary", func(t *testing.T) {
		check(t, testScenario{
			beforeExec: func(cmd string, args []string) (string, []string, error) {
				return "non_existent_binary_from_hook", args, nil
			},
			wantErr: "failed to start command \"non_existent_binary_from_hook\"",
		})
	})
}

func TestUnitHttpexeczTrace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Trace-Response", "pong")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("trace body ok"))
	}))
	defer ts.Close()

	type testScenario struct {
		failRequest   bool
		urlPath       string
		wantBody      string
		wantInTrace   []string
		wantErrSubstr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		traceDir := t.TempDir()

		tr := NewTransport(Options{
			Trace:   traceDir,
			Timeout: 2 * time.Second,
		})

		targetURL := ts.URL + s.urlPath
		if s.failRequest {
			targetURL = "http://127.0.0.1:49151/fail"
		}

		req, err := http.NewRequest("GET", targetURL, nil)
		require.NoError(t, err)

		resp, err := tr.RoundTrip(req)
		if s.wantErrSubstr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErrSubstr)
			require.Contains(t, err.Error(), traceDir)
		} else {
			require.NoError(t, err)
			defer resp.Body.Close()

			body, readErr := io.ReadAll(resp.Body)
			require.NoError(t, readErr)
			require.Equal(t, s.wantBody, string(body))
		}

		entries, err := os.ReadDir(traceDir)
		require.NoError(t, err)
		require.Len(t, entries, 1)

		filename := entries[0].Name()
		require.True(t, strings.HasPrefix(filename, "dump-"), "filename should start with dump-")
		require.True(t, strings.HasSuffix(filename, ".log"), "filename should end with .log")

		content, err := os.ReadFile(filepath.Join(traceDir, filename))
		require.NoError(t, err)
		require.NotEmpty(t, content)

		for _, sub := range s.wantInTrace {
			require.Contains(t, string(content), sub)
		}
	}

	t.Run("Successful request writes trace dump file with wire headers and body", func(t *testing.T) {
		check(t, testScenario{
			urlPath:  "/api/users",
			wantBody: "trace body ok",
			wantInTrace: []string{
				"X-Trace-Response",
				"trace body ok",
			},
		})
	})

	t.Run("Failed request writes trace dump file and includes path in error message", func(t *testing.T) {
		check(t, testScenario{
			failRequest:   true,
			wantErrSubstr: "(trace: ",
		})
	})
}

func TestUnitHttpexeczOutScript(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Method", r.Method)
		if r.Body != nil {
			b, err := io.ReadAll(r.Body)
			if err == nil && len(b) > 0 {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("get response ok"))
	}))
	defer ts.Close()

	type testScenario struct {
		method          string
		payload         string
		headers         http.Header
		wantInScript    []string
		wantPayloadFile bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		scriptDir := t.TempDir()

		tr := NewTransport(Options{
			OutScript: scriptDir,
		})

		var bodyReader io.Reader
		if s.payload != "" {
			bodyReader = strings.NewReader(s.payload)
		}

		req, err := http.NewRequest(s.method, ts.URL+"/v1/resource", bodyReader)
		require.NoError(t, err)
		if s.headers != nil {
			req.Header = s.headers
		}

		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		if s.payload != "" {
			require.Equal(t, s.payload, string(respBody))
		} else {
			require.Equal(t, "get response ok", string(respBody))
		}

		// Find generated .sh file and .payload file
		entries, err := os.ReadDir(scriptDir)
		require.NoError(t, err)

		var shFileName string
		var payloadFileName string
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".sh") {
				shFileName = e.Name()
			}
			if strings.HasSuffix(e.Name(), ".payload") {
				payloadFileName = e.Name()
			}
		}

		require.NotEmpty(t, shFileName, "expected .sh script file to be created")
		require.True(t, strings.HasPrefix(shFileName, "dump-"), "script file should start with dump-")

		shPath := filepath.Join(scriptDir, shFileName)
		shContent, err := os.ReadFile(shPath)
		require.NoError(t, err)

		info, err := os.Stat(shPath)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0755), info.Mode().Perm(), "script should have 0755 executable permissions")

		for _, sub := range s.wantInScript {
			require.Contains(t, string(shContent), sub)
		}

		if s.wantPayloadFile {
			require.NotEmpty(t, payloadFileName, "expected .payload file to be created")
			payloadPath := filepath.Join(scriptDir, payloadFileName)
			payloadContent, err := os.ReadFile(payloadPath)
			require.NoError(t, err)
			require.Equal(t, s.payload, string(payloadContent))
		} else {
			require.Empty(t, payloadFileName, "did not expect .payload file for request without body")
		}

		// Empirically execute the generated script and verify output
		cmd := exec.Command("bash", shPath)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "generated shell script execution failed: %s", string(out))
		if s.payload != "" {
			require.Contains(t, string(out), s.payload)
		} else {
			require.Contains(t, string(out), "get response ok")
		}
	}

	t.Run("GET request generates executable script without payload file", func(t *testing.T) {
		check(t, testScenario{
			method: "GET",
			headers: http.Header{
				"X-Custom-Header": []string{"my-token"},
			},
			wantInScript: []string{
				"#!/bin/bash",
				"set -e",
				"curl \\",
				"-s \\",
				"--http1.1 \\",
				"-i \\",
				"-H \"X-Custom-Header: my-token\" \\",
				ts.URL + "/v1/resource",
			},
			wantPayloadFile: false,
		})
	})

	t.Run("POST request with body generates script referencing .payload file", func(t *testing.T) {
		check(t, testScenario{
			method:  "POST",
			payload: `{"action":"create_test_item","count":42}`,
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			wantInScript: []string{
				"#!/bin/bash",
				"set -e",
				"DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"",
				"curl \\",
				"-s \\",
				"--http1.1 \\",
				"-i \\",
				"-X POST \\",
				"-H \"Content-Type: application/json\" \\",
				"--data-binary @\"$DIR/",
				ts.URL + "/v1/resource",
			},
			wantPayloadFile: true,
		})
	})
}
