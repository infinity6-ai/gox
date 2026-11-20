// Package cmdserver provides CLI commands to manage persistent brwjsz server daemons.
package cmdserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/spf13/cobra"
)

// ServerState encapsulates runtime metadata written to the state file.
type ServerState struct {
	Pid int    `json:"pid"`
	URL string `json:"url"`
}

// ReadState reads and parses the ServerState from the given file path.
func ReadState(path string) (*ServerState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %s: %w", path, err)
	}
	var state ServerState
	if err := json.Unmarshal(b, &state); err != nil {
		return nil, fmt.Errorf("invalid state file %s: %w", path, err)
	}
	return &state, nil
}

// WriteState writes the ServerState as formatted JSON to the given file path.
func WriteState(path string, state ServerState) error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// Prepare registers the server command group (start, stop) onto parent.
func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Manage persistent brwjsz server daemon sessions",
	}

	serverCmd.AddCommand(newStartCmd())
	serverCmd.AddCommand(newStopCmd())

	if parent != nil {
		parent.AddCommand(serverCmd)
	}

	return serverCmd
}

func newStartCmd() *cobra.Command {
	var statePath string
	var sitePath string
	var binPath string
	var headless bool
	var detach bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a persistent brwjsz server daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statePath == "" {
				return fmt.Errorf("--state is required")
			}

			if detach {
				var newArgs []string
				for _, a := range os.Args[1:] {
					if a != "--detach" && a != "-d" && !strings.HasPrefix(a, "--detach=") && !strings.HasPrefix(a, "-d=") {
						newArgs = append(newArgs, a)
					}
				}
				child := exec.Command(os.Args[0], newArgs...)
				child.Stdout = nil
				child.Stderr = os.Stderr
				child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
				if err := child.Start(); err != nil {
					return fmt.Errorf("failed to start detached server: %w", err)
				}
				for i := 0; i < 50; i++ {
					time.Sleep(100 * time.Millisecond)
					state, err := ReadState(statePath)
					if err == nil && state.URL != "" {
						fmt.Printf("Server started [pid: %d, url: %s]\n", state.Pid, state.URL)
						return nil
					}
				}
				return fmt.Errorf("timed out waiting for detached server to initialize")
			}

			ctx := cmd.Context()
			opts := brwjsz.Options{}
			opts.Chrome(binPath, headless, "")
			if sitePath != "" {
				opts.Site = &brwjsz.Site{
					Disk: pathz.MustParse(sitePath),
				}
			}

			b := brwjsz.New(ctx, opts)
			defer b.Close()

			if err := b.Start(ctx); err != nil {
				return fmt.Errorf("failed to start browser: %w", err)
			}

			serverUrl := b.Server().Base().String()
			state := ServerState{
				Pid: os.Getpid(),
				URL: serverUrl,
			}
			if err := WriteState(statePath, state); err != nil {
				return fmt.Errorf("failed to write state file: %w", err)
			}
			defer os.Remove(statePath)

			fmt.Printf("Server started [pid: %d, url: %s]\n", state.Pid, state.URL)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			select {
			case <-sigChan:
			case <-b.WaitStop():
			case <-ctx.Done():
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&statePath, "state", "", "Path to state file (required)")
	cmd.Flags().StringVar(&sitePath, "site", "", "Disk directory path to serve as a static site")
	cmd.Flags().StringVar(&binPath, "bin", "", "Browser binary path (defaults to auto-detect)")
	cmd.Flags().BoolVar(&headless, "headless", true, "Run browser in headless mode")
	cmd.Flags().BoolVarP(&detach, "detach", "d", false, "Run server in the background")

	return cmd
}

func newStopCmd() *cobra.Command {
	var statePath string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running brwjsz server daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statePath == "" {
				return fmt.Errorf("--state is required")
			}

			state, err := ReadState(statePath)
			if err != nil {
				return err
			}

			stopUrl := state.URL + handlers.StopPath
			resp, err := http.Post(stopUrl, "application/json", nil)
			if err != nil {
				_ = os.Remove(statePath)
				return nil
			}
			defer resp.Body.Close()

			for i := 0; i < 50; i++ {
				time.Sleep(50 * time.Millisecond)
				if _, err := os.Stat(statePath); os.IsNotExist(err) {
					break
				}
			}
			_ = os.Remove(statePath)
			fmt.Println("Server stopped")
			return nil
		},
	}

	cmd.Flags().StringVar(&statePath, "state", "", "Path to state file (required)")

	return cmd
}
