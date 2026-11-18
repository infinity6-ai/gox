package cmdhttpz_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/infinity6-ai/gox/httpz/cmdhttpz"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdHttpzPrepare(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name         string
		expectedName string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cmd := cmdhttpz.Prepare(ctx)
		require.NotNil(t, cmd)
		require.Equal(t, s.expectedName, cmd.Use)
		require.NotEmpty(t, cmd.Commands())
	}

	t.Run("Prepare returns root command", func(t *testing.T) {
		check(t, testScenario{
			name:         "Prepare returns root command",
			expectedName: "httpz",
		})
	})

	t.Run("Root command contains site and sample subcommands", func(t *testing.T) {
		cmd := cmdhttpz.Prepare(ctx)
		siteCmd, _, err := cmd.Find([]string{"site"})
		require.NoError(t, err)
		require.NotNil(t, siteCmd)
		require.Equal(t, "site", siteCmd.Use)

		sampleCmd, _, err := siteCmd.Find([]string{"sample"})
		require.NoError(t, err)
		require.NotNil(t, sampleCmd)
		require.Equal(t, "sample", sampleCmd.Use)

		serveCmd, _, err := sampleCmd.Find([]string{"serve"})
		require.NoError(t, err)
		require.NotNil(t, serveCmd)
		require.Equal(t, "serve", serveCmd.Use)

		extractCmd, _, err := sampleCmd.Find([]string{"extract"})
		require.NoError(t, err)
		require.NotNil(t, extractCmd)
		require.Equal(t, "extract", extractCmd.Use)
	})

	t.Run("Execute site sample extract command via root command", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := cmdhttpz.Prepare(ctx)
		cmd.SetArgs([]string{"site", "sample", "extract", "--out", tempDir})
		err := cmd.ExecuteContext(ctx)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tempDir, "index.html"))
		require.NoError(t, err)
		require.Contains(t, string(content), "Sample Site")
	})
}
