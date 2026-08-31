package httpzserver_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/httpz/server/httpzserver"
	"github.com/stretchr/testify/require"
)

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
