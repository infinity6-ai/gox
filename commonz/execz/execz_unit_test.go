package execz_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/execz"
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
		proc := execz.New(ctx, execz.Options{
			Bin:  s.bin,
			Args: s.args,
		})
		defer proc.Close()

		err := proc.Run(ctx, 0)
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
			bin:         "non_existent_binary_gox_execz",
			expectError: true,
		})
	})
}

func TestUnitGetStdoutAndGetStderr(t *testing.T) {
	type testScenario struct {
		command string
		wantOut []execz.Line
		wantErr []execz.Line
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := context.Background()
		proc := execz.New(ctx, execz.Options{
			Bin:           "bash",
			Args:          []string{"-c", s.command},
			CaptureStdout: 64,
			CaptureStderr: 64,
		})
		defer proc.Close()
		require.NoError(t, proc.Start(ctx))

		var gotOut []execz.Line
		var gotErr []execz.Line
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

		require.NoError(t, proc.WaitFor(ctx, 0))
		wg.Wait()

		require.Equal(t, s.wantOut, gotOut)
		require.Equal(t, s.wantErr, gotErr)
	}

	t.Run("Single line stdout", func(t *testing.T) {
		check(t, testScenario{
			command: "echo hello",
			wantOut: []execz.Line{{Line: "hello", Partial: false}},
			wantErr: nil,
		})
	})

	t.Run("Multi-line stdout and stderr", func(t *testing.T) {
		check(t, testScenario{
			command: "echo out1; echo err1 >&2; echo out2; echo err2 >&2",
			wantOut: []execz.Line{
				{Line: "out1", Partial: false},
				{Line: "out2", Partial: false},
			},
			wantErr: []execz.Line{
				{Line: "err1", Partial: false},
				{Line: "err2", Partial: false},
			},
		})
	})

	t.Run("No newline stdout yields partial line", func(t *testing.T) {
		check(t, testScenario{
			command: "printf 'no-nl'",
			wantOut: []execz.Line{{Line: "no-nl", Partial: true}},
			wantErr: nil,
		})
	})

	t.Run("Empty lines preserved", func(t *testing.T) {
		check(t, testScenario{
			command: "echo ''; echo 'non-empty'; echo ''",
			wantOut: []execz.Line{
				{Line: "", Partial: false},
				{Line: "non-empty", Partial: false},
				{Line: "", Partial: false},
			},
			wantErr: nil,
		})
	})
}

func TestUnitLineBufferSize(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:            "bash",
		Args:           []string{"-c", "echo '1234567890'"},
		CaptureStdout:  64,
		LineBufferSize: 4,
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	var gotOut []execz.Line
	for line := range proc.GetStdout() {
		gotOut = append(gotOut, line)
	}
	require.NoError(t, proc.WaitFor(ctx, 0))

	expected := []execz.Line{
		{Line: "1234", Partial: true},
		{Line: "5678", Partial: true},
		{Line: "90", Partial: false},
	}
	require.Equal(t, expected, gotOut)
}

func TestUnitGetStdinAndCloseStdin(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:           "tr",
		Args:          []string{"a-z", "A-Z"},
		CaptureStdout: 64,
		EnableStdin:   64,
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	var gotOut []execz.Line
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range proc.GetStdout() {
			gotOut = append(gotOut, line)
		}
	}()

	proc.GetStdin() <- execz.Line{Line: "hello", Partial: false}
	proc.GetStdin() <- execz.Line{Line: "world", Partial: false}
	require.NoError(t, proc.CloseStdin())

	<-done
	require.NoError(t, proc.WaitFor(ctx, 0))
	require.Equal(t, []execz.Line{
		{Line: "HELLO", Partial: false},
		{Line: "WORLD", Partial: false},
	}, gotOut)
}

func TestUnitCloseViaChannel(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:           "tr",
		Args:          []string{"a-z", "A-Z"},
		CaptureStdout: 64,
		EnableStdin:   64,
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	var gotOut []execz.Line
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range proc.GetStdout() {
			gotOut = append(gotOut, line)
		}
	}()

	proc.GetStdin() <- execz.Line{Line: "channel-close", Partial: false}
	close(proc.GetStdin())

	<-done
	require.NoError(t, proc.WaitFor(ctx, 0))
	require.Equal(t, []execz.Line{
		{Line: "CHANNEL-CLOSE", Partial: false},
	}, gotOut)
}

func TestUnitStdinPartial(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:           "cat",
		CaptureStdout: 64,
		EnableStdin:   64,
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	var gotOut []execz.Line
	done := make(chan struct{})
	go func() {
		defer close(done)
		for line := range proc.GetStdout() {
			gotOut = append(gotOut, line)
		}
	}()

	proc.GetStdin() <- execz.Line{Line: "hel", Partial: true}
	proc.GetStdin() <- execz.Line{Line: "lo", Partial: true}
	close(proc.GetStdin())

	<-done
	require.NoError(t, proc.WaitFor(ctx, 0))
	require.Equal(t, []execz.Line{
		{Line: "hello", Partial: true},
	}, gotOut)
}

func TestUnitDiscardByDefault(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "seq 1 50000"},
	})
	defer proc.Close()
	require.Nil(t, proc.GetStdout())
	require.Nil(t, proc.GetStderr())
	require.Nil(t, proc.GetStdin())

	require.NoError(t, proc.Run(ctx, 0))
}

func TestUnitEnvsAndDir(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	realTmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	proc := execz.New(ctx, execz.Options{
		Bin:           "bash",
		Args:          []string{"-c", "echo $MY_VAR && pwd"},
		Dir:           realTmpDir,
		CaptureStdout: 64,
		Envs: map[string]string{
			"MY_VAR": "custom_value",
		},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	var gotOut []string
	for line := range proc.GetStdout() {
		gotOut = append(gotOut, line.Line)
	}
	require.NoError(t, proc.WaitFor(ctx, 0))

	require.Equal(t, []string{"custom_value", realTmpDir}, gotOut)
}

func TestUnitLargeOutputChannel(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:           "bash",
		Args:          []string{"-c", "seq 1 10000"},
		CaptureStdout: 256,
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	count := 0
	for line := range proc.GetStdout() {
		require.False(t, line.Partial)
		count++
	}
	require.NoError(t, proc.WaitFor(ctx, 0))
	require.Equal(t, 10000, count)
}

func TestUnitKillDeadline(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "trap '' TERM; sleep 10"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	time.Sleep(50 * time.Millisecond)
	err := proc.Kill(ctx, 100*time.Millisecond)
	require.Error(t, err)
	require.ErrorIs(t, err, execz.ErrKilled)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorIs(t, proc.WaitFor(ctx, 0), execz.ErrKilled)
}

func TestUnitKillGraceful(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "sleep",
		Args: []string{"10"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	time.Sleep(50 * time.Millisecond)
	err := proc.Kill(ctx, 2*time.Second)
	require.Error(t, err)
	require.NotErrorIs(t, err, execz.ErrKilled)
}

func TestUnitKillContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	proc := execz.New(context.Background(), execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "trap '' TERM; sleep 10"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(context.Background()))

	time.Sleep(50 * time.Millisecond)
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := proc.Kill(ctx, 5*time.Second)
	require.Error(t, err)
	require.ErrorIs(t, err, execz.ErrKilled)
	require.ErrorIs(t, err, context.Canceled)
}

func TestUnitProcessAlreadyStarted(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))
	err := proc.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "process already started")

	err2 := proc.Run(ctx, 0)
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "process already started")

	require.NoError(t, proc.WaitFor(ctx, 0))
}

func TestUnitConcurrentWait(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "sleep 0.1 && exit 0"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	const numWaiters = 5
	errs := make(chan error, numWaiters)
	for i := 0; i < numWaiters; i++ {
		go func() {
			errs <- proc.WaitFor(ctx, 0)
		}()
	}

	for i := 0; i < numWaiters; i++ {
		err := <-errs
		require.NoError(t, err)
	}
}

func TestUnitKillNotStarted(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "echo",
		Args: []string{"test"},
	})
	defer proc.Close()
	err := proc.Kill(ctx, 1*time.Second)
	require.NoError(t, err)
}

func TestUnitCloseIdempotent(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:         "bash",
		Args:        []string{"-c", "exit 0"},
		EnableStdin: 10,
	})
	require.NoError(t, proc.CloseStdin())
	require.NoError(t, proc.CloseStdin())
	require.NoError(t, proc.Close())
	require.NoError(t, proc.Close())
}

func TestUnitExitCode(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 42"},
	})
	defer proc.Close()
	err := proc.Run(ctx, 0)
	require.Error(t, err)
	require.ErrorIs(t, proc.WaitFor(ctx, 0), proc.Err())
}

func TestUnitUtf8Stdin(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execzv2_or_execz_Options())
	require.NoError(t, proc.Start(ctx))

	proc.GetStdin() <- execz.Line{Line: "Hello 世界 🚀", Partial: false}
	require.NoError(t, proc.CloseStdin())

	var got []execz.Line
	for line := range proc.GetStdout() {
		got = append(got, line)
	}
	require.NoError(t, proc.WaitFor(ctx, 0))
	require.Equal(t, []execz.Line{
		{Line: "Hello 世界 🚀", Partial: false},
	}, got)
}

func execzv2_or_execz_Options() execz.Options {
	return execz.Options{
		Bin:           "cat",
		CaptureStdout: 10,
		EnableStdin:   10,
	}
}

func TestUnitWaitForContextCancelled(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "sleep",
		Args: []string{"5"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	waitCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := proc.WaitFor(waitCtx, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cancelled while waiting for process")

	_ = proc.Kill(ctx, 100*time.Millisecond)
}

func TestUnitStartContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	proc := execz.New(context.Background(), execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	defer proc.Close()

	err := proc.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cancelled before starting process")
}

func TestUnitRunDeadlineExceeded(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "sleep",
		Args: []string{"5"},
	})
	defer proc.Close()
	err := proc.Run(ctx, 50*time.Millisecond)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUnitWaitForDeadlineExceeded(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "sleep",
		Args: []string{"5"},
	})
	defer proc.Close()
	require.NoError(t, proc.Start(ctx))

	err := proc.WaitFor(ctx, 50*time.Millisecond)
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	_ = proc.Kill(ctx, 100*time.Millisecond)
}

func TestUnitIsRunning(t *testing.T) {
	ctx := context.Background()
	proc := execz.New(ctx, execz.Options{
		Bin:  "sleep",
		Args: []string{"0.2"},
	})
	defer proc.Close()
	require.False(t, proc.IsRunning())

	require.NoError(t, proc.Start(ctx))
	require.True(t, proc.IsRunning())

	require.NoError(t, proc.WaitFor(ctx, 0))
	require.False(t, proc.IsRunning())

	failProc := execz.New(ctx, execz.Options{
		Bin: "non_existent_binary_xyz",
	})
	defer failProc.Close()
	require.False(t, failProc.IsRunning())
	require.Error(t, failProc.Start(ctx))
	require.False(t, failProc.IsRunning())
}

func TestUnitDeferCloseAfterNew(t *testing.T) {
	ctx := context.Background()

	// 1. Defer close on unstarted process is a no-op and marks it finished
	func() {
		proc := execz.New(ctx, execz.Options{
			Bin:  "sleep",
			Args: []string{"10"},
		})
		defer proc.Close()
	}()

	// 2. Defer close on running process kills it
	var runningProc *execz.Process
	func() {
		runningProc = execz.New(ctx, execz.Options{
			Bin:  "sleep",
			Args: []string{"10"},
		})
		defer runningProc.Close()
		require.NoError(t, runningProc.Start(ctx))
		require.True(t, runningProc.IsRunning())
	}()
	require.False(t, runningProc.IsRunning())

	// 3. Defer close on cleanly finished process is a no-op
	func() {
		proc := execz.New(ctx, execz.Options{
			Bin:  "bash",
			Args: []string{"-c", "exit 0"},
		})
		defer proc.Close()
		require.NoError(t, proc.Run(ctx, 0))
		require.False(t, proc.IsRunning())
	}()
}

func TestUnitCannotRestart(t *testing.T) {
	ctx := context.Background()

	// Closed before start cannot be started
	proc1 := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	require.NoError(t, proc1.Close())
	err := proc1.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already finished or closed")

	// Finished process cannot be started again
	proc2 := execz.New(ctx, execz.Options{
		Bin:  "bash",
		Args: []string{"-c", "exit 0"},
	})
	defer proc2.Close()
	require.NoError(t, proc2.Run(ctx, 0))
	err2 := proc2.Start(ctx)
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "already finished or closed")

	err3 := proc2.Run(ctx, 0)
	require.Error(t, err3)
	require.Contains(t, err3.Error(), "already finished or closed")
}
