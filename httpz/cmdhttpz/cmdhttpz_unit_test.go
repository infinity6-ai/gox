package cmdhttpz_test

import (
	"context"
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
}
