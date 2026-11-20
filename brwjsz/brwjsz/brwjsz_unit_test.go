package brwjsz_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/stretchr/testify/require"
)

func TestUnitChromeOptions(t *testing.T) {
	type testScenario struct {
		name        string
		headless    bool
		userDataDir string
		wantTemp    bool
		wantFlag    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		opts := brwjsz.Options{}
		opts.Chrome("", s.headless, s.userDataDir)

		require.NotEmpty(t, opts.Command.Bin)

		foundHeadless := false
		foundUserData := false
		foundThrottle := false
		for _, arg := range opts.Command.Args {
			if strings.HasPrefix(arg, "--headless") {
				foundHeadless = true
			}
			if strings.HasPrefix(arg, "--user-data-dir") {
				foundUserData = true
			}
			if strings.HasPrefix(arg, "--disable-background-timer-throttling") {
				foundThrottle = true
			}
		}

		require.True(t, foundUserData, "user-data-dir flag should be present")
		require.True(t, foundThrottle, "anti-throttling flag should be present")
		require.Equal(t, s.headless, foundHeadless)
	}

	t.Run("Headless with auto temp dir", func(t *testing.T) {
		check(t, testScenario{
			name:        "headless auto temp",
			headless:    true,
			userDataDir: "",
			wantTemp:    true,
		})
	})

	t.Run("Headed with custom persistent dir", func(t *testing.T) {
		check(t, testScenario{
			name:        "headed custom dir",
			headless:    false,
			userDataDir: "/tmp/custom-profile",
			wantTemp:    false,
		})
	})
}

func TestUnitBrowserE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	opts := brwjsz.Options{
		TabCount: 2,
	}
	opts.Chrome("", false, "") // # let is non headless
	opts.Command.CaptureStderr = 100
	opts.Command.CaptureStdout = 100

	b := brwjsz.New(ctx, opts)
	defer b.Close()

	err := b.Start(ctx)
	require.NoError(t, err)

	tab0 := b.Tab(0)
	tab1 := b.Tab(1)

	// Tab 0 operations
	err = brwclient.Proc(ctx, tab0, "console.log", "hello tab 0")
	require.NoError(t, err)

	// Arrow function evaluation
	err = brwclient.Proc(ctx, tab0, "((x) => console.log('Arrow:', x))", "tested successfully")
	require.NoError(t, err)

	// Eval arithmetic
	val, err := brwclient.Eval[int](ctx, tab0, "21 * 2")
	require.NoError(t, err)
	require.Equal(t, 42, val)

	// Reference creation and usage
	type UserInfo struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	userRef := brwclient.MustCreate[UserInfo]()
	closer, err := brwclient.Set(ctx, tab0, userRef, UserInfo{Name: "Alice", Score: 100})
	require.NoError(t, err)
	defer closer.Close()

	// Use ref in function
	err = brwclient.Proc(ctx, tab0, `(u) => { window.__aliceName = u.Name; }`, userRef)
	require.NoError(t, err)

	aliceName, err := brwclient.Eval[string](ctx, tab0, "window.__aliceName")
	require.NoError(t, err)
	require.Equal(t, "Alice", aliceName)

	// Tab 1 operations (isolated from Tab 0)
	valTab1, err := brwclient.Eval[int](ctx, tab1, "100 + 200")
	require.NoError(t, err)
	require.Equal(t, 300, valTab1)

	// Tab 1 should not have Tab 0's window.__aliceName
	hasAlice, err := brwclient.Eval[bool](ctx, tab1, "typeof window.__aliceName !== 'undefined'")
	require.NoError(t, err)
	require.False(t, hasAlice)
}
