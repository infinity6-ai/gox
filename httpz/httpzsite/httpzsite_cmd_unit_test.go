package httpzsite_test

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/httpz/httpzsite"
	"github.com/stretchr/testify/require"
)

func TestUnitSiteCmd(t *testing.T) {
	ctx := t.Context()

	t.Run("Cmd default use name and custom name", func(t *testing.T) {
		site := &httpzsite.Site{Pack: testPackName}
		cmdDefault := site.Cmd()
		require.Equal(t, "site", cmdDefault.Use)

		cmdCustom := site.Cmd("sample")
		require.Equal(t, "sample", cmdCustom.Use)

		serveCmd, _, err := cmdCustom.Find([]string{"serve"})
		require.NoError(t, err)
		require.NotNil(t, serveCmd)

		extractCmd, _, err := cmdCustom.Find([]string{"extract"})
		require.NoError(t, err)
		require.NotNil(t, extractCmd)
	})

	t.Run("Cmd extract subcommand extracts files", func(t *testing.T) {
		site := &httpzsite.Site{Pack: testPackName}
		tempDir := t.TempDir()

		cmd := site.Cmd("sample")
		cmd.SetArgs([]string{"extract", "--out", tempDir})
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tempDir, "index.html"))
		require.NoError(t, err)
		require.Equal(t, "<html><body>Home</body></html>", string(content))
	})

	t.Run("Cmd extract fails when out is empty", func(t *testing.T) {
		site := &httpzsite.Site{Pack: testPackName}
		cmd := site.Cmd()
		cmd.SetArgs([]string{"extract"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
	})

	t.Run("Cmd extract fails when extract fails", func(t *testing.T) {
		site := &httpzsite.Site{Pack: "nonexistent-pack"}
		cmd := site.Cmd()
		tempDir := t.TempDir()
		cmd.SetArgs([]string{"extract", "--out", tempDir})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract")
	})

	t.Run("Cmd serve with invalid user format returns error", func(t *testing.T) {
		site := &httpzsite.Site{Pack: testPackName}
		cmd := site.Cmd()
		cmd.SetArgs([]string{"serve", "--user", "invaliduserwithoutcolon"})
		err := cmd.ExecuteContext(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid user format")
	})

	t.Run("Cmd serve runs server with address and user flags and shuts down on context cancel", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close())

		site := &httpzsite.Site{Pack: testPackName}
		cmd := site.Cmd("sample")
		cmd.SetArgs([]string{"serve", "--address", addr, "--user", "guest:notasecret"})

		serveCtx, cancel := context.WithCancel(t.Context())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- cmd.ExecuteContext(serveCtx)
		}()

		// Wait until server is listening
		require.Eventually(t, func() bool {
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return true
			}
			return false
		}, 2*time.Second, 10*time.Millisecond)

		// Unauthenticated request receives 401
		resp, err := http.Get("http://" + addr + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Authenticated request with guest:notasecret receives 200
		req, err := http.NewRequest(http.MethodGet, "http://"+addr+"/", nil)
		require.NoError(t, err)
		req.SetBasicAuth("guest", "notasecret")
		respAuth, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer respAuth.Body.Close()
		require.Equal(t, http.StatusOK, respAuth.StatusCode)

		// Cancel context to trigger server shutdown
		cancel()

		select {
		case execErr := <-errChan:
			require.NoError(t, execErr)
		case <-time.After(2 * time.Second):
			t.Fatal("serve command did not exit after context cancellation")
		}
	})
}
