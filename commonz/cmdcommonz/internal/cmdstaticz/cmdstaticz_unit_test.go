package cmdstaticz_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestUnitCmdStaticzGenerateFile(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name       string
		secondRun  func(srcDir, outFile string)
		expectDiff bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "stzfiles")
		require.NoError(t, os.MkdirAll(srcDir, 0755))
		testFile := filepath.Join(srcDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("initial"), 0644))

		outFile := filepath.Join(tmpDir, "out.gen.go")

		runCmd := func() {
			parent := &cobra.Command{Use: "parent"}
			cmdstaticz.Prepare(ctx, parent)
			parent.SetArgs([]string{"staticz", "generate", "--code", "stzfiles", "--dir", srcDir, "--out", outFile})
			require.NoError(t, parent.ExecuteContext(ctx))
		}

		// First generation
		runCmd()
		_, err := os.Stat(outFile)
		require.NoError(t, err)

		// Set modification time in the past to detect if the file is rewritten
		pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, os.Chtimes(outFile, pastTime, pastTime))

		s.secondRun(srcDir, outFile)

		// Second generation
		runCmd()
		stat2, err := os.Stat(outFile)
		require.NoError(t, err)

		if s.expectDiff {
			require.False(t, stat2.ModTime().Equal(pastTime), "out file should be rewritten when content changes")
		} else {
			require.True(t, stat2.ModTime().Equal(pastTime), "out file must not be rewritten when content is unchanged")
		}
	}

	t.Run("Subsequent run with unchanged input does not touch output file", func(t *testing.T) {
		check(t, testScenario{
			name: "Subsequent run with unchanged input does not touch output file",
			secondRun: func(srcDir, outFile string) {
				// No changes to files
			},
			expectDiff: false,
		})
	})

	t.Run("Subsequent run with modified input updates output file", func(t *testing.T) {
		check(t, testScenario{
			name: "Subsequent run with modified input updates output file",
			secondRun: func(srcDir, outFile string) {
				testFile := filepath.Join(srcDir, "test.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("updated content"), 0644))
			},
			expectDiff: true,
		})
	})
}

