package execzv2_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/execzv2"
	"github.com/stretchr/testify/require"
)

func TestUnitBasicRun(t *testing.T) {
	type testScenario struct {
		bin         string
		args        []string
		expectError bool
		errMsg      string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := context.Background()
		proc := execzv2.New(ctx, execzv2.Options{
			Bin:  s.bin,
			Args: s.args,
		})
		err := proc.Run()
		if s.expectError {
			require.Error(t, err)
			if s.errMsg != "" {
				require.Contains(t, err.Error(), s.errMsg)
			}
			return
		}
		require.NoError(t, err)
	}

	t.Run("Successful bash command", func(t *testing.T) {
		check(t, testScenario{
			bin:  "bash",
			args: []string{"-c", "exit 0"},
		})
	})

	t.Run("Failing bash command", func(t *testing.T) {
		check(t, testScenario{
			bin:         "bash",
			args:        []string{"-c", "exit 2"},
			expectError: true,
		})
	})

	t.Run("Non-existent command", func(t *testing.T) {
		check(t, testScenario{
			bin:         "non_existent_binary_gox_execzv2",
			expectError: true,
		})
	})
}

func TestUnitStdOutAndStdErr(t *testing.T) {
	type testScenario struct {
		command string
		wantOut []string
		wantErr []string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := context.Background()
		proc := execzv2.New(ctx, execzv2.Options{
			Bin:  "bash",
			Args: []string{"-c", s.command},
		})
		require.NoError(t, proc.Run())

		var gotOut []string
		for {
			line, ok := proc.StdOut()
			if !ok {
				break
			}
			gotOut = append(gotOut, line)
		}
		require.Equal(t, s.wantOut, gotOut)

		var gotErr []string
		for {
			line, ok := proc.StdErr()
			if !ok {
				break
			}
			gotErr = append(gotErr, line)
		}
		require.Equal(t, s.wantErr, gotErr)
	}

	t.Run("Single line stdout", func(t *testing.T) {
		check(t, testScenario{
			command: "echo hello",
			wantOut: []string{"hello"},
			wantErr: nil,
		})
	})

	t.Run("Multi-line stdout and stderr", func(t *testing.T) {
		check(t, testScenario{
			command: "echo out1; echo err1 >&2; echo out2; echo err2 >&2",
			wantOut: []string{"out1", "out2"},
			wantErr: []string{"err1", "err2"},
		})
	})

	t.Run("No newline stdout", func(t *testing.T) {
		check(t, testScenario{
			command: "printf 'no-nl'",
			wantOut: []string{"no-nl"},
			wantErr: nil,
		})
	})

	t.Run("Empty lines preserved", func(t *testing.T) {
		check(t, testScenario{
			command: "echo ''; echo 'non-empty'; echo ''",
			wantOut: []string{"", "non-empty", ""},
			wantErr: nil,
		})
	})
}

func TestUnitStdInAndClose(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "tr",
		Args: []string{"a-z", "A-Z"},
	})
	require.NoError(t, proc.Start())

	proc.StdIn("hello\n")
	proc.StdIn("world\n")
	require.NoError(t, proc.Close())

	line1, ok1 := proc.StdOut()
	require.True(t, ok1)
	require.Equal(t, "HELLO", line1)

	line2, ok2 := proc.StdOut()
	require.True(t, ok2)
	require.Equal(t, "WORLD", line2)

	_, ok3 := proc.StdOut()
	require.False(t, ok3)

	require.NoError(t, proc.WaitFor())
}

func TestUnitEnvsAndDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "echo $MY_VAR && pwd"},
		Dir:  realTmpDir,
		Envs: map[string]string{
			"MY_VAR": "custom_value",
		},
	})
	require.NoError(t, proc.Run())

	line1, ok1 := proc.StdOut()
	require.True(t, ok1)
	require.Equal(t, "custom_value", line1)

	line2, ok2 := proc.StdOut()
	require.True(t, ok2)
	require.Equal(t, realTmpDir, line2)
}

func TestUnitKillDeadline(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "trap '' TERM; sleep 10"},
	})
	require.NoError(t, proc.Start())

	time.Sleep(50 * time.Millisecond)
	err := proc.Kill(ctx, 100*time.Millisecond)
	require.Error(t, err)
	require.ErrorIs(t, err, execzv2.ErrKilled)
	require.ErrorIs(t, proc.WaitFor(), execzv2.ErrKilled)
}

func TestUnitKillGraceful(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "sleep",
		Args: []string{"10"},
	})
	require.NoError(t, proc.Start())

	time.Sleep(50 * time.Millisecond)
	err := proc.Kill(ctx, 2*time.Second)
	require.Error(t, err)
	require.NotErrorIs(t, err, execzv2.ErrKilled)
}

func TestUnitKillContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	proc := execzv2.New(context.Background(), execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "trap '' TERM; sleep 10"},
	})
	require.NoError(t, proc.Start())

	time.Sleep(50 * time.Millisecond)
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := proc.Kill(ctx, 5*time.Second)
	require.Error(t, err)
	require.ErrorIs(t, err, execzv2.ErrKilled)
}

func TestUnitCustomOptions(t *testing.T) {
	var capturedIn []string
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "echo",
		Args: []string{"test"},
		StdOut: func() (string, bool) {
			return "custom_out", true
		},
		StdErr: func() (string, bool) {
			return "custom_err", true
		},
		StdIn: func(str string) {
			capturedIn = append(capturedIn, str)
		},
	})
	proc.StdIn("hello")
	require.Equal(t, []string{"hello"}, capturedIn)

	out, ok := proc.StdOut()
	require.True(t, ok)
	require.Equal(t, "custom_out", out)

	errStr, ok := proc.StdErr()
	require.True(t, ok)
	require.Equal(t, "custom_err", errStr)
}

func TestUnitProcessAlreadyStarted(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	require.NoError(t, proc.Start())
	err := proc.Start()
	require.Error(t, err)
	require.Contains(t, err.Error(), "process already started")

	err2 := proc.Run()
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "process already started")

	require.NoError(t, proc.WaitFor())
}

func TestUnitConcurrentWait(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "sleep 0.1 && exit 0"},
	})
	require.NoError(t, proc.Start())

	const numWaiters = 5
	errs := make(chan error, numWaiters)
	for i := 0; i < numWaiters; i++ {
		go func() {
			errs <- proc.WaitFor()
		}()
	}

	for i := 0; i < numWaiters; i++ {
		err := <-errs
		require.NoError(t, err)
	}
}

func TestUnitKillNotStarted(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "echo",
		Args: []string{"test"},
	})
	err := proc.Kill(ctx, 1*time.Second)
	require.NoError(t, err)
}

func TestUnitCloseIdempotent(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	require.NoError(t, proc.Close())
	require.NoError(t, proc.Close())
}

func TestUnitNilContext(t *testing.T) {
	proc := execzv2.New(nil, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "echo nil_ctx"},
	})
	require.NoError(t, proc.Run())

	line, ok := proc.StdOut()
	require.True(t, ok)
	require.Equal(t, "nil_ctx", line)
}

func TestUnitLargeOutput(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "seq 1 10000"},
	})
	require.NoError(t, proc.Run())

	count := 0
	for {
		_, ok := proc.StdOut()
		if !ok {
			break
		}
		count++
	}
	require.Equal(t, 10000, count)
}

func TestUnitExitCode(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 42"},
	})
	err := proc.Run()
	require.Error(t, err)
	require.ErrorIs(t, proc.WaitFor(), proc.Err)
}

func TestUnitUtf8StdIn(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin: "cat",
	})
	require.NoError(t, proc.Start())

	proc.StdIn("Hello 世界 🚀\n")
	require.NoError(t, proc.Close())

	line, ok := proc.StdOut()
	require.True(t, ok)
	require.Equal(t, "Hello 世界 🚀", line)

	_, ok = proc.StdOut()
	require.False(t, ok)

	require.NoError(t, proc.WaitFor())
}
