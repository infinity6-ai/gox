// Package cmdexec provides the CLI command for executing JavaScript in a browser session.
package cmdexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/brwjsz"
	"github.com/infinity6-ai/gox/brwjsz/internal/cmd/cmdserver"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/spf13/cobra"
)

type execOptions struct {
	code            string
	file            string
	site            string
	state           string
	args            []string
	argFiles        []string
	raw             bool
	disableHeadless bool
	bin             string
	timeout         time.Duration
}

func parseArgValue(s string) any {
	var val any
	if err := json.Unmarshal([]byte(s), &val); err == nil {
		return val
	}
	return s
}

func printResult(res any, raw bool) error {
	if res == nil {
		return nil
	}
	if str, ok := res.(string); ok && raw {
		fmt.Println(str)
		return nil
	}
	data, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("%v\n", res)
		return nil
	}
	fmt.Println(string(data))
	return nil
}

// Prepare configures and attaches the exec command to the parent cobra command.
func Prepare(ctx context.Context, parent *cobra.Command) *cobra.Command {
	opts := &execOptions{}

	cmd := &cobra.Command{
		Use:   "exec [code]",
		Short: "Execute JavaScript in a headless browser session or running server",
		RunE: func(cmd *cobra.Command, args []string) error {
			code := opts.code
			if code == "" && opts.file != "" {
				if opts.file == "-" {
					b, err := io.ReadAll(os.Stdin)
					if err != nil {
						return fmt.Errorf("failed to read from stdin: %w", err)
					}
					code = string(b)
				} else {
					b, err := os.ReadFile(opts.file)
					if err != nil {
						return fmt.Errorf("failed to read file %s: %w", opts.file, err)
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
			for _, a := range opts.args {
				resolvedArgs = append(resolvedArgs, parseArgValue(a))
			}
			for _, f := range opts.argFiles {
				b, err := os.ReadFile(f)
				if err != nil {
					return fmt.Errorf("failed to read arg file %s: %w", f, err)
				}
				resolvedArgs = append(resolvedArgs, parseArgValue(string(b)))
			}

			// Mode A: Connected to running server daemon via --state
			if opts.state != "" {
				state, err := cmdserver.ReadState(opts.state)
				if err != nil {
					return fmt.Errorf("failed to read state: %w", err)
				}

				reqPayload, err := json.Marshal(handlers.ExecRequest{
					Code: code,
					Args: resolvedArgs,
				})
				if err != nil {
					return fmt.Errorf("failed to marshal exec payload: %w", err)
				}

				execUrl := state.URL + handlers.ExecPath
				httpReq, err := http.NewRequestWithContext(cmd.Context(), "POST", execUrl, bytes.NewReader(reqPayload))
				if err != nil {
					return fmt.Errorf("failed to create http request: %w", err)
				}
				httpReq.Header.Set("Content-Type", "application/json")

				httpClient := &http.Client{Timeout: opts.timeout}
				httpResp, err := httpClient.Do(httpReq)
				if err != nil {
					return fmt.Errorf("failed to send exec request to server daemon (pid %d): %w", state.Pid, err)
				}
				defer httpResp.Body.Close()

				var execResp handlers.ExecResponse
				if err := json.NewDecoder(httpResp.Body).Decode(&execResp); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}

				if !execResp.Ok {
					return fmt.Errorf("execution error: %s", execResp.Error)
				}

				var resVal any
				if len(execResp.Result) > 0 {
					if err := json.Unmarshal(execResp.Result, &resVal); err != nil {
						return fmt.Errorf("failed to unmarshal exec result: %w", err)
					}
				}

				return printResult(resVal, opts.raw)
			}

			// Mode B: Standalone one-shot browser session
			cmdCtx := cmd.Context()
			if opts.timeout > 0 {
				var cancel context.CancelFunc
				cmdCtx, cancel = context.WithTimeout(cmdCtx, opts.timeout)
				defer cancel()
			}

			brwOpts := brwjsz.Options{}
			brwOpts.Chrome(opts.bin, !opts.disableHeadless, "")
			if opts.site != "" {
				brwOpts.Site = &brwjsz.Site{
					Disk: pathz.MustParse(opts.site),
				}
			}

			b := brwjsz.New(cmdCtx, brwOpts)
			defer b.Close()

			if err := b.Start(cmdCtx); err != nil {
				return fmt.Errorf("failed to start browser: %w", err)
			}

			res, err := brwclient.Exec(cmdCtx, b.Client(), code, resolvedArgs...)
			if err != nil {
				return fmt.Errorf("execution error: %w", err)
			}

			return printResult(res, opts.raw)
		},
	}

	cmd.Flags().StringVar(&opts.code, "code", "", "JavaScript code to execute")
	cmd.Flags().StringVar(&opts.file, "file", "", "JavaScript file path to execute (or - for stdin)")
	cmd.Flags().StringVar(&opts.site, "site", "", "Disk directory path to serve as a static site (accessible via fetch)")
	cmd.Flags().StringVar(&opts.state, "state", "", "Path to state file of running server daemon")
	cmd.Flags().StringArrayVar(&opts.args, "arg", nil, "Argument to pass to the function (can be specified multiple times)")
	cmd.Flags().StringArrayVar(&opts.argFiles, "arg-file", nil, "File containing argument value (can be specified multiple times)")
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "Print string output without JSON quotes")
	cmd.Flags().BoolVar(&opts.disableHeadless, "disable-headless", false, "Disable headless browser mode")
	cmd.Flags().StringVar(&opts.bin, "bin", "", "Browser binary path (defaults to auto-detect)")
	cmd.Flags().DurationVar(&opts.timeout, "timeout", 30*time.Second, "Execution timeout")

	if parent != nil {
		parent.AddCommand(cmd)
	}

	return cmd
}
