package execzv2_test

import (
	"context"
	"path/filepath"
	"sync"
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

func TestUnitGetStdoutAndGetStderr(t *testing.T) {
	type testScenario struct {
		command string
		wantOut []string
		wantErr []string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := context.Background()
		proc := execzv2.New(ctx, execzv2.Options{
			Bin:           "bash",
			Args:          []string{"-c", s.command},
			CaptureStdout: 64,
			CaptureStderr: 64,
		})
		require.NoError(t, proc.Start())

		var gotOut []string
		var gotErr []string
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			for line := range proc.GetStdout() {
				gotOut = append(gotOut, line)
			}
		}()
		go func() {
			defer wg.Done()
			for line := range proc.GetStderr() {
				gotErr = append(gotErr, line)
			}
		}()

		require.NoError(t, proc.WaitFor())
		wg.Wait()

		require.Equal(t, s.wantOut, gotOut)
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

func TestUnitGetStdinAndClose(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:           "tr",
		Args:          []string{"a-z", "A-Z"},
		CaptureStdout: 64,
		EnableStdin:   64,
	})
	require.NoError(t, proc.Start())

	var gotOut []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range proc.GetStdout() {
			gotOut = append(gotOut, line)
		}
	}()

	proc.GetStdin() <- "hello\n"
	proc.GetStdin() <- "world\n"
	require.NoError(t, proc.Close())

	<-done
	require.NoError(t, proc.WaitFor())
	require.Equal(t, []string{"HELLO", "WORLD"}, gotOut)
}

func TestUnitCloseViaChannel(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:           "tr",
		Args:          []string{"a-z", "A-Z"},
		CaptureStdout: 64,
		EnableStdin:   64,
	})
	require.NoError(t, proc.Start())

	var gotOut []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range proc.GetStdout() {
			gotOut = append(gotOut, line)
		}
	}()

	proc.GetStdin() <- "channel-close\n"
	close(proc.GetStdin())

	<-done
	require.NoError(t, proc.WaitFor())
	require.Equal(t, []string{"CHANNEL-CLOSE"}, gotOut)
}

func TestUnitDiscardByDefault(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:  "bash",
		Args: []string{"-c", "seq 1 50000"},
	})
	require.Nil(t, proc.GetStdout())
	require.Nil(t, proc.GetStderr())
	require.Nil(t, proc.GetStdin())

	// Runs without hanging because output goes to io.Discard
	require.NoError(t, proc.Run())
}

func TestUnitEnvsAndDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	proc := execzv2.New(ctx, execzv2.Options{
		Bin:           "bash",
		Args:          []string{"-c", "echo $MY_VAR && pwd"},
		Dir:           realTmpDir,
		CaptureStdout: 64,
		Envs: map[string]string{
			"MY_VAR": "custom_value",
		},
	})
	require.NoError(t, proc.Start())

	var gotOut []string
	for line := range proc.GetStdout() {
		gotOut = append(gotOut, line)
	}
	require.NoError(t, proc.WaitFor())

	require.Equal(t, []string{"custom_value", realTmpDir}, gotOut)
}

func TestUnitLargeOutputChannel(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:           "bash",
		Args:          []string{"-c", "seq 1 10000"},
		CaptureStdout: 256,
	})
	require.NoError(t, proc.Start())

	count := 0
	for range proc.GetStdout() {
		count++
	}
	require.NoError(t, proc.WaitFor())
	require.Equal(t, 10000, count)
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
		Bin:         "bash",
		Args:        []string{"-c", "exit 0"},
		EnableStdin: 10,
	})
	require.NoError(t, proc.Close())
	require.NoError(t, proc.Close())
}

func TestUnitNilContext(t *testing.T) {
	proc := execzv2.New(nil, execzv2.Options{
		Bin:           "bash",
		Args:          []string{"-c", "echo nil_ctx"},
		CaptureStdout: 10,
	})
	require.NoError(t, proc.Start())

	var got []string
	for line := range proc.GetStdout() {
		got = append(got, line)
	}
	require.NoError(t, proc.WaitFor())
	require.Equal(t, []string{"nil_ctx"}, got)
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

func TestUnitUtf8Stdin(t *testing.T) {
	ctx := context.Background()
	proc := execzv2.New(ctx, execzv2.Options{
		Bin:           "cat",
		CaptureStdout: 10,
		EnableStdin:   10,
	})
	require.NoError(t, proc.Start())

	proc.GetStdin() <- "Hello 世界 🚀\n"
	require.NoError(t, proc.Close())

	var got []string
	for line := range proc.GetStdout() {
		got = append(got, line)
	}
	require.NoError(t, proc.WaitFor())
	require.Equal(t, []string{"Hello 世界 🚀"}, got)
}

func TestUnitMethodAliases(t *testing.T) {
	proc := execzv2.New(context.Background(), execzv2.Options{
		Bin:           "bash",
		CaptureStdout: 10,
		CaptureStderr: 10,
		EnableStdin:   10,
	})
	require.Equal(t, (<-chan string)(proc.GetStdout()), proc.GetStdOut())
	require.Equal(t, (<-chan string)(proc.GetStderr()), proc.GetStdErr())
	require.Equal(t, (chan<- string)(proc.GetStdin()), proc.GetStdIn())
}
