// Package cmdserver provides CLI commands to manage persistent brwjsz server daemons.
package cmdserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func readState(path string) (*ServerState, error) {
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

func writeState(path string, state ServerState) error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func parseArgValue(s string) any {
	var val any
	if err := json.Unmarshal([]byte(s), &val); err == nil {
		return val
	}
	return s
}

// Prepare registers the server command group (start, exec, stop) onto parent.
func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Manage persistent brwjsz server daemon sessions",
	}

	serverCmd.AddCommand(newStartCmd())
	serverCmd.AddCommand(newExecCmd())
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
					state, err := readState(statePath)
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
			if err := writeState(statePath, state); err != nil {
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

func newExecCmd() *cobra.Command {
	var statePath string
	var code string
	var file string
	var argsList []string
	var argFiles []string
	var raw bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "exec [code]",
		Short: "Execute JavaScript in the running brwjsz server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if statePath == "" {
				return fmt.Errorf("--state is required")
			}

			state, err := readState(statePath)
			if err != nil {
				return err
			}

			if code == "" && file != "" {
				if file == "-" {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
					code = string(b)
				} else {
					b, err := os.ReadFile(file)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", file, err)
					}
					code = string(b)
				}
			} else if code == "" && len(args) > 0 {
				code = args[0]
			}

			if code == "" {
				return fmt.Errorf("no code specified (use --code, --file, or code argument)")
			}

			var resolvedArgs []any
			for _, a := range argsList {
				resolvedArgs = append(resolvedArgs, parseArgValue(a))
			}
			for _, f := range argFiles {
				b, err := os.ReadFile(f)
				if err != nil {
					return fmt.Errorf("failed to read arg file %s: %w", f, err)
				}
				resolvedArgs = append(resolvedArgs, parseArgValue(string(b)))
			}

			reqPayload, err := json.Marshal(handlers.ExecRequest{
				Code: code,
				Args: resolvedArgs,
			})
			if err != nil {
				return err
			}

			execUrl := state.URL + handlers.ExecPath
			httpReq, err := http.NewRequestWithContext(cmd.Context(), "POST", execUrl, bytes.NewReader(reqPayload))
			if err != nil {
				return err
			}
			httpReq.Header.Set("Content-Type", "application/json")

			httpClient := &http.Client{Timeout: timeout}
			httpResp, err := httpClient.Do(httpReq)
			if err != nil {
				return fmt.Errorf("failed to send exec request to server: %w", err)
			}
			defer httpResp.Body.Close()

			var execResp handlers.ExecResponse
			if err := json.NewDecoder(httpResp.Body).Decode(&execResp); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			if !execResp.Ok {
				return fmt.Errorf("execution error: %s", execResp.Error)
			}

			if execResp.Result == nil {
				return nil
			}

			if str, ok := execResp.Result.(string); ok && raw {
				fmt.Println(str)
				return nil
			}

			data, err := json.Marshal(execResp.Result)
			if err != nil {
				fmt.Printf("%v\n", execResp.Result)
				return nil
			}
			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&statePath, "state", "", "Path to state file (required)")
	cmd.Flags().StringVar(&code, "code", "", "JavaScript code to execute")
	cmd.Flags().StringVar(&file, "file", "", "JavaScript file path to execute (or - for stdin)")
	cmd.Flags().StringArrayVar(&argsList, "arg", nil, "Argument to pass to the function (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&argFiles, "arg-file", nil, "File containing argument value (can be specified multiple times)")
	cmd.Flags().BoolVar(&raw, "raw", false, "Print string output without JSON quotes")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "Execution timeout")

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

			state, err := readState(statePath)
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
