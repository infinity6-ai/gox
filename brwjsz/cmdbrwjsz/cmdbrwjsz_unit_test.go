package cmdbrwjsz_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/cmdbrwjsz"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdBrwjszPrepare(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name         string
		expectedName string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cmd := cmdbrwjsz.Prepare(ctx)
		require.NotNil(t, cmd)
		require.Equal(t, s.expectedName, cmd.Use)
		require.NotEmpty(t, cmd.Commands())
	}

	t.Run("Prepare returns root command", func(t *testing.T) {
		check(t, testScenario{
			name:         "Prepare returns root command",
			expectedName: "brwjsz",
		})
	})

	t.Run("Root command contains version subcommand", func(t *testing.T) {
		cmd := cmdbrwjsz.Prepare(ctx)
		verCmd, _, err := cmd.Find([]string{"version"})
		require.NoError(t, err)
		require.NotNil(t, verCmd)
		require.Equal(t, "version", verCmd.Use)
	})
}
