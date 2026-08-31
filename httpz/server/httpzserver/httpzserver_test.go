package httpzserver_test

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

// suffixResponseWriter is a wrapper around http.ResponseWriter that allows appending a suffix to the response body.
type suffixResponseWriter struct {
	http.ResponseWriter
	suffix      []byte
	headers     http.Header
	wroteHeader bool
	t           *testing.T
}

// newSuffixResponseWriter creates a new suffixResponseWriter.
func newSuffixResponseWriter(w http.ResponseWriter, suffix []byte, t *testing.T) *suffixResponseWriter {
	return &suffixResponseWriter{
		ResponseWriter: w,
		suffix:         suffix,
		headers:        make(http.Header),
		t:              t,
	}
}

// Header returns the header map that will be sent by WriteHeader.
func (srw *suffixResponseWriter) Header() http.Header {
	return srw.headers
}

// Write writes the data to the connection as part of an HTTP reply.
func (srw *suffixResponseWriter) Write(data []byte) (int, error) {
	if !srw.wroteHeader {
		srw.WriteHeader(http.StatusOK)
	}
	return srw.ResponseWriter.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (srw *suffixResponseWriter) WriteHeader(statusCode int) {
	if srw.wroteHeader {
		return
	}

	// Copy headers from the original response writer in case they were set before wrapping.
	for k, v := range srw.ResponseWriter.Header() {
		if _, ok := srw.headers[k]; !ok {
			srw.headers[k] = v
		}
	}

	// If handler set a content length, we need to adjust it.
	if cl := srw.headers.Get("Content-Length"); cl != "" {
		if clInt, err := strconv.Atoi(cl); err == nil {
			clInt += len(srw.suffix)
			srw.headers.Set("Content-Length", strconv.Itoa(clInt))
		}
	}

	// Copy our headers to the underlying response writer.
	for k, v := range srw.headers {
		srw.ResponseWriter.Header()[k] = v
	}

	srw.ResponseWriter.WriteHeader(statusCode)
	srw.wroteHeader = true
}

// flush writes the suffix to the response body.
func (srw *suffixResponseWriter) flush() {
	if !srw.wroteHeader {
		srw.WriteHeader(http.StatusOK)
	}
	_, err := srw.ResponseWriter.Write(srw.suffix)
	require.NoError(srw.t, err)
}

func TestUnitListen(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer s.Close()
	s.Listen()
}

func TestUnitServer_EndToEnd(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})

	// Add a filter
	s.AddFilter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Filter", "true")
			next.ServeHTTP(w, r)
		})
	})

	// Add a handler with a path parameter
	s.AddHandlerPatternFunc("GET /test/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "ID is %s", id)
	})

	// Add a prefix handler
	s.AddHandlerPrefixFunc("/prefix/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "prefix matched")
	}))

	s.Listen()
	serverAddr := s.Addr().String()

	serverErr := make(chan error, 1)
	go func() {
		s.Start()
		serverErr <- nil
	}()

	// Allow some time for the server to start
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{}

	// Test pattern handler
	t.Run("PatternHandler", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/test/123", serverAddr))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "true", resp.Header.Get("X-Test-Filter"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "ID is 123", string(body))
	})

	// Test prefix handler
	t.Run("PrefixHandler", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/prefix/anything", serverAddr))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "true", resp.Header.Get("X-Test-Filter"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "prefix matched", string(body))
	})

	// Test not found
	t.Run("NotFound", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("http://%s/not/found", serverAddr))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		// The filter should still run for 404s served by the mux
		require.Equal(t, "true", resp.Header.Get("X-Test-Filter"))
	})

	// Shutdown server
	require.NoError(t, s.Close())

	select {
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("server did not shut down gracefully")
	}
}

func TestUnitServer_FilterModifiesBody(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})

	// Add filter to modify request and response bodies
	s.AddFilter(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Modify request body without full read
			reqSuffix := "-modified-req"
			if r.ContentLength != -1 {
				r.ContentLength += int64(len(reqSuffix))
			}
			r.Body = io.NopCloser(io.MultiReader(r.Body, strings.NewReader(reqSuffix)))

			// Wrap the response writer to append a suffix to the response body
			respSuffix := "-modified-resp"
			srw := newSuffixResponseWriter(w, []byte(respSuffix), t)
			defer srw.flush()

			next.ServeHTTP(srw, r)
		})
	})

	// Add handler that echoes the request body
	s.AddHandlerPatternFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "echo:%s", string(body))
	})

	s.Listen()
	serverAddr := s.Addr().String()

	serverErr := make(chan error, 1)
	go func() {
		s.Start()
		serverErr <- nil
	}()

	// Allow some time for the server to start
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{}

	// Make request
	reqBody := "hello"
	resp, err := client.Post(fmt.Sprintf("http://%s/echo", serverAddr), "text/plain", strings.NewReader(reqBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	expectedRespBody := "echo:hello-modified-req-modified-resp"
	require.Equal(t, expectedRespBody, string(respBody))

	// Shutdown server
	require.NoError(t, s.Close())

	select {
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("server did not shut down gracefully")
	}
}
