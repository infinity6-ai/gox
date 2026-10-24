package cmdstaticz_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/infinity6-ai/gox/commonz/cmdcommonz/internal/cmdstaticz"
	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdStaticzPrepare(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		parent := &cobra.Command{Use: "parent"}
		cmd := cmdstaticz.Prepare(ctx, parent)
		require.NotNil(t, cmd)
		require.Equal(t, "staticz", cmd.Use)
		require.Equal(t, version.Version(), cmd.Version)

		generateCmd, _, err := cmd.Find([]string{"generate"})
		require.NoError(t, err)
		require.NotNil(t, generateCmd)
		require.Equal(t, "generate", generateCmd.Name())

		require.NotNil(t, generateCmd.Flags().Lookup("imp"))
		require.NotNil(t, generateCmd.Flags().Lookup("code"))
		require.NotNil(t, generateCmd.Flags().Lookup("dir"))
		require.NotNil(t, generateCmd.Flags().Lookup("out"))
		require.Equal(t, "internal/stzfiles/stzfiles.gen.go", generateCmd.Flags().Lookup("out").DefValue)
	}

	t.Run("Prepare mounts staticz and configures generate command flags", func(t *testing.T) {
		check(t, testScenario{
			name: "Prepare mounts staticz and configures generate command flags",
		})
	})
}

func TestUnitCmdStaticzGenerateStdout(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		tmpDir := t.TempDir()
		parent := &cobra.Command{Use: "parent"}
		cmdstaticz.Prepare(ctx, parent)

		var outBuf bytes.Buffer
		parent.SetOut(&outBuf)
		parent.SetArgs([]string{"staticz", "generate", "--code", "stzfiles", "--dir", tmpDir, "--out", "-"})

		err := parent.ExecuteContext(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, outBuf.String())
	}

	t.Run("Generate writes to stdout when out is dash", func(t *testing.T) {
		check(t, testScenario{
			name: "Generate writes to stdout when out is dash",
		})
	})
}
