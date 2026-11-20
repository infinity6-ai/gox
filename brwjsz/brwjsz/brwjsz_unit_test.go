package brwjsz_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/require"
)

func TestUnitChromeOptions(t *testing.T) {
	type testScenario struct {
		name        string
		headless    bool
		userDataDir string
		wantFlag    string
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
}

func TestUnitBrowserE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	opts := brwjsz.Options{}
	opts.Chrome("", true, "")
	opts.Command.CaptureStderr = 100

	b := brwjsz.New(ctx, opts)
	defer b.Close()

	err := b.Start(ctx)
	if err != nil {
		t.Logf("Error starting: %v", err)
		if b.Process() != nil && b.Process().GetStderr() != nil {
			for {
				select {
				case line := <-b.Process().GetStderr():
					t.Logf("Chrome stderr: %s", line.Line)
				default:
					goto doneStderr
				}
			}
		}
	doneStderr:
	}
	require.NoError(t, err)

	client := b.Client()

	// 1. Proc with standard function
	err = brwclient.Proc(ctx, client, "console.log", "hello from Go")
	require.NoError(t, err)

	// 2. Proc with anonymous arrow function
	err = brwclient.Proc(ctx, client, "((x) => console.log('Arrow received:', x))", "tested successfully")
	require.NoError(t, err)

	// 3. Eval arithmetic expression
	val, err := brwclient.Eval[int](ctx, client, "21 * 2")
	require.NoError(t, err)
	require.Equal(t, 42, val)

	// 4. Remote reference creation and lifecycle
	type UserInfo struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	}

	userRef := brwclient.MustCreate[UserInfo]("user")
	require.True(t, strings.HasPrefix(string(userRef), "user_"))
	closer, err := brwclient.Set(ctx, client, userRef, UserInfo{Name: "Alice", Score: 100})
	require.NoError(t, err)

	// Use ref inside JS function
	err = brwclient.Proc(ctx, client, `(u) => { window.__aliceName = u.name; }`, userRef)
	require.NoError(t, err)

	aliceName, err := brwclient.Eval[string](ctx, client, "window.__aliceName")
	require.NoError(t, err)
	require.Equal(t, "Alice", aliceName)

	// Func returning a calculated reference
	calcRef, err := brwclient.Func[int](ctx, client, `(u) => u.score * 3`, userRef)
	require.NoError(t, err)
	require.NotEmpty(t, calcRef)
	require.True(t, strings.HasPrefix(string(calcRef), "fn_"))

	calcVal, err := brwclient.Get[int](ctx, client, calcRef)
	require.NoError(t, err)
	require.Equal(t, 300, calcVal)

	// Non-transferable JS objects: DOM elements and Canvas context
	canvasRef := brwclient.MustFuncWithPrefix[any](ctx, client, "canvas", `() => document.createElement('canvas')`)
	require.True(t, strings.HasPrefix(string(canvasRef), "canvas_"))
	ctx2dRef := brwclient.MustFuncWithPrefix[any](ctx, client, "ctx2d", `(c) => c.getContext('2d')`, canvasRef)
	require.True(t, strings.HasPrefix(string(ctx2dRef), "ctx2d_"))

	// Perform RefDump (all active refs)
	dumpAll := brwclient.MustRefDump(ctx, client)
	require.Contains(t, dumpAll[string(canvasRef)], "HTMLCanvasElement")
	require.Contains(t, dumpAll[string(ctx2dRef)], "CanvasRenderingContext2D")

	// Perform selective RefDump (only canvasRef)
	dumpSelective := brwclient.MustRefDump(ctx, client, canvasRef)
	require.Contains(t, dumpSelective[string(canvasRef)], "HTMLCanvasElement")
	require.NotContains(t, dumpSelective, string(ctx2dRef))

	brwclient.MustDelete(ctx, client, canvasRef)
	brwclient.MustDelete(ctx, client, ctx2dRef)

	// Clean up references
	err = closer.Close()
	require.NoError(t, err)

	err = brwclient.Delete(ctx, client, calcRef)
	require.NoError(t, err)
}

func TestUnitBrowserSite(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	tempDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tempDir, "data.json"), []byte(`{"message":"hello from disk site"}`), 0o644)
	require.NoError(t, err)

	opts := brwjsz.Options{
		Site: &brwjsz.Site{
			Disk: pathz.MustParse(tempDir),
		},
	}
	opts.Chrome("", true, "")

	b := brwjsz.New(ctx, opts)
	defer b.Close()

	err = b.Start(ctx)
	require.NoError(t, err)

	client := b.Client()

	fetched, err := brwclient.Call[string](ctx, client, `async () => {
		const res = await fetch('/data.json');
		const json = await res.json();
		return json.message;
	}`)
	require.NoError(t, err)
	require.Equal(t, "hello from disk site", fetched)
}
