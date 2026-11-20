package cmdserver_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdserver"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdServerPrepare(t *testing.T) {
	ctx := context.Background()
	parent := &cobra.Command{Use: "root"}
	cmd := cmdserver.Prepare(ctx, parent)

	require.NotNil(t, cmd)
	require.Equal(t, "server", cmd.Use)

	startCmd, _, err := cmd.Find([]string{"start"})
	require.NoError(t, err)
	require.NotNil(t, startCmd)
	require.NotNil(t, startCmd.Flags().Lookup("state"))
	require.NotNil(t, startCmd.Flags().Lookup("detach"))

	execCmd, _, err := cmd.Find([]string{"exec"})
	require.NoError(t, err)
	require.NotNil(t, execCmd)
	require.NotNil(t, execCmd.Flags().Lookup("state"))
	require.NotNil(t, execCmd.Flags().Lookup("code"))

	stopCmd, _, err := cmd.Find([]string{"stop"})
	require.NoError(t, err)
	require.NotNil(t, stopCmd)
	require.NotNil(t, stopCmd.Flags().Lookup("state"))
}
