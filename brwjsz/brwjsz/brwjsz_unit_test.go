package brwjsz_test

import (
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/stretchr/testify/require"
)

func TestUnitChromeOptions(t *testing.T) {
	type testScenario struct {
		name        string
		headless    bool
		userDataDir string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		opts := brwjsz.Options{}
		opts.Chrome("", s.headless, s.userDataDir)

		require.NotEmpty(t, opts.Command.Bin)

		foundHeadless := false
		foundUserData := false
		foundLogging := false
		for _, arg := range opts.Command.Args {
			if strings.HasPrefix(arg, "--headless") {
				foundHeadless = true
			}
			if strings.HasPrefix(arg, "--user-data-dir") {
				foundUserData = true
			}
			if arg == "--enable-logging=stderr" {
				foundLogging = true
			}
		}

		require.True(t, foundUserData, "user-data-dir flag should be present")
		require.True(t, foundLogging, "--enable-logging=stderr flag should be present")
		require.Equal(t, s.headless, foundHeadless)
	}

	t.Run("Headless with auto temp dir", func(t *testing.T) {
		check(t, testScenario{
			name:        "headless auto temp",
			headless:    true,
			userDataDir: "",
		})
	})

	t.Run("Headed with custom persistent dir", func(t *testing.T) {
		check(t, testScenario{
			name:        "headed custom dir",
			headless:    false,
			userDataDir: "/tmp/custom-profile",
		})
	})

	t.Run("Headless disabled via configz", func(t *testing.T) {
		closer := brwjsz.BrwjszDisableHeadless.MustSet(t.Context(), "true")
		defer closer.Close()

		opts := brwjsz.Options{}
		opts.Chrome("", true, "")

		for _, arg := range opts.Command.Args {
			require.False(t, strings.HasPrefix(arg, "--headless"), "headless flag should not be present when disabled via configz")
		}
	})
}
