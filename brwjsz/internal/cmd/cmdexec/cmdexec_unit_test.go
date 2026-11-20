package cmdexec_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdexec"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdExecPrepare(t *testing.T) {
	ctx := context.Background()
	parent := &cobra.Command{Use: "root"}
	cmd := cmdexec.Prepare(ctx, parent)

	require.NotNil(t, cmd)
	require.Equal(t, "exec [code]", cmd.Use)

	found, _, err := parent.Find([]string{"exec"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, cmd, found)

	// Verify flags
	require.NotNil(t, cmd.Flags().Lookup("code"))
	require.NotNil(t, cmd.Flags().Lookup("file"))
	require.NotNil(t, cmd.Flags().Lookup("site"))
	require.NotNil(t, cmd.Flags().Lookup("state"))
	require.NotNil(t, cmd.Flags().Lookup("arg"))
	require.NotNil(t, cmd.Flags().Lookup("arg-file"))
	require.NotNil(t, cmd.Flags().Lookup("raw"))
	require.NotNil(t, cmd.Flags().Lookup("disable-headless"))
	require.NotNil(t, cmd.Flags().Lookup("bin"))
	require.NotNil(t, cmd.Flags().Lookup("timeout"))
}
