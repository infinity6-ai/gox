package cmdqunit_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdqunit"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdQunitPrepare(t *testing.T) {
	ctx := context.Background()
	parent := &cobra.Command{Use: "root"}
	cmd := cmdqunit.Prepare(ctx, parent)

	require.NotNil(t, cmd)
	require.Equal(t, "qunit", cmd.Use)

	found, _, err := parent.Find([]string{"qunit"})
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, cmd, found)

	// Verify flags
	require.NotNil(t, cmd.Flags().Lookup("site"))
	require.NotNil(t, cmd.Flags().Lookup("suffix"))
	require.NotNil(t, cmd.Flags().Lookup("state"))
	require.NotNil(t, cmd.Flags().Lookup("bin"))
	require.NotNil(t, cmd.Flags().Lookup("disable-headless"))
	require.NotNil(t, cmd.Flags().Lookup("timeout"))
}
