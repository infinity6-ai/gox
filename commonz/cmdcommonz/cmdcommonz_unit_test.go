package cmdcommonz_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/commonz/cmdcommonz"
	"github.com/infinity6-ai/gox/versionz/version"
	"github.com/stretchr/testify/require"
)

func TestUnitCmdCommonzPrepare(t *testing.T) {
	ctx := context.Background()

	type testScenario struct {
		name         string
		expectedName string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cmd := cmdcommonz.Prepare(ctx)
		require.NotNil(t, cmd)
		require.Equal(t, s.expectedName, cmd.Use)
		require.Equal(t, version.Version(), cmd.Version)
		require.NotEmpty(t, cmd.Commands())

		hasVersion := false
		hasStaticz := false
		for _, c := range cmd.Commands() {
			if c.Name() == "version" {
				hasVersion = true
			}
			if c.Name() == "staticz" {
				hasStaticz = true
			}
		}
		require.True(t, hasVersion)
		require.True(t, hasStaticz)
	}

	t.Run("Prepare returns root command with subcommands", func(t *testing.T) {
		check(t, testScenario{
			name:         "Prepare returns root command with subcommands",
			expectedName: "commonz",
		})
	})
}
